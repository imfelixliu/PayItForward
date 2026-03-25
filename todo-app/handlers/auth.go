package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"todo-app/apperror"
	"todo-app/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
}

func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{userRepo: userRepo}
}

// GET /auth/github
func (h *AuthHandler) GitHubLogin(c *gin.Context) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&scope=user:email",
		clientID,
	)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// GET /auth/github/callback
func (h *AuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.Error(apperror.New(http.StatusBadRequest, "INVALID_INPUT", "missing code"))
		return
	}

	accessToken, err := exchangeCodeForToken(code)
	if err != nil {
		c.Error(apperror.New(http.StatusInternalServerError, "OAUTH_ERROR", "failed to exchange token"))
		return
	}

	ghUser, err := fetchGitHubUser(accessToken)
	if err != nil {
		c.Error(apperror.New(http.StatusInternalServerError, "OAUTH_ERROR", "failed to fetch github user"))
		return
	}

	githubID := int64(ghUser["id"].(float64))
	name, _ := ghUser["name"].(string)
	email, _ := ghUser["email"].(string)
	avatarURL, _ := ghUser["avatar_url"].(string)

	user, err := h.userRepo.Upsert(githubID, email, name, avatarURL)
	if err != nil {
		c.Error(apperror.ErrInternal)
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		c.Error(apperror.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func exchangeCodeForToken(code string) (string, error) {
	params := url.Values{}
	params.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	params.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	params.Set("code", code)

	req, _ := http.NewRequest(http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(params.Encode()),
	)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("github oauth error: %s", result.Error)
	}
	return result.AccessToken, nil
}

func fetchGitHubUser(accessToken string) (map[string]interface{}, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var user map[string]interface{}
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}
	return user, nil
}

func generateToken(userID int) (string, error) {
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
