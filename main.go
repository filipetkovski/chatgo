package main

import (
	"chatgo/db"
	"chatgo/internal/auth"
	"chatgo/internal/friendship"

	"github.com/gin-gonic/gin"
)

func main() {
	database := db.Connect()

	r := gin.Default()
	r.POST("/auth/register", auth.RegisterHandler(database))
	r.POST("/auth/login", auth.LoginHandler(database))

	// Protected — token required for ALL of these
	protected := r.Group("/")
	protected.Use(auth.AuthMiddleware())
	protected.POST("/friends/add", friendship.AddFriendHandler(database))
	protected.POST("/friends/accept", friendship.AcceptFriendRequest(database))
	protected.GET("/friends/active", friendship.ActiveFriendsHandler(database))

	r.Run(":8080")
}
