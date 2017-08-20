package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
)

func getByte(value int32) []byte {

	b_buf := bytes.NewBuffer([]byte{})

	b_buf.WriteByte((byte)((value >> 24) & 0xFF))
	b_buf.WriteByte((byte)((value >> 16) & 0xFF))
	b_buf.WriteByte((byte)((value >> 8) & 0xFF))
	b_buf.WriteByte((byte)(value & 0xFF))

	return b_buf.Bytes()
}

func doEcho(c net.Conn) {
	defer c.Close()
	var outBodyLength int32
	var outRespCode int32
	for {
		inBuf := make([]byte, 4096)
		n, err := c.Read(inBuf[:])
		if n == 0 || err != nil {
			return
		}

		b_buf := bytes.NewBuffer([]byte{})

		bw := bufio.NewWriter(b_buf)
		bw.Write(inBuf[0:4])
		outBodyLength = 4
		bw.Write(getByte(outBodyLength))
		outRespCode = 200
		bw.Write(getByte(outRespCode))

		//////////////////////////////////////////////////
		bw.Flush()
		//fmt.Println("head_body len ", len(b_buf.Bytes()))
		////////////////////////////////////////////////////////////
		//返回byte

		c.Write(b_buf.Bytes())

	}
}

func doStart(host string) {
	if host == "" {
		fmt.Println("Press input host name")
		return
	}
	fmt.Println("Listening:" + host)
	ln, err := net.Listen("tcp", host)
	if err != nil {
		fmt.Printf("Listen Error:")
		return
	}
	defer ln.Close()

	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		doEcho(c)
	}
}

func main() {
	doStart(":8081")
}
