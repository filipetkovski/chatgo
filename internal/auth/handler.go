package auth

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		c.ShouldBindJSON(&req)

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
			return
		}

		_, err = db.ExecContext(c.Request.Context(), `
            INSERT INTO users (username, email, password) 
            VALUES ($1, $2, $3)`,
			req.Username, req.Email, string(hashedPassword))

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var userID string
		db.QueryRowContext(c.Request.Context(), `
            SELECT id FROM users WHERE email = $1`, req.Email).Scan(&userID)

		token, err := GenerateToken(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"Token": token})
	}
}

func LoginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		c.ShouldBindJSON(&req)

		var user struct {
			ID           string
			PasswordHash string
		}

		err := db.QueryRowContext(c.Request.Context(), `
            SELECT id, password FROM users WHERE email = $1`,
			req.Email).Scan(&user.ID, &user.PasswordHash)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		_, err = db.ExecContext(c.Request.Context(), `
		UPDATE users SET is_online = true
		WHERE email = $1`, req.Email)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Loggin failed!"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Login successful!"})
	}
}
