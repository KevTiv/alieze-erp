package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/KevTiv/alieze-erp/internal/modules/auth/utils"

	"github.com/google/uuid"
)

// AuthMiddleware is a middleware that validates JWT tokens and sets user context
type AuthMiddleware struct {
	jwtService *utils.JWTService
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: utils.NewJWTService(),
	}
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public routes
		if isPublicRoute(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Set user context
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		ctx = context.WithValue(ctx, "organizationID", claims.OrganizationID)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "isSuperAdmin", claims.IsSuperAdmin)

		// Set organization ID in context for database operations
		ctx = context.WithValue(ctx, "orgID", claims.OrganizationID)

		// Continue with the request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// isPublicRoute checks if the route should be accessible without authentication
func isPublicRoute(path string) bool {
	publicRoutes := []string{
		"/auth/register",
		"/auth/login",
		"/health",
		"/",
	}

	for _, route := range publicRoutes {
		if path == route {
			return true
		}
	}

	return false
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value("userID").(uuid.UUID)
	return userID, ok
}

// GetOrganizationIDFromContext extracts organization ID from context
func GetOrganizationIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	orgID, ok := ctx.Value("organizationID").(uuid.UUID)
	return orgID, ok
}

// GetRoleFromContext extracts role from context
func GetRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value("role").(string)
	return role, ok
}

// GetIsSuperAdminFromContext extracts super admin status from context
func GetIsSuperAdminFromContext(ctx context.Context) (bool, bool) {
	isSuperAdmin, ok := ctx.Value("isSuperAdmin").(bool)
	return isSuperAdmin, ok
}
