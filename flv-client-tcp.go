package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

// first : go run flv-server-example.go

// Compared to HTTP connections, TCP needs to simulate requests and manually parse data on its own
// need to build http request header
// need to parse http response header
// need to parse chunk data
// data format: [http response header ][\r\n\r\n][chunk size][\r\n][chunk data][\r\n][chunk size][\r\n][chunk data][\r\n]

var (
	// http://192.168.120.120:8080/edge/22100de0-0000-03b5-8e48-d6b125f8b000/mark.flv

	flvServerIP string = "localhost:13370"
	flvUrl      string = "/flv.flv"
)

func main() {

	fp, _ := os.OpenFile("./flv_output.flv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)

	// tcp socket
	var httpReqStr = fmt.Sprintf("GET %s HTTP/1.1\r\n", flvUrl)
	httpReqStr += fmt.Sprintf("Host: %s\r\n", "127.0.0.1")
	httpReqStr += "User-Agent: Stream-proxy\r\n"
	httpReqStr += "Accept: */*\r\n"
	httpReqStr += "Accept-Encoding: gzip, deflate\r\n"
	httpReqStr += "Accept-Language: zh-CN,zh;q=0.9,ja;q=0.8,en;q=0.7\r\n"
	// httpReqStr += "Origin: null\r\n"
	httpReqStr += "\r\n"
	addr, err := net.ResolveTCPAddr("tcp", flvServerIP)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	// send http request
	conn.Write([]byte(httpReqStr))

	var hasHttpResponeHeader bool
	var httpResponeHeader []byte
	var nextChunkSize int64

	buff := make([]byte, 0) // response buff
	tmp := make([]byte, 0xfffff)
	for {
		// read from server
		n, err := conn.Read(tmp)
		fmt.Println("read:", n, err)
		if err != nil {
			break
		}

		// copy one packet to buff
		buff = append(buff, tmp[:n]...)

		// parse http response header
		if !hasHttpResponeHeader {
			for i := 0; i < n-4; i++ {
				if buff[i] == 0x0d && buff[i+1] == 0x0a && buff[i+2] == 0x0d && buff[i+3] == 0x0a {
					httpResponeHeader = buff[:i+4]
					hasHttpResponeHeader = true

					fmt.Println("find http response header ", hasHttpResponeHeader, httpResponeHeader)
					fmt.Println(string(httpResponeHeader))

					// move buff ptr
					buff = buff[i+4:]
					break
				}
			}
		}

		// parse chunk
		// [chunk size][\r\n][chunk data][\r\n][chunk size][\r\n][chunk data][\r\n]
		var length = len(buff)

		// no sufficient chunk size
		if nextChunkSize > 0 {
			if int64(length)+7 < nextChunkSize {
				continue
			}
		}

		if buff[0] == 0x0d && buff[1] == 0x0a {
			buff = buff[2:]
		}

		for i := 0; i < length-1; i++ {
			if buff[i] == 0x0d && buff[i+1] == 0x0a {
				// get chunk size
				offset, _ := strconv.ParseInt(string(buff[:i]), 16, 64)
				nextChunkSize = offset

				if len(buff[int64(i)+2:]) < int(nextChunkSize) {
					break
				}

				// write to file
				n, err := fp.Write(buff[i+2 : int64(i)+2+offset])
				fmt.Println("write to file:", n, err)

				// move buf ptr
				nextChunkSize = 0
				buff = buff[int64(i)+2+offset:]
				break
			}

		}
	}
}
