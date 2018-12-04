package models

import "github.com/gorilla/websocket"

// Message types
const (
	MTMessage = iota + 1
	MTUserConnect
	MTUserDisconnect
	MTFile
	MTLifeCheck
)

// User is a user connected in a chat.
type User struct {
	UUID, Username string
	WS             *websocket.Conn `json:",omitempty"`
}

// Message is a message sent or received via Websockets.
type Message struct {
	Type     int     `json:",omitempty"`
	Success  bool    `json:",omitempty"`
	Message  string  `json:",omitempty"`
	Captcha  string  `json:",omitempty"`
	Username string  `json:",omitempty"`
	Room     string  `json:",omitempty"`
	User     *User   `json:",omitempty"`
	UserUUID string  `json:",omitempty"`
	Users    *[]User `json:",omitempty"`
}
