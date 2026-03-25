package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"todo-app/apperror"
	"todo-app/repository"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	todoRepo *repository.TodoRepository
}

func NewTodoHandler(todoRepo *repository.TodoRepository) *TodoHandler {
	return &TodoHandler{todoRepo: todoRepo}
}

type createTodoRequest struct {
	Title string `json:"title" binding:"required,min=1,max=500"`
}

func (h *TodoHandler) ListTodos(c *gin.Context) {
	userID := c.GetInt("user_id")
	todos, err := h.todoRepo.FindByUserID(userID)
	if err != nil {
		c.Error(apperror.ErrInternal)
		return
	}
	c.JSON(http.StatusOK, gin.H{"todos": todos})
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req createTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.New(http.StatusBadRequest, "INVALID_INPUT", err.Error()))
		return
	}
	todo, err := h.todoRepo.Create(userID, req.Title)
	if err != nil {
		c.Error(apperror.ErrInternal)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"todo": todo})
}

func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	userID := c.GetInt("user_id")
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Error(apperror.New(http.StatusBadRequest, "INVALID_INPUT", "invalid todo id"))
		return
	}
	deleted, err := h.todoRepo.Delete(todoID, userID)
	if err != nil {
		c.Error(apperror.ErrInternal)
		return
	}
	if !deleted {
		c.Error(apperror.New(http.StatusNotFound, "TODO_NOT_FOUND", "todo not found"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "todo deleted"})
}

func (h *TodoHandler) CompleteTodo(c *gin.Context) {
	userID := c.GetInt("user_id")
	todoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Error(apperror.New(http.StatusBadRequest, "INVALID_INPUT", "invalid todo id"))
		return
	}
	todo, err := h.todoRepo.Complete(todoID, userID)
	if err == sql.ErrNoRows {
		c.Error(apperror.New(http.StatusNotFound, "TODO_NOT_FOUND", "todo not found"))
		return
	} else if err != nil {
		c.Error(apperror.ErrInternal)
		return
	}
	c.JSON(http.StatusOK, gin.H{"todo": todo})
}
