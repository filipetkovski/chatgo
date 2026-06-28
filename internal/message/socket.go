package message

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var connections = make(map[string]*websocket.Conn)
var mu sync.Mutex

func SocketHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		// Upgrade HTTP to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open socket"})
			return
		}

		// Add connection to map
		mu.Lock()
		connections[userID] = conn
		mu.Unlock()

		db.Exec(`UPDATE users SET is_online = true WHERE id = $1`, userID)
		fmt.Println("User connected:", userID)

		// Remove connection when user disconnects
		defer func() {
			mu.Lock()
			delete(connections, userID)
			mu.Unlock()
			db.Exec(`UPDATE users SET is_online = false WHERE id = $1`, userID)
			conn.Close()
			fmt.Println("User connected:", userID)
		}()

		// Send messages
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				break
			}

			// Get receiver's ID
			var receiverID string
			err = db.QueryRow(`SELECT id FROM users WHERE username = $1`, msg.ReceiverUsername).Scan(&receiverID)
			if err != nil {
				continue
			}

			// Check if they are friends
			var friendshipExists bool
			err = db.QueryRowContext(c.Request.Context(), `
				SELECT EXISTS (
					SELECT 1 FROM friendships
					WHERE (user_id = $1 AND friend_id = $2)
					OR (user_id = $2 AND friend_id = $1)
					AND status = 'accepted'
				)`, userID, receiverID).Scan(&friendshipExists)

			if err != nil || !friendshipExists {
				conn.WriteJSON(gin.H{"error": "You are not friends with this user"})
				continue
			}

			// Save message to database
			db.Exec(`INSERT INTO messages (sender_id, receiver_id, content)
			VALUES ($1, $2, $3)`, userID, receiverID, msg.Content)

			// Send to receiver if online
			mu.Lock()
			receiverConn, online := connections[receiverID]
			mu.Unlock()

			if online {
				receiverConn.WriteJSON(gin.H{
					"sender_username":   msg.SenderUsername,
					"receiver_username": msg.ReceiverUsername,
					"content":           msg.Content,
				})
			}
		}
	}
}
