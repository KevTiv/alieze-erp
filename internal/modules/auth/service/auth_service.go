package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/KevTiv/alieze-erp/internal/modules/auth/types"
	"github.com/KevTiv/alieze-erp/internal/modules/auth/repository"
	"github.com/KevTiv/alieze-erp/internal/modules/auth/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   repository.AuthRepository
	logger *log.Logger
}

var (
	accessTokenExp = time.Hour * 24 // 24 hours - should match JWT utils
)

func NewAuthService(repo repository.AuthRepository) *AuthService {
	return &AuthService{
		repo:   repo,
		logger: log.New(log.Writer(), "auth-service: ", log.LstdFlags),
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, req types.RegisterRequest) (*types.UserProfile, error) {
	// Validate email format
	if !isValidEmail(req.Email) {
		return nil, errors.New("invalid email format")
	}

	// Check if user already exists
	existingUser, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	now := time.Now()
	user := types.User{
		ID:                uuid.New(),
		Email:             req.Email,
		EncryptedPassword: hashedPassword,
		EmailConfirmedAt:  &now,
		ConfirmedAt:       &now,
		CreatedAt:         now,
		UpdatedAt:         now,
		IsSuperAdmin:      false,
	}

	createdUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create organization
	orgID, err := s.repo.CreateOrganization(ctx, req.OrganizationName, createdUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Create organization user (owner role)
	orgUser := types.OrganizationUser{
		ID:             uuid.New(),
		OrganizationID: *orgID,
		UserID:         createdUser.ID,
		Role:           "owner",
		IsActive:       true,
		JoinedAt:       now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err = s.repo.CreateOrganizationUser(ctx, orgUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization user: %w", err)
	}

	s.logger.Printf("User registered successfully: %s (organization: %s)", createdUser.ID, *orgID)

	return &types.UserProfile{
		ID:           createdUser.ID,
		Email:        createdUser.Email,
		IsSuperAdmin: createdUser.IsSuperAdmin,
	}, nil
}

func (s *AuthService) LoginUser(ctx context.Context, req types.LoginRequest) (*types.LoginResponse, error) {
	// Find user by email
	user, err := s.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check password
	if err := checkPassword(req.Password, user.EncryptedPassword); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Get user's organizations
	orgUsers, err := s.repo.FindOrganizationUsersByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	if len(orgUsers) == 0 {
		return nil, errors.New("user has no organization access")
	}

	// For now, use the first organization (in production, we'd need to handle multiple orgs)
	orgUser := orgUsers[0]

	// Update last sign in time
	now := time.Now()
	user.LastSignInAt = &now
	user.UpdatedAt = now

	_, err = s.repo.UpdateUser(ctx, *user)
	if err != nil {
		return nil, fmt.Errorf("failed to update last sign in: %w", err)
	}

	// Generate JWT tokens
	jwtService := utils.NewJWTService()
	accessToken, err := jwtService.GenerateAccessToken(user.ID, orgUser.OrganizationID, orgUser.Role, user.IsSuperAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := jwtService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresIn := int(accessTokenExp.Seconds()) // Convert duration to seconds

	s.logger.Printf("User logged in successfully: %s (organization: %s)", user.ID, orgUser.OrganizationID)

	return &types.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
		User: types.UserProfile{
			ID:           user.ID,
			Email:        user.Email,
			IsSuperAdmin: user.IsSuperAdmin,
		},
	}, nil
}

func (s *AuthService) GetUserProfile(ctx context.Context, userID uuid.UUID) (*types.UserProfile, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return &types.UserProfile{
		ID:           user.ID,
		Email:        user.Email,
		IsSuperAdmin: user.IsSuperAdmin,
	}, nil
}

func (s *AuthService) GetOrganizationID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	orgUsers, err := s.repo.FindOrganizationUsersByUserID(ctx, userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	if len(orgUsers) == 0 {
		return uuid.Nil, errors.New("user has no organization access")
	}

	// Return the first organization ID
	return orgUsers[0].OrganizationID, nil
}

func (s *AuthService) GetUserRole(ctx context.Context, userID, orgID uuid.UUID) (string, error) {
	orgUser, err := s.repo.FindOrganizationUser(ctx, orgID, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get organization user: %w", err)
	}
	if orgUser == nil {
		return "", errors.New("user not found in organization")
	}

	return orgUser.Role, nil
}

func (s *AuthService) CheckPermission(ctx context.Context, userID, orgID uuid.UUID, permission string) error {
	// For now, implement basic permission checking
	// In production, this would check against the permission system

	role, err := s.GetUserRole(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to get user role: %w", err)
	}

	// Super admins and owners have all permissions
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user != nil && user.IsSuperAdmin {
		return nil
	}

	if role == "owner" || role == "admin" {
		return nil
	}

	// For other roles, we'd check specific permissions
	// This is a placeholder for the full permission system
	return errors.New("permission denied")
}

// Helper functions
func isValidEmail(email string) bool {
	return len(email) >= 5 && strings.Contains(email, "@") && strings.Contains(email, ".")
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	// Add more password validation rules as needed
	return nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

func checkPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
