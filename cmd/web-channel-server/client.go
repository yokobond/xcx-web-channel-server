package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const readLimit = 512 // Maximum message size allowed from client

type Message struct {
	Action  string `json:"action"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	topics map[string]bool
	closed bool
	mu     sync.Mutex
}

func (c *Client) closeSend() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		close(c.send)
		c.closed = true
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.closeSend()
		c.conn.Close()
	}()

	c.conn.SetReadLimit(readLimit)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("error parsing message: %v", err)
			continue
		}

		switch msg.Action {
		case "subscribe":
			if msg.Topic != "" {
				c.hub.subscribe(c, msg.Topic)
				response := Message{
					Action:  "subscribed",
					Topic:   msg.Topic,
					Message: "Successfully subscribed",
				}
				c.mu.Lock()
				if !c.closed {
					if data, err := json.Marshal(response); err == nil {
						c.send <- data
					}
				}
				c.mu.Unlock()
			}

		case "publish":
			if msg.Topic != "" && msg.Message != "" {
				c.hub.publish(msg.Topic, []byte(msg.Message))
			}

		default:
			log.Printf("unknown action: %s", msg.Action)
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.closeSend()
		c.conn.Close()
	}()

	for message := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}

	// If the channel is closed, send a close message
	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}
