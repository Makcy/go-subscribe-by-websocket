package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var dialer = websocket.Dialer{} // use default options

func main() {
	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func() {
		c.Close()
		log.Println("Main Close")
	}()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv:%s", string(message))
			defer c.Close()
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := c.WriteMessage(websocket.TextMessage, []byte("Server,"))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
