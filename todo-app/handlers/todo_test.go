package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"todo-app/config"
	"todo-app/handlers"
	"todo-app/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	auth := r.Group("/auth")
	auth.POST("/register", handlers.Register)
	auth.POST("/login", handlers.Login)

	todos := r.Group("/todos", middleware.JWTAuth())
	todos.GET("", handlers.ListTodos)
	todos.POST("", handlers.CreateTodo)
	todos.DELETE("/:id", handlers.DeleteTodo)
	todos.PATCH("/:id/complete", handlers.CompleteTodo)

	return r
}

func setupTestDB(t *testing.T) {
	t.Helper()
	// Requires TEST_DB_DSN env var or a running postgres instance
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=tododb_test sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping integration test: db unreachable: %v", err)
	}
	config.DB = db

	// Clean slate
	db.Exec(`DROP TABLE IF EXISTS todos`)
	db.Exec(`DROP TABLE IF EXISTS users`)
	config.InitDB()
}

func TestRegisterAndLogin(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	body, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "secret123"})

	// Register
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Login
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["token"] == "" {
		t.Fatal("expected token in response")
	}
}

func TestTodoCRUD(t *testing.T) {
	setupTestDB(t)
	r := setupRouter()

	// Register + login to get token
	creds, _ := json.Marshal(map[string]string{"email": "crud@example.com", "password": "secret123"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(creds))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(creds))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var loginResp map[string]string
	json.NewDecoder(w.Body).Decode(&loginResp)
	token := loginResp["token"]

	authHeader := "Bearer " + token

	// Create todo
	todoBody, _ := json.Marshal(map[string]string{"title": "Buy groceries"})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var createResp map[string]map[string]interface{}
	json.NewDecoder(w.Body).Decode(&createResp)
	todoID := int(createResp["todo"]["id"].(float64))

	// List todos
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/todos", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Complete todo
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, "/todos/"+itoa(todoID)+"/complete", nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Delete todo
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/todos/"+itoa(todoID), nil)
	req.Header.Set("Authorization", authHeader)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
