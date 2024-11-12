package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pashest/word-of-wisdom/internal/model"
)

// ParseMessage - parses Message from str, checks header and payload
func ParseMessage(str string) (*model.Message, error) {
	str = strings.TrimSpace(str)
	var msgType int
	parts := strings.Split(str, "|")
	if !(len(parts) == 1 || len(parts) == 3) { // only 1 or 3 parts allowed
		return nil, fmt.Errorf("message doesn't match protocol")
	}
	// try to parse header
	msgType, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse header")
	}
	msg := model.Message{
		Type: model.MessageType(msgType),
	}
	// try to parse payload
	if len(parts) == 3 {
		msg.RequestID = parts[1]
		msg.Payload = parts[2]
	}
	return &msg, nil
}
