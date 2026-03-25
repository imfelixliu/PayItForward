package main

import (
	"log"
	"os"
	"todo-app/config"
	"todo-app/handlers"
	"todo-app/middleware"
	"todo-app/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDB()

	userRepo := repository.NewUserRepository(config.DB)
	todoRepo := repository.NewTodoRepository(config.DB)

	authHandler := handlers.NewAuthHandler(userRepo)
	todoHandler := handlers.NewTodoHandler(todoRepo)

	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.GET("/github", authHandler.GitHubLogin)
		auth.GET("/github/callback", authHandler.GitHubCallback)
	}

	todos := r.Group("/todos", middleware.JWTAuth())
	{
		todos.GET("", todoHandler.ListTodos)
		todos.POST("", todoHandler.CreateTodo)
		todos.DELETE("/:id", todoHandler.DeleteTodo)
		todos.PATCH("/:id/complete", todoHandler.CompleteTodo)
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
