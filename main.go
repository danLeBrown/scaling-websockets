package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const port = ":3000"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var ctx = context.Background()
var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan []byte)            // Broadcast channel
var mutex = &sync.Mutex{}                    // Protect clients map

func main() {
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := os.Getenv("REDIS_DB")

	// log.Println("REDIS_HOST", redisHost)
	// log.Println("REDIS_PORT", redisPort)
	// log.Println("REDIS_PASSWORD", redisPassword)
	// log.Println("REDIS_DB", redisDB)

	db, err := strconv.Atoi(redisDB)
	if err != nil {
		log.Println("Error converting Redis DB to int:", err)
		return
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       db,
	})

	sub := rdb.Subscribe(ctx, "chat-app")

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/chat-app", appHandler)
	http.HandleFunc("/ws", wsHandler)
	go publishMessages(rdb)
	go subscribeMessages(sub)
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
	}
}

func publishMessages(rdb *redis.Client) {
	defer rdb.Close()

	for {
		// Grab the next message from the broadcast channel
		message := <-broadcast

		// Send the message to all connected clients
		mutex.Lock()

		err := rdb.Publish(ctx, "chat-app", message).Err()
		if err != nil {
			log.Println("Error publishing message:", err)
			mutex.Unlock()
			return
		}

		mutex.Unlock()
	}
}

func subscribeMessages(sub *redis.PubSub) {
	defer sub.Close()

	for {
		// message := <-broadcast
		msg, err := sub.ReceiveMessage(ctx)

		if err != nil {
			log.Println("Error publishing message:", err)
			return
		}

		// clients is a list of connections
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}

		// log.Println("Message:", msg.Payload)
	}
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	// Read the HTML file
	htmlContent, err := os.ReadFile("public/index.html")
	if err != nil {
		log.Printf("Error reading index.html: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get APPID from environment variable
	appID := os.Getenv("APPID")
	if appID == "" {
		appID = "N/A"
	}

	// Replace placeholder with actual APPID
	htmlString := string(htmlContent)
	htmlString = strings.ReplaceAll(htmlString, "{{APPID}}", appID)

	// Set content type and serve
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlString))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
