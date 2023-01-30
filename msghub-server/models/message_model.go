package models

type MessageModel struct {
	Content     string `json:"message"`
	From        string `json:"from"`
	To          string `json:"to"`
	Time        string `json:"time"`
	Status      string `json:"status"`
	ContentType string `json:"type"`
	Order       float64
}

type GrpMsgModel struct {
	Id          string
	Name        string
	Avatar      string
	Message     string
	Sender      string
	Time        string
	ContentType string
	Order       float64
}
