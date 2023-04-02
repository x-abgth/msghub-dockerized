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

type Message struct {
	MsgId       int    `json:"msg_id"`
	FromUserId  string `json:"from_user_id"`
	ToUserId    string `json:"to_user_id"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	SentTime    string `json:"sent_time"`
	Status      string `json:"status"`
	IsRecent    bool   `json:"is_recent"`
}
