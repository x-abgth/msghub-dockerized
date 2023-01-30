package socket

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/x-abgth/msghub/msghub-server/logic"
	"github.com/x-abgth/msghub/msghub-server/models"
)

// Hub.
type Hub struct {
	// clients
	Clients map[string]map[*GClient]bool

	// register channel
	Register chan *GClient

	// unregister channel
	Unregister chan *GClient

	// broadcast channel
	Broadcast chan *WSMessage

	// mutex
	mutex *sync.RWMutex
}

// Add client to Hub.
func (h *Hub) add(c *GClient) {
	if _, ok := h.Clients[c.Room]; !ok {
		h.Clients[c.Room] = make(map[*GClient]bool)
	}
	h.Clients[c.Room][c] = true
	log.Printf("Client added to room #%s, Number of clients in room: %d\n", c.Room, len(h.Clients[c.Room]))
}

// Remove client from Hub.
func (h *Hub) delete(c *GClient) {
	if clients, ok := h.Clients[c.Room]; ok {
		delete(clients, c)
		c.Conn.Close()
	}
	var m *WSMessage = &WSMessage{
		Type: "left",
		Payload: GMessage{
			Body: strconv.Itoa(len(h.Clients[c.Room])),
			Room: c.Room,
		},
	}
	h.broadcast(m)
	log.Printf("Removed client from room #%s, Number of clients in Room: %d\n", c.Room, len(h.Clients[c.Room]))
}

// Broadcast message to all connected clients.
func (h *Hub) broadcast(m *WSMessage) {

	if clients, ok := h.Clients[m.Payload.Room]; ok {
		var g logic.GroupDataLogicModel

		if m.Type == "message" {
			data := models.GroupMessageModel{
				GroupId:  m.Payload.Room,
				SenderId: m.Payload.By,
				Content:  m.Payload.Body,
				Type:     logic.TEXT,
				Status:   logic.IS_SENT,
				Time:     m.Payload.Time,
			}
			err := g.InsertMessagesToGroup(data)
			if err != nil {
				log.Fatal("Error happened when inserting elements - ", err)
			}
		} else if m.Type == "image" {
			data := models.GroupMessageModel{
				GroupId:  m.Payload.Room,
				SenderId: m.Payload.By,
				Content:  m.Payload.Body,
				Type:     logic.IMAGE,
				Status:   logic.IS_SENT,
				Time:     m.Payload.Time,
			}
			err := g.InsertMessagesToGroup(data)
			if err != nil {
				log.Fatal("Error happened when inserting elements - ", err)
			}
		}

		for k := range clients {
			k.Send <- m
		}
	} else {
		fmt.Println("ERROR HAPPENED -- ")
	}
}

// Run
func (h *Hub) Run() {
	log.Println("Hub is running")
	for {
		select {
		case client := <-h.Register:
			h.add(client)
		case client := <-h.Unregister:
			h.delete(client)
		case m := <-h.Broadcast:
			h.broadcast(m)
		}
	}
}
