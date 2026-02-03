package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gosnmp/gosnmp"
	"golang.org/x/net/ipv4"
)

const (
	ListenAddr     = "0.0.0.0:1611"
	InterfaceCount = 8 // テストしたいIF数
)

// snmpbulkget -v2c -c public -Cn0 -Cr18 127.0.0.1:1611 .1.3.6.1.2.1.2.2.1.10

func main() {
	port := flag.Int("port", 1611, "UDP port to listen on")
	ifCount := flag.Int("if", 8, "Number of interfaces to simulate")
	delayMs := flag.Int("delay", 0, "Artificial delay in milliseconds before responding (negative value to simulate failure/silent)")
	flag.Parse()

	ListenAddr := fmt.Sprintf("0.0.0.0:%d", *port)
	conn, err := net.ListenPacket("udp", ListenAddr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close()

	pc := ipv4.NewPacketConn(conn)
	if err := pc.SetControlMessage(ipv4.FlagDst, true); err != nil {
		log.Fatalf("Error setting control message: %v", err)
	}

	fmt.Printf("SNMP Bulk Responder running on %s (Logging Destination IP)...\n", ListenAddr)
	fmt.Printf("Settings: Interfaces=%d, Delay=%dms\n", *ifCount, *delayMs)
	if *delayMs < 0 {
		fmt.Println("Simulating failure mode: No responses will be sent.")
	}

	buf := make([]byte, 4096)
	for {
		n, cm, srcAddr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			continue
		}

		dstIP := "Unknown"
		if cm != nil {
			dstIP = cm.Dst.String()
		}

		// 受信データをデコード
		packet, err := gosnmp.Default.SnmpDecodePacket(buf[:n])
		if err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}

		// GetBulkRequest以外は無視
		if packet.PDUType != gosnmp.GetBulkRequest {
			continue
		}

		log.Printf("[Bulk Request] From: %s -> To: %s | Community: %s", srcAddr, dstIP, packet.Community)
		if *delayMs < 0 {
			log.Printf("   >> Failure mode: dropping request from %s", srcAddr)
			continue // 負値の場合は何も返さずループの先頭に戻る
		}

		if *delayMs > 0 {
			time.Sleep(time.Duration(*delayMs) * time.Millisecond)
		}

		// レスポンスの構築
		response := &gosnmp.SnmpPacket{
			Version:   packet.Version,
			Community: packet.Community,
			PDUType:   gosnmp.GetResponse,
			RequestID: packet.RequestID,
			Error:     gosnmp.NoError,
			Variables: generateIfMetrics(*ifCount),
		}

		out, err := response.MarshalMsg()
		if err != nil {
			log.Printf("Marshal error: %v", err)
			continue
		}
		conn.WriteTo(out, srcAddr)
	}
}

// 6つのメトリクス (In/Out x Octets/Errors/Discards) を生成
func generateIfMetrics(count int) []gosnmp.SnmpPDU {
	var pduList []gosnmp.SnmpPDU

	// OIDの末尾番号 (IF-MIB定義)
	// 10:InOct, 13:InDisc, 14:InErr, 16:OutOct, 19:OutDisc, 20:OutErr
	metricOids := []int{10, 13, 14, 16, 19, 20}

	for i := 1; i <= count; i++ {
		for _, suffix := range metricOids {
			oid := fmt.Sprintf(".1.3.6.1.2.1.2.2.1.%d.%d", suffix, i)
			pduList = append(pduList, gosnmp.SnmpPDU{
				Name:  oid,
				Type:  gosnmp.Counter32,
				Value: uint32(1000 * i * suffix), // テスト用の適当な数値
			})
		}
	}
	return pduList
}
