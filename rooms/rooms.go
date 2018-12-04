package rooms

import (
	"github.com/VolticFroogo/Chat-Anonymous/models"
)

var Rooms = make(map[string]Room)

type Room struct {
	Name       string
	Broadcast  chan models.Message
	NewUser    chan models.User
	RemoveUser chan string
}

// StartBroadcaster starts the thread for a channel's broadcasting.
func StartBroadcaster(room *Room) {
	users := []models.User{}

	for {
		select {
		case message := <-room.Broadcast:
			for i, user := range users {
				err := user.WS.WriteJSON(message)
				if err != nil {
					users = append(users[:i], users[i+1:]...)
				}
			}
			break

		case user := <-room.NewUser:
			err := user.WS.WriteJSON(models.Message{
				Success: true,
				Users:   &users,
			})
			if err != nil {
				break
			}

			users = append(users, user)

			room.Broadcast <- models.Message{
				Type: models.MTUserConnect,
				User: &user,
			}
			break

		case userUUID := <-room.RemoveUser:
			for i, user := range users {
				if user.UUID == userUUID {
					users = append(users[:i], users[i+1:]...)
				}
			}

			if len(users) <= 0 {
				delete(Rooms, room.Name)
				return
			}

			room.Broadcast <- models.Message{
				Type:     models.MTUserDisconnect,
				UserUUID: userUUID,
			}
			break
		}
	}
}
