package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const port = ":8080"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan []byte)            // Broadcast channel
var mutex = &sync.Mutex{}                    // Protect clients map

func main() {
	http.HandleFunc("/app/", appHandler)
	http.HandleFunc("/ws", wsHandler)
	go handleMessages()
	http.ListenAndServe(port, nil)
	log.Println("Server is running on port", port)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	handleConnection(conn)
}

func handleConnection(conn *websocket.Conn) {
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			log.Println("Error reading message:", err)
			break
		}
		log.Printf("Received: %s\n", message)

		broadcast <- message

		// if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		// 	log.Println("Error writing message:", err)
		// 	break
		// }
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		message := <-broadcast

		// Send the message to all connected clients
		mutex.Lock()

		// clients is a list of connections
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}
