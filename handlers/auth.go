package handlers

import (
	"context"
	"ezqueue/auth"
	"ezqueue/common"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	App         *common.App
	Providers   map[string]auth.Provider
	RefreshRepo *auth.RefreshTokenRepo
	UserRepo    *auth.UserRepo
}

func NewAuthHandler(application *common.App, providers map[string]auth.Provider) *AuthHandler {
	return &AuthHandler{
		App:         application,
		Providers:   providers,
		RefreshRepo: &auth.RefreshTokenRepo{Client: application.FSClient},
		UserRepo:    &auth.UserRepo{Client: application.FSClient},
	}
}

func (h *AuthHandler) JWTAuth(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		c.AbortWithStatus(401)
		return
	}

	token := strings.TrimPrefix(header, "Bearer ")
	claims, err := auth.ParseAccessToken(token)
	if err != nil {
		c.AbortWithStatus(401)
		return
	}

	c.Set("user_id", claims.UserID)
	c.Set("roles", claims.Roles)
	c.Next()

}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {

	c.JSON(200, gin.H{
		"user_id": c.GetString("user_id"),
	})

	//userID := c.GetString("user_id")
	//doc, err := h.App.FSClient.Collection("users").Doc(userID).Get(c.Request.Context())
	//if err != nil {
	//	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	//	return
	//}
	//
	//var user models.User
	//if err := doc.DataTo(&user); err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user"})
	//	return
	//}
	//
	//c.JSON(http.StatusOK, user)
}

type LoginRequest struct {
	Provider string `json:"provider"`
	Token    string `json:"token"`
}

func (a *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, common.GetErrorResponse("Bad Request"))
		return
	}

	provider, ok := a.Providers[req.Provider]
	if !ok {
		c.JSON(400, common.GetErrorResponse("Unknown provider"))
		return
	}

	userInfo, err := provider.Verify(context.Background(), req.Token)
	if err != nil {
		c.JSON(401, common.GetErrorResponse("Invalid token"))
		return
	}

	userID, err := a.UserRepo.FindOrCreateUser(
		context.Background(),
		*userInfo,
	)
	if err != nil {
		c.JSON(500, common.GetErrorResponse("Failed to find or create user"))
		return
	}

	roles := []string{"user"}

	access, _ := auth.GenerateAccessToken(userID, userInfo.Email, roles)
	refreshToken, expires := auth.GenerateRefreshToken()
	hash := auth.HashToken(refreshToken)

	_ = a.RefreshRepo.Save(c, hash, userID, expires)

	c.JSON(200, gin.H{
		"access_token":  access,
		"refresh_token": refreshToken,
	})
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (a *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, common.GetErrorResponse("Bad request"))
		return
	}

	tokenHash := auth.HashToken(req.RefreshToken)

	rt, err := a.RefreshRepo.Get(context.Background(), tokenHash)
	if err != nil {
		// ❗ token не найден → возможно reuse → можно инвалидировать все refresh пользователя
		c.JSON(401, common.GetErrorResponse("Invalid Refresh token"))
		return
	}

	// TODO: verify signature

	if time.Now().After(rt.ExpiresAt) {
		a.RefreshRepo.Delete(context.Background(), tokenHash)
		c.JSON(401, common.GetErrorResponse("Refresh token expired"))
		return
	}

	// 🔥 ROTATION
	_ = a.RefreshRepo.Delete(context.Background(), tokenHash)

	// new tokens
	access, _ := auth.GenerateAccessToken(rt.UserID, "", []string{"user"})
	newRefresh, expires := auth.GenerateRefreshToken()
	newHash := auth.HashToken(newRefresh)

	_ = a.RefreshRepo.Save(
		context.Background(),
		newHash,
		rt.UserID,
		expires,
	)

	c.JSON(200, gin.H{
		"access_token":  access,
		"refresh_token": newRefresh,
	})
}
