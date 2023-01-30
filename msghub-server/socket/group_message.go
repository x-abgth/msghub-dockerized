package socket

type WSMessage struct {
	Type    string   `json:"type"`
	Payload GMessage `json:"payload"`
	ClientID string
	TargetID string
}
