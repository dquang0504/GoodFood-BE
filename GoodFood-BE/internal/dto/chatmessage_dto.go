package dto


type ChatMessage struct {
	FromID    int    `json:"from_id"`
	ToID      int    `json:"to_id"`
	Sender    string `json:"sender"` // user or admin
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}