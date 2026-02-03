package main

import (
	"fmt"
	"log"
	"net"

	"github.com/gosnmp/gosnmp"
)

const (
	ListenAddr     = "0.0.0.0:1611"
	InterfaceCount = 8 // テストしたいIF数
)

func main() {
	conn, err := net.ListenPacket("udp", ListenAddr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close()

	fmt.Printf("SNMP Bulk Responder running on %s...\n", ListenAddr)

	buf := make([]byte, 4096)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			continue
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

		// レスポンスの構築
		response := &gosnmp.SnmpPacket{
			Version:   packet.Version,
			Community: packet.Community,
			PDUType:   gosnmp.GetResponse,
			RequestID: packet.RequestID,
			Error:     gosnmp.NoError,
			Variables: generateIfMetrics(),
		}

		out, err := response.MarshalMsg()
		if err != nil {
			log.Printf("Marshal error: %v", err)
			continue
		}
		conn.WriteTo(out, addr)
	}
}

// 6つのメトリクス (In/Out x Octets/Errors/Discards) を生成
func generateIfMetrics() []gosnmp.SnmpPDU {
	var pduList []gosnmp.SnmpPDU

	// OIDの末尾番号 (IF-MIB定義)
	// 10:InOct, 13:InDisc, 14:InErr, 16:OutOct, 19:OutDisc, 20:OutErr
	metricOids := []int{10, 13, 14, 16, 19, 20}

	for i := 1; i <= InterfaceCount; i++ {
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
