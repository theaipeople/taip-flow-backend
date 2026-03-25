package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"taip-flow-backend/internal/auth"
	"taip-flow-backend/internal/config"
)

type AuthController struct {
	svc *auth.Service
	cfg *config.Config
}

func NewAuthController(svc *auth.Service, cfg *config.Config) *AuthController {
	return &AuthController{svc: svc, cfg: cfg}
}

// setTokenCookies writes access + refresh tokens as httpOnly cookies.
func (ctrl *AuthController) setTokenCookies(c *gin.Context, accessToken, refreshToken string) {
	secure := ctrl.cfg.CookieSecure
	domain := ctrl.cfg.CookieDomain

	c.SetCookie("access_token", accessToken,
		ctrl.cfg.JWTAccessMins*60, "/", domain, secure, true)

	c.SetCookie("refresh_token", refreshToken,
		ctrl.cfg.JWTRefreshDays*24*3600, "/auth/refresh", domain, secure, true)
}

func (ctrl *AuthController) clearTokenCookies(c *gin.Context) {
	domain := ctrl.cfg.CookieDomain
	c.SetCookie("access_token", "", -1, "/", domain, false, true)
	c.SetCookie("refresh_token", "", -1, "/auth/refresh", domain, false, true)
}

// POST /auth/register
func (ctrl *AuthController) Register(c *gin.Context) {
	var input auth.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.svc.Register(input)
	if err != nil {
		if errors.Is(err, auth.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}

	accessToken, err := ctrl.svc.IssueAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}
	refreshToken, err := ctrl.svc.IssueRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	ctrl.setTokenCookies(c, accessToken, refreshToken)
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

// POST /auth/login
func (ctrl *AuthController) Login(c *gin.Context) {
	var input auth.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ctrl.svc.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	accessToken, err := ctrl.svc.IssueAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}
	refreshToken, err := ctrl.svc.IssueRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	ctrl.setTokenCookies(c, accessToken, refreshToken)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// POST /auth/refresh  — rotates both tokens
func (ctrl *AuthController) Refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie("refresh_token")
	if err != nil || rawRefresh == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no refresh token"})
		return
	}

	user, newRefresh, err := ctrl.svc.RotateRefreshToken(rawRefresh)
	if err != nil {
		ctrl.clearTokenCookies(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired, please log in again"})
		return
	}

	accessToken, err := ctrl.svc.IssueAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	ctrl.setTokenCookies(c, accessToken, newRefresh)
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// POST /auth/logout
func (ctrl *AuthController) Logout(c *gin.Context) {
	if raw, err := c.Cookie("refresh_token"); err == nil {
		ctrl.svc.RevokeRefreshToken(raw)
	}
	ctrl.clearTokenCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// GET /auth/me  — returns current user from access token
func (ctrl *AuthController) Me(c *gin.Context) {
	userID, _ := c.Get(auth.UserIDKey)
	user, err := ctrl.svc.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// GET /auth/google  — stub, wire up oauth2 package when ready
func (ctrl *AuthController) GoogleLogin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"message": "Google OAuth2 not yet configured. Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET in .env.",
	})
}

// GET /auth/google/callback  — stub
func (ctrl *AuthController) GoogleCallback(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Google OAuth2 callback not yet configured."})
}

// helper: cookie expiry as time.Time (unused externally but useful for tests)
func cookieExpiry(seconds int) time.Time {
	return time.Now().Add(time.Duration(seconds) * time.Second)
}
