package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"tcp-client/msg"
	"time"
)

const (
	CONN_HOST = "127.0.0.1"
	CONN_PORT = "9018"
	CONN_TYPE = "tcp"
)

func main() {
	conn, err := net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	defer conn.Close()
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
		go func() {
			fmt.Printf("Packet %d, %X\n", i, msg)

			_, write_err := conn.Write(msg)
			if write_err != nil {
				log.Printf("failed: %s", write_err.Error())
			}
			log.Println("Sent")
			buf, read_err := io.ReadAll(conn)
			if read_err != nil {
				log.Printf("failed: %s", read_err.Error())
			}
			log.Println("Read out")

			if len(buf) < 3 {
				log.Printf("packet length error")
			}
			if buf[0] != 0x02 {
				log.Printf("packet type  error")
			}

			if getCRCfromBytes(msg) != getCRCfromBytes(buf) {
				log.Printf("CRC error")
			}
			log.Println("CRC - ok")
			// time.Sleep(1 * time.Second)}
		}()
	}
	time.Sleep(210 * time.Second)

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
