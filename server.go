package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/redis.v3"
	"log"
	"net/http"
	"os"
	"strings"
)


var addr = flag.String("addr", "localhost:8080", "http service address")

var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
	DB:   0,
})

var upgrader = websocket.Upgrader{}
var clientConn []*websocket.Conn
var count int

func OperationNotify(w http.ResponseWriter, r *http.Request) {
	log.Println("URL:", r.URL.Path, "Method:", r.Method)
	if r.URL.Path != "/" {
		http.Error(w, "Not Found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	clientConn = append(clientConn, c)
	count += 1
	if err != nil {
		log.Fatal("upgrade:", err)
		return
	}
	fmt.Println("No--->", count)
	defer c.Close()

	//keep-alive
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv:%s", string(message))

		if strings.EqualFold(string(message), "Server,") {
			err = c.WriteMessage(mt, []byte("Server,,"))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}
func RedisSubscribe() {
	//redisClient subscribe Channel
	pubsub, err := redisClient.Subscribe("Channel-Name")
	if err != nil {
		log.Fatal("Sub:", err)
	}
	//Redis notify
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			log.Println("Rev:", err)
		}
		for _, c := range clientConn {
			err = c.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				log.Println("notify-all-client:", err)
				continue
			}
		}
		log.Println(msg.Channel, msg.Payload)
	}
}

func main() {
	fmt.Println("Go Operation is running...")
	flag.Parse()
	go RedisSubscribe()
	http.HandleFunc("/", OperationNotify)
	err := http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
