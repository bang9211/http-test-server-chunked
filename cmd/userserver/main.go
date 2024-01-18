package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		rc.EnableFullDuplex()

		fmt.Println(r.Proto, r.Method, r.URL.Path)
		for k, v := range r.Header {
			fmt.Println(k, v)
		}
		fmt.Println()
		go readAsync(r)
		writeAsyncWithScanln(w, rc)
		r.Body.Close()

		fmt.Println("============================================")
	})

	fmt.Println("Serving...")
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func readAsync(r *http.Request) {
	buf := make([]byte, 1024)
	for {
		if r.Close {
			fmt.Println("Connection closed")
			break
		}
		n, err := r.Body.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Read error =>", err.Error())
				break
			}
			fmt.Print(string(buf[:n]))
			break
		}
		fmt.Print(string(buf[:n]))
	}
	fmt.Println("Read done.")
}

func writeAsyncWithScanln(w io.Writer, wc *http.ResponseController) {
	for {
		var input string
		fmt.Scanln(&input)
		if len(input) == 0 {
			break
		}
		fmt.Fprintln(w, input)
		wc.Flush()
	}
	fmt.Println("Write done.")
}
