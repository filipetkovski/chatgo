package message

type Message struct {
	SenderUsername   string `json:"sender_username"`
	ReceiverUsername string `json:"receiver_username"`
	Content          string `json:"content"`
}
