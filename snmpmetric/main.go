package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
	"golang.org/x/net/ipv4"
)

// ulimit -n 2048
// snmpbulkget -v2c -c public -Cn0 -Cr18 127.0.0.1:1611 .1.3.6.1.2.1.2.2.1.10

// アドレスごとの設定構造体
type AgentConfig struct {
	IP      string
	IfCount int
	DelayMs int
}

func main() {
	port := flag.Int("port", 1611, "UDP port to listen on")
	csvPath := flag.String("csv", "config.csv", "Path to the configuration CSV file")
	flag.Parse()

	configs, err := loadConfig(*csvPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var wg sync.WaitGroup
	for _, cfg := range configs {
		wg.Add(1)
		go func(cfg AgentConfig) {
			defer wg.Done()
			runAgent(cfg, *port)
		}(cfg)
	}

	fmt.Printf(">> Successfully initialized %d agents on port %d\n", len(configs), *port)
	wg.Wait()
}

func runAgent(cfg AgentConfig, port int) {
	listenAddr := fmt.Sprintf("%s:%d", cfg.IP, port)
	// パケットコネクションの作成
	lc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Printf("![%s] Bind Error: %v", cfg.IP, err)
		return
	}
	defer lc.Close()

	// 宛先IP取得用の設定
	pc := ipv4.NewPacketConn(lc)
	_ = pc.SetControlMessage(ipv4.FlagDst, true)
	log.Printf("[%s] Active (IFs: %d, Delay: %dms)", cfg.IP, cfg.IfCount, cfg.DelayMs)
	buf := make([]byte, 4096)
	for {
		n, cm, srcAddr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("[%s] Read error: %v", cfg.IP, err)
			continue
		}

		// パケットデコード
		packet, err := gosnmp.Default.SnmpDecodePacket(buf[:n])
		if err != nil {
			continue
		}

		// GetBulkRequestに対する処理
		if packet.PDUType == gosnmp.GetBulkRequest {
			// 1. 障害シミュレーション (delay < 0)
			if cfg.DelayMs < 0 {
				log.Printf("[%s] DROP Request from %s (Failure Mode)", cfg.IP, srcAddr)
				continue
			}

			// 2. 遅延シミュレーション
			if cfg.DelayMs > 0 {
				time.Sleep(time.Duration(cfg.DelayMs) * time.Millisecond)
			}

			// 3. レスポンス生成・送信
			dstIP := cfg.IP
			if cm != nil {
				dstIP = cm.Dst.String()
			}

			response := &gosnmp.SnmpPacket{
				Version:   packet.Version,
				Community: packet.Community,
				PDUType:   gosnmp.GetResponse,
				RequestID: packet.RequestID,
				Variables: generateIfMetrics(cfg.IfCount),
			}

			out, _ := response.MarshalMsg()
			_, err = lc.WriteTo(out, srcAddr)

			if err == nil {
				log.Printf("[%s] SENT Response to %s", dstIP, srcAddr)
			}
		}
	}
}

func loadConfig(path string) ([]AgentConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var configs []AgentConfig
	for i, r := range records {
		if i == 0 || len(r) < 3 { // ヘッダー飛ばし & 列数チェック
			continue
		}

		ifCnt, _ := strconv.Atoi(r[1])
		delay, _ := strconv.Atoi(r[2])

		configs = append(configs, AgentConfig{
			IP: r[0], IfCount: ifCnt, DelayMs: delay,
		})
	}
	return configs, nil
}

func generateIfMetrics(count int) []gosnmp.SnmpPDU {
	var pduList []gosnmp.SnmpPDU
	// 10:InOct, 13:InDisc, 14:InErr, 16:OutOct, 19:OutDisc, 20:OutErr
	metricOids := []int{10, 13, 14, 16, 19, 20}

	for i := 1; i <= count; i++ {
		for _, suffix := range metricOids {
			pduList = append(pduList, gosnmp.SnmpPDU{
				Name:  fmt.Sprintf(".1.3.6.1.2.1.2.2.1.%d.%d", suffix, i),
				Type:  gosnmp.Counter32,
				Value: uint32(time.Now().Unix() % 0xFFFFFFFF),
			})
		}
	}
	return pduList
}
