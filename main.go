package main

import (
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
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	defer func() {
		_ = conn.Close()
	}()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ReadNWrite(conn)
	if err != nil {
		log.Printf("Alarm, %s", err.Error())
	}

}

func ReadNWrite(conn net.Conn) error {

	msgs := msg.GetMsgs()

	for i, msg := range msgs {
		fmt.Printf("Packet %d, %X\n", i, msg)

		_, write_err := conn.Write(msg)
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

		if getCRCfromBytes(msg) != getCRCfromBytes(buf) {
			return fmt.Errorf("CRC error")
		}
		log.Println("CRC - ok")
	}
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
