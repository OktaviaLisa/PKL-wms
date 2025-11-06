package handlers

import (
	"fmt"
	"net/http"
	"time"
	"wms-backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) LoginGin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	var user models.User
	var hashedPassword string
	err := h.DB.QueryRow(
		"SELECT id, username, email, password, is_staff, roles, first_name, last_name, is_active FROM auth_user WHERE username = $1 AND is_active = true",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword, &user.IsStaff, &user.Role, &user.FirstName, &user.LastName, &user.IsActive)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := fmt.Sprintf("token_%d_%d", user.ID, time.Now().Unix())
	response := models.TokenResponse{
		Access:  token,
		Refresh: token,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) RegisterGin(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	var userID int
	err = h.DB.QueryRow(
		"INSERT INTO auth_user (username, email, password, first_name, last_name, roles, is_staff, is_superuser, is_active, date_joined) VALUES ($1, $2, $3, $4, $5, $6, false, false, true, NOW()) RETURNING id",
		req.Username, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.Role,
	).Scan(&userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	user := models.User{
		ID:        userID,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		IsStaff:   false,
		IsActive:  true,
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"message": "User created successfully",
	})
}