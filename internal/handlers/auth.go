package handlers

import (
	"crm-backend/internal/db"
	"crm-backend/internal/middleware"
	"crm-backend/internal/models"
	"crm-backend/internal/websocket"
	"net/http"

	"crm-backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AppState holds application state
type AppState struct {
	DB               *db.DatabaseManager
	Config           *config.Config
	WebSocketManager *websocket.Manager
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// OtpRequest represents an OTP request
type OtpRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a reset password request
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token      string      `json:"token"`
	ExpiresIn  int64       `json:"expires_in"`
	User       UserResponse `json:"user"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Login handles user login
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, middleware.BadRequest("Invalid request: "+err.Error()))
		return
	}

	state, _ := c.Get("state")
	appState := state.(*AppState)

	// Look up user by email
	var user models.User
	err := appState.DB.ReadPool().QueryRow(
		"SELECT id, name, email, password_hash, role FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		middleware.ErrorResponse(c, middleware.Unauthorized("Invalid email or password"))
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		middleware.ErrorResponse(c, middleware.Unauthorized("Invalid email or password"))
		return
	}

	// Check if user is active
	var status string
	err = appState.DB.ReadPool().QueryRow("SELECT status FROM users WHERE id = $1", user.ID).Scan(&status)
	if err != nil || status != "active" {
		middleware.ErrorResponse(c, middleware.Unauthorized("Account is disabled"))
		return
	}

	// Generate JWT
	token, err := middleware.GenerateToken(user.ID, user.Email, appState.Config.JWTSecret, appState.Config.JWTExpirySecs)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Failed to generate token", err))
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:      token,
		ExpiresIn:  appState.Config.JWTExpirySecs,
		User: UserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	})
}

// Register handles user registration
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, middleware.BadRequest("Invalid request: "+err.Error()))
		return
	}

	state, _ := c.Get("state")
	appState := state.(*AppState)

	// Check if email already exists
	var exists bool
	err := appState.DB.ReadPool().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Database error", err))
		return
	}
	if exists {
		middleware.ErrorResponse(c, middleware.BadRequest("Email already registered"))
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Failed to hash password", err))
		return
	}

	// Create user
	userID := uuid.New()
	_, err = appState.DB.WritePool().Exec(
		"INSERT INTO users (id, name, email, password_hash, role, status) VALUES ($1, $2, $3, $4, 'agent', 'active')",
		userID, req.Name, req.Email, passwordHash,
	)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Failed to create user", err))
		return
	}

	// Generate JWT
	token, err := middleware.GenerateToken(userID, req.Email, appState.Config.JWTSecret, appState.Config.JWTExpirySecs)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Failed to generate token", err))
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:      token,
		ExpiresIn:  appState.Config.JWTExpirySecs,
		User: UserResponse{
			ID:    userID.String(),
			Name:  req.Name,
			Email: req.Email,
			Role:  "agent",
		},
	})
}

// OTP handles OTP verification
func OTP(c *gin.Context) {
	var req OtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, middleware.BadRequest("Invalid request: "+err.Error()))
		return
	}

	// Validate OTP format
	if len(req.OTP) != 6 {
		middleware.ErrorResponse(c, middleware.BadRequest("Invalid OTP format"))
		return
	}

	state, _ := c.Get("state")
	appState := state.(*AppState)

	// Look up user by email
	var user models.User
	err := appState.DB.ReadPool().QueryRow(
		"SELECT id, name, email, role FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role)

	if err != nil {
		middleware.ErrorResponse(c, middleware.Unauthorized("User not found"))
		return
	}

	// Generate JWT (in production, verify OTP from DB/Redis first)
	token, err := middleware.GenerateToken(user.ID, user.Email, appState.Config.JWTSecret, appState.Config.JWTExpirySecs)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Failed to generate token", err))
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token:      token,
		ExpiresIn:  appState.Config.JWTExpirySecs,
		User: UserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	})
}

// Logout handles user logout
func Logout(c *gin.Context) {
	// Stateless JWT: client discards token; optional blacklist in Redis
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Me returns current user info
func Me(c *gin.Context) {
	state, _ := c.Get("state")
	appState := state.(*AppState)

	userID, _ := c.Get("user_id")

	var user models.User
	err := appState.DB.ReadPool().QueryRow(
		"SELECT id, name, email, role FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Role)

	if err != nil {
		middleware.ErrorResponse(c, middleware.Unauthorized("User not found"))
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	})
}

// ForgotPassword handles forgot password requests
func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, middleware.BadRequest("Invalid request: "+err.Error()))
		return
	}

	state, _ := c.Get("state")
	appState := state.(*AppState)

	// Check if user exists
	var exists bool
	err := appState.DB.ReadPool().QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Database error", err))
		return
	}

	if !exists {
		// Don't reveal if email exists
		c.JSON(http.StatusOK, gin.H{"message": "If an account exists, a recovery code has been sent."})
		return
	}

	// TODO: Generate OTP, save to DB/Redis, send email
	// For now, just log
	c.JSON(http.StatusOK, gin.H{"message": "If an account exists, a recovery code has been sent."})
}

// ResetPassword handles password reset
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, middleware.BadRequest("Invalid request: "+err.Error()))
		return
	}

	// Validate OTP format
	if len(req.OTP) != 6 {
		middleware.ErrorResponse(c, middleware.BadRequest("Invalid OTP"))
		return
	}

	state, _ := c.Get("state")
	appState := state.(*AppState)

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Failed to hash password", err))
		return
	}

	// Update password
	result, err := appState.DB.WritePool().Exec(
		"UPDATE users SET password_hash = $1 WHERE email = $2",
		passwordHash, req.Email,
	)
	if err != nil {
		middleware.ErrorResponse(c, middleware.Internal("Failed to update password", err))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		middleware.ErrorResponse(c, middleware.BadRequest("User not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Password updated successfully"})
}
