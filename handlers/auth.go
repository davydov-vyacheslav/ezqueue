package handlers

import (
	"ezqueue/app"
	"ezqueue/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	App *app.App
}

func NewAuthHandler(application *app.App) *AuthHandler {
	return &AuthHandler{App: application}
}

type GoogleAuthRequest struct {
	IDToken string `json:"idToken" binding:"required"`
}

func (h *AuthHandler) HandleGoogleAuth(c *gin.Context) {
	var req GoogleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.App.AuthClient.VerifyIDToken(c.Request.Context(), req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Get or create user
	userDoc := h.App.FSClient.Collection("users").Doc(token.UID)
	doc, err := userDoc.Get(c.Request.Context())

	var user models.User
	if err != nil {
		// Create new user
		user = models.User{
			ID:          token.UID,
			Email:       token.Claims["email"].(string),
			DisplayName: token.Claims["name"].(string),
			PhotoURL:    token.Claims["picture"].(string),
			CreatedAt:   time.Now(),
		}

		_, err = userDoc.Set(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	} else {
		if err := doc.DataTo(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": req.IDToken,
	})
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("userID")
	doc, err := h.App.FSClient.Collection("users").Doc(userID).Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
