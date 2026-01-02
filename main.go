package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const port = ":8080"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/app/", appHandler)
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(port, nil)
	log.Println("Server is running on port", port)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go handleConnection(conn)
}

func handleConnection(conn *websocket.Conn) {
	// defer conn.Close()

	// for {
	// 	messageType, p, err := conn.ReadMessage()
	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}

	// 	log.Println(string(p))
	// 	if err := conn.WriteMessage(messageType, p); err != nil {
	// 		log.Println(err)
	// 		return
	// 	}
	// }

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		log.Printf("Received: %s\n", message)

		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}
