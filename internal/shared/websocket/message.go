package websocket

import "encoding/json"

// Message is a generic WebSocket message
type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}
