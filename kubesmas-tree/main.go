package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"main/persistence/redis"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var CLIENT_NUM = "CLIENT_NUM"

func reader(conn *websocket.Conn) {
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("recv: ", string(p))
		p = []byte(handler(p))
		log.Println("res: ", string(p))
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "please connect via WebSocket")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	redisPath := os.Getenv("REDIS_PATH")

	client, err := redis.New(redisPath)
	if err != nil {
		log.Printf("failed to get redis client: %v\n", err)
		return
	}
	defer client.Close()

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client Connected")

	err = client.Get(CLIENT_NUM).Err()
	if err == redis.Nil {
		log.Println("CLIENT_NUM does not exist. creating now...")

		err = client.Set(CLIENT_NUM, 1, time.Hour*24).Err()
		if err != nil {
			log.Printf("failed to set CLIENT_NUM: %v\n", err)
			return
		}
	} else if err != nil {
		log.Printf("failed to get Client_NUM: %v\n", err)
	} else {
		currentNum, err := client.Incr(CLIENT_NUM).Result()
		if err != nil {
			log.Printf("failed to incr CLIENT_NUM: %v\n", err)
		}
		log.Printf("currentNum: %d\n", currentNum)
	}

	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}
	reader(ws)
	log.Println("Client Disconnected.")
	currentNum, err := client.Decr(CLIENT_NUM).Result()
	if err != nil {
		fmt.Printf("failed to decr CLIENT_NUM: %v\n", err)
		return
	}
	fmt.Printf("Successfully decrement CLIENT_NUM.\ncurrent num is :%d\n", currentNum)
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Hello World!")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), nil))
}
