package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"todo-app/config"
	"todo-app/models"

	"github.com/gin-gonic/gin"
)

type createTodoRequest struct {
	Title string `json:"title" binding:"required,min=1,max=500"`
}

// ListTodos godoc
// GET /todos
func ListTodos(c *gin.Context) {
	userID := c.GetInt("user_id")

	rows, err := config.DB.Query(
		`SELECT id, user_id, title, completed, created_at, updated_at FROM todos WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch todos"})
		return
	}
	defer rows.Close()

	todos := []models.Todo{}
	for rows.Next() {
		var t models.Todo
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan todo"})
			return
		}
		todos = append(todos, t)
	}

	c.JSON(http.StatusOK, gin.H{"todos": todos})
}

// CreateTodo godoc
// POST /todos
func CreateTodo(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req createTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var todo models.Todo
	err := config.DB.QueryRow(
		`INSERT INTO todos (user_id, title) VALUES ($1, $2) RETURNING id, user_id, title, completed, created_at, updated_at`,
		userID, req.Title,
	).Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"todo": todo})
}

// DeleteTodo godoc
// DELETE /todos/:id
func DeleteTodo(c *gin.Context) {
	userID := c.GetInt("user_id")
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo id"})
		return
	}

	result, err := config.DB.Exec(
		`DELETE FROM todos WHERE id = $1 AND user_id = $2`,
		todoID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete todo"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "todo deleted"})
}

// CompleteTodo godoc
// PATCH /todos/:id/complete
func CompleteTodo(c *gin.Context) {
	userID := c.GetInt("user_id")
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid todo id"})
		return
	}

	var todo models.Todo
	err = config.DB.QueryRow(
		`UPDATE todos SET completed = TRUE, updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, title, completed, created_at, updated_at`,
		todoID, userID,
	).Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "todo not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"todo": todo})
}
