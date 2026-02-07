package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/gosnmp/gosnmp"
	"golang.org/x/net/ipv4"
)

type AgentConfig struct {
	IfCount int
	DelayMs int
}

func main() {
	port := flag.Int("port", 1611, "UDP port to listen on")
	csvPath := flag.String("csv", "snmp_config.csv", "Path to the configuration CSV file")
	ifaceName := flag.String("iface", "snmp-dummy", "Interface name to bind to")
	flag.Parse()

	// IPをキーにしたマップで設定を保持
	configs, err := loadConfigMap(*csvPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 1. UDPコネクション作成
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("0.0.0.0:%d", *port))
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatalf("Listen Error: %v", err)
	}
	defer conn.Close()

	// 2. ソケットオプションの設定 (FileDescriptor経由)
	file, _ := conn.File()
	fd := int(file.Fd())

	// IP_FREEBIND: 自分のインターフェースにないIP宛のバインド/受信を許可
	syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_FREEBIND, 1)
	// SO_BINDTODEVICE: 特定のIFに紐付け
	syscall.SetsockoptString(fd, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, *ifaceName)
	file.Close()

	// 3. PacketConnに変換してDst IPを取得可能にする
	pc := ipv4.NewPacketConn(conn)
	pc.SetControlMessage(ipv4.FlagDst, true)

	fmt.Printf(">> SNMP Simulator: listening on %s (Freebind enabled)\n", *ifaceName)

	buf := make([]byte, 4096)
	for {
		n, cm, srcAddr, err := pc.ReadFrom(buf)
		if err != nil {
			continue
		}

		// 宛先IPが取れない場合は無視
		if cm == nil {
			continue
		}
		dstIP := cm.Dst.String()

		cfg, ok := configs[dstIP]
		if !ok {
			continue
		}

		// デコード以降の処理は前回と同じ
		packet, err := gosnmp.Default.SnmpDecodePacket(buf[:n])
		if err == nil {
			if packet.PDUType == gosnmp.GetBulkRequest || packet.PDUType == gosnmp.GetNextRequest || packet.PDUType == gosnmp.GetRequest {
				// 非同期でレスポンス処理（メインループを止めないため）
				go handleRequest(conn, srcAddr, dstIP, packet, cfg)
			}
		}
	}
}

func handleRequest(lc net.PacketConn, srcAddr net.Addr, dstIP string, packet *gosnmp.SnmpPacket, cfg AgentConfig) {
	// 1. 障害シミュレーション
	if cfg.DelayMs < 0 {
		log.Printf("[%s] DROP Request from %s", dstIP, srcAddr)
		return
	}

	// 2. 遅延シミュレーション
	if cfg.DelayMs > 0 {
		time.Sleep(time.Duration(cfg.DelayMs) * time.Millisecond)
	}

	// 全データ（MIBツリー）を生成
	allMetrics := generateIfMetrics(cfg.IfCount)

	var responseVariables []gosnmp.SnmpPDU

	switch packet.PDUType {
	case gosnmp.GetBulkRequest:
		// GetBulkの特殊パラメータ取得
		nonRepeaters := int(packet.NonRepeaters)
		maxRepetitions := int(packet.MaxRepetitions)

		for i, v := range packet.Variables {
			if i < nonRepeaters {
				// Non-repeaters: 次の1つだけを返す
				responseVariables = append(responseVariables, getNextOID(v.Name, allMetrics, 1)...)
			} else {
				// Repeaters: Max-repetitions 分だけ次々と返す
				responseVariables = append(responseVariables, getNextOID(v.Name, allMetrics, maxRepetitions)...)
			}
		}
	case gosnmp.GetNextRequest:
		// GetNext: 各変数の「次の1つ」を返す
		for _, v := range packet.Variables {
			responseVariables = append(responseVariables, getNextOID(v.Name, allMetrics, 1)...)
		}
	default:
		// GetRequest: そのものズバリを返す（今回は簡易的に全検索）
		for _, v := range packet.Variables {
			for _, m := range allMetrics {
				if m.Name == v.Name {
					responseVariables = append(responseVariables, m)
					break
				}
			}
		}
	}

	// レスポンス送信
	response := &gosnmp.SnmpPacket{
		Version:   packet.Version,
		Community: packet.Community,
		PDUType:   gosnmp.GetResponse,
		RequestID: packet.RequestID,
		Variables: responseVariables,
	}

	out, err := response.MarshalMsg()
	if err != nil {
		return
	}

	_, err = lc.WriteTo(out, srcAddr)
	if err == nil {
		log.Printf("[%s] SENT Response to %s (IFs: %d)", dstIP, srcAddr, cfg.IfCount)
	}
}

// 指定されたOIDより「後ろ」にあるOIDをcount個分返す関数
func getNextOID(requestedOID string, allMetrics []gosnmp.SnmpPDU, count int) []gosnmp.SnmpPDU {
	var results []gosnmp.SnmpPDU
	foundCount := 0

	for _, m := range allMetrics {
		// OIDの文字列比較（簡易版）。本来は数値配列での比較が正確ですが
		// 今回の固定OID構造なら文字列辞書順でも概ね動作します。
		if m.Name > requestedOID {
			results = append(results, m)
			foundCount++
			if foundCount >= count {
				break
			}
		}
	}

	// もし次が何もなければ EndOfMibView を入れるのがマナー
	if len(results) == 0 {
		results = append(results, gosnmp.SnmpPDU{
			Name: requestedOID,
			Type: gosnmp.EndOfMibView,
		})
	}
	return results
}

func loadConfigMap(path string) (map[string]AgentConfig, error) {
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

	configs := make(map[string]AgentConfig)
	for i, r := range records {
		if i == 0 || len(r) < 3 {
			continue
		}
		ifCnt, _ := strconv.Atoi(r[1])
		delay, _ := strconv.Atoi(r[2])
		configs[r[0]] = AgentConfig{IfCount: ifCnt, DelayMs: delay}
	}
	return configs, nil
}

func generateIfMetrics(count int) []gosnmp.SnmpPDU {
	var pduList []gosnmp.SnmpPDU
	metricOids := []int{10, 13, 14, 16, 19, 20} // ifInOctets, ifInDiscards, ifInErrors, ifOutOctets, ifOutDiscards, ifOutErrors
	now := uint32(time.Now().Unix() % 0xFFFFFFFF)

	for i := 1; i <= count; i++ {
		for _, suffix := range metricOids {
			pduList = append(pduList, gosnmp.SnmpPDU{
				Name:  fmt.Sprintf(".1.3.6.1.2.1.2.2.1.%d.%d", suffix, i),
				Type:  gosnmp.Counter32,
				Value: now,
			})
		}
	}
	metricOids = []int{6, 10} // ifHCInOctets, ifHCOutOctets

	for i := 1; i <= count; i++ {
		for _, suffix := range metricOids {
			pduList = append(pduList, gosnmp.SnmpPDU{
				Name:  fmt.Sprintf(".1.3.6.1.2.1.31.1.1.1.%d.%d", suffix, i),
				Type:  gosnmp.Counter32,
				Value: now,
			})
		}
	}
	return pduList
}
