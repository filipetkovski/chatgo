package message

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetMessageHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		friendUsername := c.Query("username")

		var friendID string
		err := db.QueryRowContext(c.Request.Context(), `
		SELECT id FROM users WHERE username = $1`, friendUsername).Scan(&friendID)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		rows, err := db.QueryContext(c.Request.Context(), `
		SELECT content, sender_id, sent_at FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
		ORDER BY sent_at ASC`, userID, friendID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer rows.Close()

		var messages []gin.H
		for rows.Next() {
			var content, senderID, sentAT string
			rows.Scan(&content, &senderID, &sentAT)
			messages = append(messages, gin.H{
				"content":   content,
				"sender_id": senderID,
				"sent_at":   sentAT,
			})
		}

		c.JSON(http.StatusOK, gin.H{"messages": messages})
	}
}
