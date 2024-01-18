package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func main() {
	rd, wr := io.Pipe()
	u, _ := url.Parse("http://localhost:8888")

	req := &http.Request{
		Method:     "POST",
		ProtoMajor: 1,
		ProtoMinor: 1,
		URL:        u,
		// ContentLength: 10,
		TransferEncoding: []string{"chunked"},
		Body:             rd,
		Header:           make(map[string][]string),
	}

	client := http.DefaultClient

	go writeWithScanln(wr)

	fmt.Println("Requesting...")
	resp, err := client.Do(req)
	if nil != err {
		fmt.Println("Do error =>", err.Error())
		return
	}
	defer resp.Body.Close()

	readAsync(resp.Body)

	fmt.Println(resp.Proto, resp.Status)
	for k, v := range resp.Header {
		fmt.Println(k, v)
	}
}

func readAsync(rd io.ReadCloser) {
	buf := make([]byte, 1024)
	for {
		n, err := rd.Read(buf)
		if nil != err {
			if err == io.EOF {
				break
			}
			fmt.Println("Read error =>", err.Error())
			break
		}
		if n == 0 {
			break
		}
		fmt.Print(string(buf[:n]))
	}
	fmt.Println("Read done.")
}

func writeWithScanln(wr io.WriteCloser) {
	defer wr.Close()
	for {
		var input string
		fmt.Scanln(&input)
		if len(input) == 0 {
			break
		}
		fmt.Fprintln(wr, input)
	}
	fmt.Println("Write done.")
}
