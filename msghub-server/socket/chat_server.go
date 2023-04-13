package socket

import (
	"fmt"
	"log"
	"os"

	"github.com/x-abgth/msghub-dockerized/msghub-server/logic"
	"github.com/x-abgth/msghub-dockerized/msghub-server/models"
)

//	Because our ChatServer acts like a hub for connecting the parts of our chat application,
//	we will use it to keep track of all the rooms that will be created.

type WsServer struct {
	users      []models.UserModel
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *WSMessage
}

// NewWebSocketServer :- First we create this server.
func NewWebSocketServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *WSMessage),
	}
}

// Run our websocket server, accepting various requests
// This function will run finely and listens to the channels
func (server *WsServer) Run() {
	for {
		select {
		case client := <-server.register:
			server.registerClient(client)
		case client := <-server.unregister:
			server.unregisterClient(client)
		case message := <-server.broadcast:
			server.broadcastToClients(message)
		}
	}
}

// If a client is joined we will make the map value to true.
func (server *WsServer) registerClient(client *Client) {

	server.clients[client] = true
}

// If the client is left from the socket, we will delete the client key and value.
func (server *WsServer) unregisterClient(client *Client) {
	fmt.Println("unregistered")
	delete(server.clients, client)
}

// If the client send a message, it broadcasts to all the other users
func (server *WsServer) broadcastToClients(m *WSMessage) {
	defer func() {
		if e := recover(); e != nil {
			log.Println("Error happened in sending message")
			log.Println(e)
			os.Exit(1)
		}
	}()

	user := server.findClientByID(m.ClientID)
	target := server.findClientByID(m.TargetID)

	if target != nil {
		if user != nil {
			if server.findRoomByID(m.TargetID) == m.ClientID {
				if m.Type == "message" {
					data := models.MessageModel{
						To:          m.Payload.Room,
						From:        m.Payload.By,
						Content:     m.Payload.Body,
						ContentType: logic.TEXT,
						Status:      logic.IS_READ,
						Time:        m.Payload.Time,
					}
					userLogic.StorePersonalMessagesLogic(data)
				} else if m.Type == "image" {
					data := models.MessageModel{
						To:          m.Payload.Room,
						From:        m.Payload.By,
						Content:     m.Payload.Body,
						ContentType: logic.IMAGE,
						Status:      logic.IS_READ,
						Time:        m.Payload.Time,
					}
					userLogic.StorePersonalMessagesLogic(data)
				}
				m.Payload.Status = logic.IS_READ
			} else {
				if m.Type == "message" {
					data := models.MessageModel{
						To:          m.Payload.Room,
						From:        m.Payload.By,
						Content:     m.Payload.Body,
						ContentType: logic.TEXT,
						Status:      logic.IS_DELIVERED,
						Time:        m.Payload.Time,
					}
					userLogic.StorePersonalMessagesLogic(data)
				} else if m.Type == "image" {
					data := models.MessageModel{
						To:          m.Payload.Room,
						From:        m.Payload.By,
						Content:     m.Payload.Body,
						ContentType: logic.IMAGE,
						Status:      logic.IS_DELIVERED,
						Time:        m.Payload.Time,
					}
					userLogic.StorePersonalMessagesLogic(data)
				}
				m.Payload.Status = logic.IS_DELIVERED
			}
			user.send <- m
			target.send <- m
		} else {
			if m.Type == "message" {
				data := models.MessageModel{
					To:          m.Payload.Room,
					From:        m.Payload.By,
					Content:     m.Payload.Body,
					ContentType: logic.TEXT,
					Status:      logic.IS_NOT_SENT,
					Time:        m.Payload.Time,
				}
				userLogic.StorePersonalMessagesLogic(data)
			} else if m.Type == "image" {
				data := models.MessageModel{
					To:          m.Payload.Room,
					From:        m.Payload.By,
					Content:     m.Payload.Body,
					ContentType: logic.IMAGE,
					Status:      logic.IS_NOT_SENT,
					Time:        m.Payload.Time,
				}
				userLogic.StorePersonalMessagesLogic(data)
			}
			user.send <- m
		}
	} else {
		if user != nil {
			if m.Type == "message" {
				data := models.MessageModel{
					To:          m.Payload.Room,
					From:        m.Payload.By,
					Content:     m.Payload.Body,
					ContentType: logic.TEXT,
					Status:      logic.IS_SENT,
					Time:        m.Payload.Time,
				}
				userLogic.StorePersonalMessagesLogic(data)
			} else if m.Type == "image" {
				data := models.MessageModel{
					To:          m.Payload.Room,
					From:        m.Payload.By,
					Content:     m.Payload.Body,
					ContentType: logic.IMAGE,
					Status:      logic.IS_SENT,
					Time:        m.Payload.Time,
				}
				userLogic.StorePersonalMessagesLogic(data)
			}
			m.Payload.Status = logic.IS_SENT
			user.send <- m
		} else {
			if m.Type == "message" {
				data := models.MessageModel{
					To:          m.Payload.Room,
					From:        m.Payload.By,
					Content:     m.Payload.Body,
					ContentType: logic.TEXT,
					Status:      logic.IS_NOT_SENT,
					Time:        m.Payload.Time,
				}
				userLogic.StorePersonalMessagesLogic(data)
			} else if m.Type == "image" {
				data := models.MessageModel{
					To:          m.Payload.Room,
					From:        m.Payload.By,
					Content:     m.Payload.Body,
					ContentType: logic.IMAGE,
					Status:      logic.IS_NOT_SENT,
					Time:        m.Payload.Time,
				}
				userLogic.StorePersonalMessagesLogic(data)
			}
			user.send <- m
		}
	}

}

func (server *WsServer) findClientByID(ID string) *Client {
	var foundClient *Client = nil
	for client := range server.clients {
		if client.ID == ID {
			foundClient = client
			break
		}
	}
	return foundClient
}

func (server *WsServer) findRoomByID(target string) string {
	client := server.findClientByID(target)
	if client == nil {
		log.Println("Client is NIL")
		return ""
	}
	return client.Room
}
