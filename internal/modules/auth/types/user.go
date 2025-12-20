package types

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// User represents an authenticated user
type User struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Email             string     `json:"email" db:"email"`
	EncryptedPassword string     `json:"-" db:"encrypted_password"`
	EmailConfirmedAt  *time.Time `json:"email_confirmed_at,omitempty" db:"email_confirmed_at"`
	ConfirmedAt       *time.Time `json:"confirmed_at,omitempty" db:"confirmed_at"`
	LastSignInAt      *time.Time `json:"last_sign_in_at,omitempty" db:"last_sign_in_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	IsSuperAdmin      bool       `json:"is_super_admin" db:"is_super_admin"`
	RawUserMetaData   string     `json:"raw_user_meta_data" db:"raw_user_meta_data"`

	// Context fields (populated from token claims or session context)
	OrganizationID uuid.UUID `json:"organization_id,omitempty" db:"-"`
	Role           string    `json:"role,omitempty" db:"-"`
}

// UserProfile represents user profile information
type UserProfile struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	IsSuperAdmin bool      `json:"is_super_admin"`
}

// OrganizationUser represents user organization membership
type OrganizationUser struct {
	ID             uuid.UUID `json:"id" db:"id"`
	OrganizationID uuid.UUID `json:"organization_id" db:"organization_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Role           string    `json:"role" db:"role"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	JoinedAt       time.Time `json:"joined_at" db:"joined_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=8"`
	OrganizationName string `json:"organization_name" validate:"required,min=3"`
	FirstName        string `json:"first_name" validate:"required"`
	LastName         string `json:"last_name" validate:"required"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents successful login response
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	TokenType    string      `json:"token_type"`
	User         UserProfile `json:"user"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	jwt.RegisteredClaims
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Role           string    `json:"role"`
	IsSuperAdmin   bool      `json:"is_super_admin"`
}
