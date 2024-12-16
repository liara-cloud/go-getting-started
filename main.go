package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Upgrader to upgrade HTTP connection to WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// Client struct to manage WebSocket connection
type Client struct {
	conn *websocket.Conn
	send chan []byte
}

// MessageHub to manage all WebSocket connections
type MessageHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func newHub() *MessageHub {
	return &MessageHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (hub *MessageHub) run() {
	for {
		select {
		case client := <-hub.register:
			hub.mu.Lock()
			hub.clients[client] = true
			hub.mu.Unlock()
		case client := <-hub.unregister:
			hub.mu.Lock()
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				close(client.send)
			}
			hub.mu.Unlock()
		case message := <-hub.broadcast:
			hub.mu.Lock()
			for client := range hub.clients {
				select {
				case client.send <- message:
				default:
					delete(hub.clients, client)
					close(client.send)
				}
			}
			hub.mu.Unlock()
		}
	}
}

func (client *Client) read(hub *MessageHub) {
	defer func() {
		hub.unregister <- client
		client.conn.Close()
	}()
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			break
		}
		hub.broadcast <- message
	}
}

func (client *Client) write() {
	defer client.conn.Close()
	for message := range client.send {
		err := client.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}

func serves(hub *MessageHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256)}
	hub.register <- client

	go client.read(hub)
	go client.write()
}

func main() {
	hub := newHub()
	go hub.run()

	http.HandleFunc("/wss", func(w http.ResponseWriter, r *http.Request) {
		serves(hub, w, r)
	})
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	fmt.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
