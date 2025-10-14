package model

type Message struct {
	Sender      string `json:"sender" binding:"required"`
	Destination string `json:"destination" binding:"required"`
	Content     string `json:"content" binding:"required"`
}
