package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
	"todo-app/config"
	"todo-app/handlers"
	"todo-app/middleware"
	"todo-app/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	userRepo := repository.NewUserRepository(config.DB)
	todoRepo := repository.NewTodoRepository(config.DB)
	authHandler := handlers.NewAuthHandler(userRepo)
	todoHandler := handlers.NewTodoHandler(todoRepo)

	r := gin.New()

	auth := r.Group("/auth")
	auth.GET("/github", authHandler.GitHubLogin)
	auth.GET("/github/callback", authHandler.GitHubCallback)

	todos := r.Group("/todos", middleware.JWTAuth())
	todos.GET("", todoHandler.ListTodos)
	todos.POST("", todoHandler.CreateTodo)
	todos.DELETE("/:id", todoHandler.DeleteTodo)
	todos.PATCH("/:id/complete", todoHandler.CompleteTodo)

	return r
}

func setupTestDB(t *testing.T) {
	t.Helper()
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=tododb_test sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping integration test: db unreachable: %v", err)
	}
	config.DB = db

	db.Exec(`DROP TABLE IF EXISTS todos`)
	db.Exec(`DROP TABLE IF EXISTS users`)
	config.InitDB()
}

// makeTestToken 直接构造 JWT，绕过 GitHub OAuth 流程
func makeTestToken(userID int) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// insertTestUser 插入测试用户，返回 user id
func insertTestUser(t *testing.T, githubID int64) int {
	t.Helper()
	var id int
	err := config.DB.QueryRow(
		`INSERT INTO users (github_id, name) VALUES ($1, $2)
		 ON CONFLICT (github_id) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
		githubID, "testuser",
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	return id
}

func TestListTodosEmpty(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	userID := insertTestUser(t, 1001)
	token, _ := makeTestToken(userID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/todos", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTodoCRUD(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	userID := insertTestUser(t, 1002)
	token, _ := makeTestToken(userID)
	authHeader := "Bearer " + token

	// Create
	todoBody, _ := json.Marshal(map[string]string{"title": "Buy groceries"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var createResp map[string]map[string]interface{}
	json.NewDecoder(w.Body).Decode(&createResp)
	todoID := int(createResp["todo"]["id"].(float64))

	// List
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/todos", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Complete
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/todos/"+strconv.Itoa(todoID)+"/complete", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Delete
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/todos/"+strconv.Itoa(todoID), nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/todos", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
