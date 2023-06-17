package main

import (
	"fmt"
	"net/http"
	"os"
)

// first : go run flv-server-example.go

//There is a benefit to establishing a connection through HTTP
//Removed HTTP response header
//Parsed chunk data
//The data read is the original data sent by the server

// chunk data format :[chunk size][\r\n][chunk data][\r\n][chunk size][\r\n][chunk data][\r\n]

// flv addr
var reqUrl = "http://localhost:13370/flv.flv"

func main() {

	resp, err := http.Get(reqUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// The saved flv file can be played directly
	fp, _ := os.OpenFile("./flv_output.flv", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	buff := make([]byte, 0xfffff)
	for {
		n, err := resp.Body.Read(buff)
		fmt.Println(n, err)
		if err != nil {
			break
		}
		fp.Write(buff[:n])
	}
}
