package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService(t *testing.T) {
	svc := NewJWTService()
	userID := uuid.New()
	orgID := uuid.New()
	role := "admin"
	isSuperAdmin := false

	t.Run("Generate and validate access token", func(t *testing.T) {
		// Generate token
		token, err := svc.GenerateAccessToken(userID, orgID, role, isSuperAdmin)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate token
		claims, err := svc.ValidateToken(token)
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, orgID, claims.OrganizationID)
		assert.Equal(t, role, claims.Role)
		assert.Equal(t, isSuperAdmin, claims.IsSuperAdmin)
		assert.Equal(t, userID.String(), claims.Subject)
		assert.Equal(t, "alieze-erp", claims.Issuer)
	})

	t.Run("Generate and validate refresh token", func(t *testing.T) {
		// Generate refresh token
		refreshToken, err := svc.GenerateRefreshToken(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshToken)

		// Validate refresh token (it should validate as a standard JWT)
		token, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})
		require.NoError(t, err)
		require.True(t, token.Valid)

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		require.True(t, ok)
		assert.Equal(t, userID.String(), claims.Subject)
		assert.Equal(t, "alieze-erp", claims.Issuer)
	})

	t.Run("Invalid token validation", func(t *testing.T) {
		// Test with invalid token
		_, err := svc.ValidateToken("invalid.token.here")
		assert.Error(t, err)

		// Test with empty token
		_, err = svc.ValidateToken("")
		assert.Error(t, err)
	})

	t.Run("Expired token validation", func(t *testing.T) {
		// Save original expiration
		originalAccessExp := accessTokenExp
		originalRefreshExp := refreshTokenExp
		defer func() {
			accessTokenExp = originalAccessExp
			refreshTokenExp = originalRefreshExp
		}()

		// Set very short expiration for testing
		accessTokenExp = -1 * time.Hour // Expired 1 hour ago
		refreshTokenExp = -1 * time.Hour

		// Generate expired token
		svc := NewJWTService()
		expiredToken, err := svc.GenerateAccessToken(userID, orgID, role, isSuperAdmin)
		require.NoError(t, err)

		// Validation should fail
		_, err = svc.ValidateToken(expiredToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("Token with wrong issuer", func(t *testing.T) {
		// Save original issuer and secret
		originalIssuer := jwtIssuer
		originalSecret := jwtSecretKey
		defer func() {
			jwtIssuer = originalIssuer
			jwtSecretKey = originalSecret
		}()

		// Change issuer for testing
		jwtIssuer = "different-issuer"

		// Generate token with different issuer
		svc := NewJWTService()
		token, err := svc.GenerateAccessToken(userID, orgID, role, isSuperAdmin)
		require.NoError(t, err)

		// Restore original issuer for validation
		jwtIssuer = originalIssuer

		// Validation should fail
		_, err = svc.ValidateToken(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "issuer")
	})

	t.Run("GetTokenClaims without validation", func(t *testing.T) {
		svc := NewJWTService()
		token, err := svc.GenerateAccessToken(userID, orgID, role, isSuperAdmin)
		require.NoError(t, err)

		// Get claims without full validation
		claims, err := svc.GetTokenClaims(token)
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, orgID, claims.OrganizationID)
	})

	t.Run("GetTokenClaims with invalid token", func(t *testing.T) {
		_, err := svc.GetTokenClaims("invalid.token")
		assert.Error(t, err)
	})
}

func TestJWTConfiguration(t *testing.T) {
	t.Run("SetJWTSecretKey", func(t *testing.T) {
		originalSecret := jwtSecretKey
		defer func() { jwtSecretKey = originalSecret }()

		newSecret := "new-secret-key-for-testing"
		SetJWTSecretKey(newSecret)
		assert.Equal(t, []byte(newSecret), jwtSecretKey)

		svc := NewJWTService()
		userID := uuid.New()
		orgID := uuid.New()

		// Generate token with new secret
		token, err := svc.GenerateAccessToken(userID, orgID, "admin", false)
		require.NoError(t, err)

		// Should validate with new secret
		_, err = svc.ValidateToken(token)
		require.NoError(t, err)

		// Should fail with old secret
		jwtSecretKey = originalSecret
		_, err = svc.ValidateToken(token)
		assert.Error(t, err)
	})

	t.Run("SetTokenExpirations", func(t *testing.T) {
		originalAccessExp := accessTokenExp
		originalRefreshExp := refreshTokenExp
		defer func() {
			accessTokenExp = originalAccessExp
			refreshTokenExp = originalRefreshExp
		}()

		// Set custom expirations
		customAccessExp := time.Hour * 1
		customRefreshExp := time.Hour * 24 * 30
		SetTokenExpirations(customAccessExp, customRefreshExp)

		assert.Equal(t, customAccessExp, accessTokenExp)
		assert.Equal(t, customRefreshExp, refreshTokenExp)

		svc := NewJWTService()
		userID := uuid.New()
		orgID := uuid.New()

		// Generate token with custom expiration
		token, err := svc.GenerateAccessToken(userID, orgID, "admin", false)
		require.NoError(t, err)

		// Validate token
		claims, err := svc.ValidateToken(token)
		require.NoError(t, err)
		require.NotNil(t, claims)

		// Check that expiration is approximately 1 hour from now
		now := time.Now()
		expiresAt := claims.ExpiresAt.Time
		assert.WithinDuration(t, now.Add(customAccessExp), expiresAt, time.Minute)
	})
}
