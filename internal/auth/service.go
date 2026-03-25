package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"taip-flow-backend/internal/config"
	"taip-flow-backend/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenInvalid       = errors.New("token is invalid or expired")
)

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

type Service struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

// ── Registration ─────────────────────────────────────────────────────────────

type RegisterInput struct {
	Name     string `json:"name"     binding:"required,min=2"`
	Email    string `json:"email"    binding:"required,email"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required,min=8"`
}

func (s *Service) Register(input RegisterInput) (*models.User, error) {
	var existing models.User
	if err := s.db.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Name:         input.Name,
		Email:        input.Email,
		Phone:        input.Phone,
		PasswordHash: string(hash),
		Provider:     models.ProviderLocal,
		IsActive:     true,
	}
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// ── Login ─────────────────────────────────────────────────────────────────────

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (s *Service) Login(input LoginInput) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ? AND provider = ?", input.Email, models.ProviderLocal).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return &user, nil
}

// ── Token generation ──────────────────────────────────────────────────────────

func (s *Service) IssueAccessToken(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWTAccessMins) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "taip-flow",
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) IssueRefreshToken(userID string) (rawToken string, err error) {
	// Generate 32 random bytes as the raw token
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	rawToken = hex.EncodeToString(b)

	// Store hash in DB
	h := sha256.Sum256([]byte(rawToken))
	rt := models.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: hex.EncodeToString(h[:]),
		ExpiresAt: time.Now().AddDate(0, 0, s.cfg.JWTRefreshDays),
	}
	err = s.db.Create(&rt).Error
	return
}

// ── Token validation ──────────────────────────────────────────────────────────

func (s *Service) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

func (s *Service) RotateRefreshToken(rawToken string) (*models.User, string, error) {
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])

	var rt models.RefreshToken
	if err := s.db.Where("token_hash = ? AND expires_at > ?", hash, time.Now()).First(&rt).Error; err != nil {
		return nil, "", ErrTokenInvalid
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", rt.UserID).Error; err != nil {
		return nil, "", ErrUserNotFound
	}

	// Revoke old token
	s.db.Delete(&rt)

	// Issue new refresh token
	newRaw, err := s.IssueRefreshToken(user.ID)
	if err != nil {
		return nil, "", err
	}
	return &user, newRaw, nil
}

func (s *Service) RevokeRefreshToken(rawToken string) {
	h := sha256.Sum256([]byte(rawToken))
	hash := hex.EncodeToString(h[:])
	s.db.Where("token_hash = ?", hash).Delete(&models.RefreshToken{})
}

func (s *Service) GetUserByID(id string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}
