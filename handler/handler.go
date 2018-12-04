package handler

import (
	"log"
	"net/http"
	"os"

	"github.com/VolticFroogo/Chat-Anonymous/models"
	"github.com/VolticFroogo/Chat-Anonymous/rooms"
	"github.com/go-recaptcha/recaptcha"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/zemirco/uid"
)

var (
	captchaSecret = os.Getenv("CAPTCHA_SECRET")
	captcha       = recaptcha.New(captchaSecret)
	upgrader      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// Start the server by handling the web server.
func Start() {
	r := mux.NewRouter()
	r.StrictSlash(true)

	r.Handle("/ws", http.HandlerFunc(WebsocketHandler))

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Printf("Server started...")
	http.ListenAndServe(":86", r)
}

// WebsocketHandler handles incoming websocket connection requests.
func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Make sure we close the connection when the function returns.
	defer ws.Close()

	var message models.Message

	for {
		message = models.Message{}
		err = ws.ReadJSON(&message)
		if err != nil {
			ws.WriteJSON(models.Message{Success: false})
			continue
		}

		if message.Type == models.MTLifeCheck {
			// It was a life check, ignore it.
			continue
		}

		if message.Captcha == "" {
			ws.WriteJSON(models.Message{Success: false})
			continue
		}
		captchaSuccess, serr := captcha.Verify(message.Captcha, r.Header.Get("CF-Connecting-IP")) // Check the captcha.
		if serr != nil {
			ws.WriteJSON(models.Message{Success: false})
			log.Println(serr)
			continue
		}
		if !captchaSuccess {
			ws.WriteJSON(models.Message{Success: false})
			continue
		}

		break
	}

	user := models.User{
		UUID:     uid.New(64),
		Username: message.Username,
	}

	roomName := message.Room
	if _, ok := rooms.Rooms[message.Room]; !ok {
		room := rooms.Room{
			Name:       roomName,
			Broadcast:  make(chan models.Message, 3),
			NewUser:    make(chan models.User, 3),
			RemoveUser: make(chan string, 3),
		}
		rooms.Rooms[message.Room] = room
		go rooms.StartBroadcaster(&room)
	}

	rooms.Rooms[roomName].NewUser <- models.User{
		WS:       ws,
		UUID:     user.UUID,
		Username: message.Username,
	}

	for {
		message = models.Message{}
		err = ws.ReadJSON(&message)
		if err != nil {
			rooms.Rooms[roomName].RemoveUser <- user.UUID

			return
		}

		if message.Type == models.MTLifeCheck {
			// It was a life check, ignore it.
			continue
		}

		message.Type = models.MTMessage
		message.UserUUID = user.UUID

		rooms.Rooms[roomName].Broadcast <- message
	}
}
