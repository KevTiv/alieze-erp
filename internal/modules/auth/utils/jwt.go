package utils

import (
	"errors"
	"fmt"
	"time"

	"alieze-erp/internal/modules/auth/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// JWT configuration
	jwtSecretKey    = []byte("your-secret-key-here-change-in-production") // TODO: Move to config
	accessTokenExp  = time.Hour * 24                                      // 24 hours
	refreshTokenExp = time.Hour * 24 * 7                                  // 7 days
	jwtIssuer       = "alieze-erp"
)

// JWTService provides JWT token operations
type JWTService struct{}

// GenerateAccessToken generates a new JWT access token
func (s *JWTService) GenerateAccessToken(userID, orgID uuid.UUID, role string, isSuperAdmin bool) (string, error) {
	claims := domain.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer,
			Subject:   userID.String(),
		},
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
		IsSuperAdmin:   isSuperAdmin,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// GenerateRefreshToken generates a new JWT refresh token
func (s *JWTService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExp)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    jwtIssuer,
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*domain.TokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != jwtIssuer {
		return nil, errors.New("invalid token issuer")
	}

	// Validate expiration
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// GetTokenClaims extracts claims from a token without full validation
// Useful for checking token type before full validation
func (s *JWTService) GetTokenClaims(tokenString string) (*domain.TokenClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &domain.TokenClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*domain.TokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// SetJWTSecretKey allows setting the JWT secret key (for testing/config)
func SetJWTSecretKey(key string) {
	jwtSecretKey = []byte(key)
}

// SetTokenExpirations allows setting token expiration times
func SetTokenExpirations(accessExp, refreshExp time.Duration) {
	accessTokenExp = accessExp
	refreshTokenExp = refreshExp
}

// NewJWTService creates a new JWT service instance
func NewJWTService() *JWTService {
	return &JWTService{}
}
