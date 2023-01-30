package socket

import (
	"encoding/json"
	"fmt"
	"log"
)

type Message struct {
	Action    string  `json:"action"`
	Message   string  `json:"message"`
	Time      string  `json:"time"`
	Target    string  `json:"target"`
	Sender    *Client `json:"sender"`
	IsPrivate bool    `json:"is_bool"`
}

type GMessage struct {
	Body   string `json:"body"`
	Time   string `json:"time"`
	By     string `json:"by"`
	Room   string `json:"room"`
	Status string `json:"status"`
}

func (message *Message) encode() []byte {
	jsonStr, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}

	return jsonStr
}

func (message *Message) decode(jsonStr []byte) Message {
	isValid := json.Valid(jsonStr)

	if isValid {
		json.Unmarshal(jsonStr, &message)
		return *message
	} else {
		fmt.Println("Json is not valid!")
		return *message
	}
}
