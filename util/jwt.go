package util

import (
	"fmt"
	"time"

	"demo/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ─── Generic Jwt Provider ────────────────────────────────────────────────────

// CustomClaims holds custom JWT claims along with standard RegisteredClaims.
type CustomClaims struct {
	UserID uuid.UUID       `json:"user_id"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// JWTProvider produces valid tokens for authentification.
type JWTProvider interface {
	Generate(user *domain.User) (string, error)
	Parse(tokenString string) (*CustomClaims, error)
}

// ─── Jwt Implementation ──────────────────────────────────────────────────────

type JWT struct {
	secret []byte
}

// NewJWT prepares a JWT token provider with a secret.
func NewJWT(secret string) *JWT {
	return &JWT{
		secret: []byte(secret),
	}
}

// Generate returns a valid JWT token with a 24-hour lifespan.
func (J *JWT) Generate(user *domain.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &CustomClaims{
		UserID:    user.ID,
		Role:      user.Role,
		Subject:   user.ID.String(),
		Issuer:    "demo-api",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(expirationTime),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(J.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// Parse verifies the validity of a JWT token and returns its information.
func (J *JWT) Parse(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return J.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ─────────────────────────────────────────────────────────────────────────────
