package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"sim_monit/server/config"
	"sim_monit/server/models"
)

type AuthHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid atau password kurang dari 8 karakter"})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))
	cleanName := strings.TrimSpace(req.Name)

	if cleanName == "" || len(cleanName) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama minimal 2 karakter"})
		return
	}

	// Check if email already exists
	var count int
	err := h.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, cleanEmail).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	now := time.Now()
	res, err := h.db.Exec(`
		INSERT INTO users (name, email, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`, cleanName, cleanEmail, string(hashedPassword), now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat akun user"})
		return
	}

	userID, _ := res.LastInsertId()
	user := models.User{
		ID:        userID,
		Name:      cleanName,
		Email:     cleanEmail,
		CreatedAt: now,
	}

	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token autentikasi"})
		return
	}

	// Set secure http-only cookie with SameSite=Lax
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" || h.cfg.GinMode == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", token, 3600*24*7, "/", "", isSecure, true)

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format email atau password tidak valid"})
		return
	}

	cleanEmail := strings.ToLower(strings.TrimSpace(req.Email))

	var user models.User
	var passwordHash string
	err := h.db.QueryRow(`
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = ?
	`, cleanEmail).Scan(&user.ID, &user.Name, &user.Email, &passwordHash, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau password salah"})
		return
	}

	token, err := h.generateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token autentikasi"})
		return
	}

	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" || h.cfg.GinMode == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", token, 3600*24*7, "/", "", isSecure, true)

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.User
	err := h.db.QueryRow(`
		SELECT id, name, email, created_at
		FROM users
		WHERE id = ?
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" || h.cfg.GinMode == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("auth_token", "", -1, "/", "", isSecure, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *AuthHandler) generateToken(userID int64, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
