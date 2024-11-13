// main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	topics     map[string]map[*Client]bool
	sync.RWMutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, check origin
	},
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		topics:     make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Printf("Client %v disconnected", client.conn.RemoteAddr())
				delete(h.clients, client)
				client.closeSend() // Ensure closeSend is called only once
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					client.closeSend() // Ensure closeSend is called only once
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) subscribe(client *Client, topic string) {
	h.Lock()
	defer h.Unlock()

	if _, ok := h.topics[topic]; !ok {
		h.topics[topic] = make(map[*Client]bool)
	}
	h.topics[topic][client] = true
	client.topics[topic] = true
}

func (h *Hub) publish(topic string, message []byte) {
	h.RLock()
	defer h.RUnlock()

	if clients, ok := h.topics[topic]; ok {
		for client := range clients {
			client.mu.Lock()
			if !client.closed {
				client.send <- message
			}
			client.mu.Unlock()
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		topics: make(map[string]bool),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	// Send periodic pings to the client
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			client.mu.Lock()
			if client.closed {
				client.mu.Unlock()
				return
			}
			err := client.conn.WriteMessage(websocket.PingMessage, nil)
			client.mu.Unlock()
			if err != nil {
				return
			}
		}
	}()
}

func main() {
	enableWSS := flag.Bool("wss", false, "Enable secure WebSocket (WSS)")
	configFile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	config := &Config{
		WSPort:  8080, // default values
		WSSPort: 8443,
	}

	if *enableWSS {
		loadedConfig, err := loadConfig(*configFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("Warning: Config file %s not found. WSS will not be enabled.", *configFile)
				*enableWSS = false
			} else {
				log.Printf("Error loading config: %v. Using default ports (ws:%d)", err, config.WSPort)
			}
		} else {
			config = loadedConfig
		}
	}

	hub := newHub()
	go hub.run()

	http.HandleFunc("/wc", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	if *enableWSS {
		if config.CertFile == "" || config.KeyFile == "" {
			log.Printf("Warning: Certificate or key file path not set in config. WSS will not be enabled.")
		} else {
			// Start WSS server in a goroutine
			go func() {
				log.Printf("Starting secure WebSocket server on wss://localhost:%d", config.WSSPort)
				err := http.ListenAndServeTLS(
					fmt.Sprintf(":%d", config.WSSPort),
					config.CertFile,
					config.KeyFile,
					nil,
				)
				if err != nil {
					log.Printf("Failed to start WSS server: %v", err)
				}
			}()
		}
	}

	// Start regular WS server
	log.Printf("Starting WebSocket server on ws://localhost:%d", config.WSPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.WSPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
