package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
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

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client Connected")

	err = addValue(CLIENT_NUM)
	if err != nil {
		log.Println(err)
		ws.WriteMessage(1, []byte("Failed to addValue. close connection."))
		return
	}

	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}

	reader(ws)

	log.Println("Client Disconnected.")
	currentNum, err := declValue(CLIENT_NUM)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("Successfully decrement CLIENT_NUM.\ncurrent num is :%d\n", currentNum)
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/iot", iotEndpoint)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Hello World!")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}
