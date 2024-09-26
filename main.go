package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"tcp-client/msg"
)

const (
	CONN_HOST = "127.0.0.1"
	CONN_PORT = "9018"
	CONN_TYPE = "tcp"
)

func main() {
	packetSendPtr := flag.Uint("packet", 0, "Packet number to send")
	flag.Parse()

	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	defer func() {
		_ = conn.Close()
	}()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ReadNWrite(int64(*packetSendPtr), conn)
	if err != nil {
		log.Printf("Alarm, %s", err.Error())
	}

}

func ReadNWrite(packetNuber int64, conn net.Conn) error {

	msgs := msg.GetMsgs()

	fmt.Printf("Packet %d, %X\n", packetNuber, msgs[packetNuber])

	_, write_err := conn.Write(msgs[packetNuber])
	if write_err != nil {
		return fmt.Errorf("failed: %w", write_err)
	}

	buf, read_err := io.ReadAll(conn)
	if read_err != nil {
		return fmt.Errorf("failed: %w", read_err)
	}

	if len(buf) < 3 || buf[0] != 0x02 {
		return fmt.Errorf("packet type or length error")
	}

	if getCRCfromBytes(msgs[packetNuber]) != getCRCfromBytes(buf) {
		return fmt.Errorf("CRC error")
	}
	log.Println("CRC - ok")

	conn.(*net.TCPConn).CloseWrite()
	return nil
}

func getCRCfromBytes(packet []byte) uint16 {
	var crc uint16
	crc = uint16(packet[len(packet)-2])
	crc = crc << 8
	crc = crc + uint16(packet[len(packet)-1])
	return crc
}
