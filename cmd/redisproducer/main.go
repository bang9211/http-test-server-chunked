package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
)

func main() {
	// create new redis client for stream
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
	})
	defer redisClient.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		rc.EnableFullDuplex()

		fmt.Println(r.Proto, r.Method, r.URL.Path)
		for k, v := range r.Header {
			fmt.Println(k, v)
		}
		fmt.Println()
		go readHTTPWriteRedisStream(r, redisClient)
		readRedisWriteHTTPStream(w, rc, redisClient)

		fmt.Println("============================================")
	})

	fmt.Println("Serving...")
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func readHTTPWriteRedisStream(r *http.Request, redisClient *redis.Client) {
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

		redisClient.XAdd(r.Context(), &redis.XAddArgs{
			Stream: "request",
			Values: map[string]interface{}{
				"message": string(buf[:n]),
			},
		})
	}
	fmt.Println("Read done.")
	redisClient.XAdd(r.Context(), &redis.XAddArgs{
		Stream: "request",
		Values: map[string]interface{}{
			"message": string(""),
		},
	})
}

func readRedisWriteHTTPStream(w io.Writer, wc *http.ResponseController, redisClient *redis.Client) {
	for {
		var input string
		// subscribe to redis pubsub
		pubsub := redisClient.Subscribe(context.Background(), "response")
		msg, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			fmt.Println("Read error =>", err.Error())
			break
		}
		input = msg.Payload
		if len(input) == 0 {
			break
		}
		fmt.Println("Reading from redis stream =>", input)
		fmt.Fprintln(w, input)
		wc.Flush()
	}
	fmt.Println("Write done.")
}
