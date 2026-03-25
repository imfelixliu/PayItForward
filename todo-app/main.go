package main

import (
	"log"
	"os"
	"todo-app/config"
	"todo-app/handlers"
	"todo-app/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.GET("/github", handlers.GitHubLogin)
		auth.GET("/github/callback", handlers.GitHubCallback)
	}

	todos := r.Group("/todos", middleware.JWTAuth())
	{
		todos.GET("", handlers.ListTodos)
		todos.POST("", handlers.CreateTodo)
		todos.DELETE("/:id", handlers.DeleteTodo)
		todos.PATCH("/:id/complete", handlers.CompleteTodo)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
