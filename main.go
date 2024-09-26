package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"tcp-client/msg"
	"tcp-client/pkg"
)

const (
	CONN_HOST = "127.0.0.1"
	CONN_PORT = "9018"
	CONN_TYPE = "tcp"
)

type Packet struct {
	Header  []byte
	Length  []byte
	Payload []byte
	CRC     []byte
}

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
		log.Printf("Alarm ReadNWrite, %s", err.Error())
	}
	// time.Sleep(1 * time.Minute)
	for {
		err = listener(conn)
		if err != nil {
			log.Printf("Alarm listener, %s", err.Error())
		}
	}
}

func ReadNWrite(conn net.Conn) error {

	msgs := msg.GetMsgs()

	for i, msg := range msgs {
		var p Packet
		raw := buildPacket(i, p, msg)
		fmt.Printf("RAW %X\n", raw)

		_, write_err := conn.Write(raw)
		if write_err != nil {
			return fmt.Errorf("failed WRITE: %w", write_err)
		}

		buf := make([]byte, 3)
		_, read_err := conn.Read(buf)
		if read_err != nil {
			return fmt.Errorf("failed READ: %w", read_err)
		}

		if len(buf) < 3 {
			return fmt.Errorf("packet length error")
		}
		if buf[0] != 0x02 {
			return fmt.Errorf("packet type  error")
		}
		if getCRCfromBytes(raw) != getCRCfromBytes(buf) {
			return fmt.Errorf("CRC error")
		}
		log.Println("CRC - ok")
	}
	return nil
}

func getCRCfromBytes(packet []byte) uint16 {
	crc := binary.LittleEndian.Uint16(packet[len(packet)-2:])
	return crc
}

func buildPacket(i int, p Packet, msg []byte) []byte {
	if i == 0 {
		p.Header = append(p.Header, byte(0x01))
	} else {
		p.Header = append(p.Header, byte(0x08))
	}
	p.Payload = append(p.Payload, msg...)
	p.Length = binary.LittleEndian.AppendUint16(p.Length, uint16(len(p.Payload)))

	raw := append(p.Header, p.Length...)
	raw = append(raw, p.Payload...)
	p.CRC = binary.LittleEndian.AppendUint16(p.CRC, pkg.Crc16FFFF(raw))
	raw = append(raw, p.CRC...)
	return raw
}

func listener(conn net.Conn) error {
	var (
		localBuffer []byte
		dataLen     int
		err         error
	)
	readBuf := make([]byte, 1024)
	if dataLen, err = conn.Read(readBuf); err != nil {
		return fmt.Errorf("failed READ: %w", err)
	} else if dataLen > 0 {
		log.Printf("Incoming data siza = %d\n", dataLen)
		localBuffer = append(localBuffer, readBuf[:dataLen]...)
	}

	if len(localBuffer) < 3 {
		return fmt.Errorf("packet length error")
	}
	crc := pkg.Crc16FFFF(localBuffer[dataLen-2:])
	if crc != getCRCfromBytes(localBuffer) {
		return fmt.Errorf("CRC error")
	}
	log.Println("CRC - ok")
	return nil
}
