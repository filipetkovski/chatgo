package friendship

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddFriendHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddFriendRequest
		c.ShouldBindJSON(&req)

		userID := c.GetString("user_id")

		var friendID string
		err := db.QueryRowContext(c.Request.Context(), `
			SELECT id FROM users WHERE username = $1`,
			req.Username).Scan(&friendID)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		_, err = db.ExecContext(c.Request.Context(), `
            INSERT INTO friendships (user_id, friend_id, status)
            VALUES ($1, $2, 'pending')`,
			userID, friendID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "friend request sent"})
	}
}

func AcceptFriendRequest(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddFriendRequest
		c.ShouldBindJSON(&req)

		userID := c.GetString("user_id")

		_, err := db.ExecContext(c.Request.Context(), `
		UPDATE friendships SET status = 'accepted'
		WHERE friend_id = $1 AND user_id = (
		SELECT id FROM users WHERE username = $2)`,
			userID, req.Username)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Friend request accepted"})
	}
}

func GetFriendRequests(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		rows, err := db.QueryContext(c.Request.Context(), `
		SELECT u.username FROM users u
		JOIN friendships f ON (f.user_id = u.id)
		WHERE f.friend_id = $1`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		defer rows.Close()

		var requestedUsernames []string
		for rows.Next() {
			var username string
			rows.Scan(&username)
			requestedUsernames = append(requestedUsernames, username)
		}

		c.JSON(http.StatusOK, gin.H{"usernames": requestedUsernames})
	}
}

func ActiveFriendsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		rows, err := db.QueryContext(c.Request.Context(), `
            SELECT u.username FROM users u
            JOIN friendships f ON (f.friend_id = u.id OR f.user_id = u.id)
            WHERE (f.user_id = $1 OR f.friend_id = $1)
            AND f.status = 'accepted'
            AND u.is_online = true
            AND u.id != $1`,
			userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer rows.Close()

		var friends []string
		for rows.Next() {
			var username string
			rows.Scan(&username)
			friends = append(friends, username)
		}

		c.JSON(http.StatusOK, gin.H{"friends": friends})
	}
}
