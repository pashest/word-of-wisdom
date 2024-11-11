package model

import "fmt"

type MessageType int

// Header of message
const (
	Unknown MessageType = iota //
	RequestChallenge
	ResponseChallenge
	RequestResource
	ResponseResource
)

// Message - message struct for both server and client
type Message struct {
	Type      MessageType
	RequestID string
	Payload   string
}

// Stringify - stringify message to send it by tcp-connection
// divider between header and payload is |
func (m *Message) Stringify() string {
	return fmt.Sprintf("%d|%s|%s", m.Type, m.RequestID, m.Payload)
}
