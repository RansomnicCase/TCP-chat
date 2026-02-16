package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	conn net.Conn    //tcp socket connection
	name string      //username
	ch   chan string //mailbox
}

var (
	clients  = make(map[net.Conn]*Client) //database of who all are online
	join     = make(chan *Client)
	leave    = make(chan net.Conn)
	incoming = make(chan string)
	mutex    = &sync.Mutex{} //safety lock
	rdb      *redis.Client
	ctx      = context.Background()
)

func main() {
	//connecting to redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatal("red lowkey dead bro", err)
	}
	log.Println("conneted to redis")
	//start the server
	listener, err := net.Listen("tcp", ":8080") //open the port
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	fmt.Println("server running on port 8080")
	go localManager()
	go redisSubcriber()
	for {
		conn, err := listener.Accept() //blocking call
		if err != nil {
			continue
		}
		go handleConn(conn)
	}
}

func localManager() {
	for {
		select {
		case msg := <-incoming:
			rdb.Publish(ctx, "global chat", msg)
		case client := <-join:
			mutex.Lock()
			clients[client.conn] = client
			mutex.Unlock()
			rdb.Publish(ctx, "global chat", "🟢 "+client.name+" joined")
		case conn := <-leave:
			mutex.Lock()
			if client, ok := clients[conn]; ok {
				delete(clients, conn)
				close(client.ch)
				rdb.Publish(ctx, "global chat", "🔴 "+client.name+"left")

			}
			mutex.Unlock()
		}
	}
}

func redisSubcriber() {
	pubsub := rdb.Subscribe(ctx, "global chat")
	defer pubsub.Close()
	ch := pubsub.Channel()
	for msg := range ch {
		mutex.Lock()
		for _, client := range clients {
			client.ch <- msg.Payload
		}
		mutex.Unlock()
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	conn.Write([]byte("Enter your alias:\n"))
	if !scanner.Scan() {
		return
	}
	name := strings.TrimSpace(scanner.Text())
	client := &Client{conn: conn, name: name, ch: make(chan string)}
	go func() {
		for msg := range client.ch {
			conn.Write([]byte(msg + "\n"))
		}
	}()
	join <- client
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		incoming <- fmt.Sprintf("[%s]:%s", name, text)
	}
	leave <- conn
}
