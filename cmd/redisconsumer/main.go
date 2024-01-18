package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-redis/redis/v8"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
	})
	defer redisClient.Close()

	rd, wr := io.Pipe()
	u, _ := url.Parse("http://localhost:9999")

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

	go readRedisWriteHTTPStream(wr, redisClient)

	fmt.Println("Receving...")
	client := http.DefaultClient
	resp, err := client.Do(req)
	if nil != err {
		fmt.Println("Do error =>", err.Error())
		return
	}
	defer resp.Body.Close()

	readHTTPWriteRedisStreamAsync(resp.Body, redisClient)

	fmt.Println(resp.Proto, resp.Status)
	for k, v := range resp.Header {
		fmt.Println(k, v)
	}
}

func readHTTPWriteRedisStreamAsync(rd io.ReadCloser, redisClient *redis.Client) {
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

		// write to redis pubsub
		redisClient.Publish(context.Background(), "response", string(buf[:n]))
	}
	fmt.Println("Read done.")
	redisClient.Publish(context.Background(), "response", string(""))
}

func readRedisWriteHTTPStream(wr io.WriteCloser, redisClient *redis.Client) {
	defer wr.Close()

	for {
		res, err := redisClient.XRead(context.Background(), &redis.XReadArgs{
			Streams: []string{"request", "0"},
			Count:   1,
			Block:   0,
		}).Result()
		if nil != err {
			fmt.Println("XRead error =>", err.Error())
			break
		}
		if len(res) == 0 {
			continue
		}

		_, err = redisClient.XDel(context.Background(), "request", res[0].Messages[0].ID).Result()
		if nil != err {
			fmt.Println("XDel error =>", err.Error())
			break
		}

		done := false
		for _, msg := range res[0].Messages {
			if msg.Values["message"] == "" {
				done = true
				break
			}
			fmt.Println("Writing to http request =>", msg.Values["message"])
			fmt.Fprintln(wr, msg.Values["message"])
		}
		if done {
			break
		}
	}
	fmt.Println("Write done.")
}
