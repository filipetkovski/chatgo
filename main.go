package main

import (
	"chatgo/db"
	"chatgo/internal/auth"

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

	r.Run(":8080")
}
