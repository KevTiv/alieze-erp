package main

import (
	"context"
	"fmt"
	"log"

	"github.com/KevTiv/alieze-erp/pkg/auth"
	"github.com/KevTiv/alieze-erp/pkg/policy"
	"github.com/google/uuid"
)

func main() {
	// Example: Using the new authorization service with policy engine

	// 1. Create a policy engine (in this example, we'll use a mock mode)
	// In production, you would use NewCasbinEnforcer with a database connection
	casbinEnforcer, err := policy.NewCasbinEnforcer("", "") // Empty strings for mock mode
	if err != nil {
		log.Fatalf("Failed to create casbin enforcer: %v", err)
	}

	// 2. Create the policy engine with the casbin enforcer
	policyEngine := policy.NewEngineWithCasbin(casbinEnforcer)

	// 3. Create the authorization service
	authService := auth.NewBasicAuthorizationService(policyEngine)

	// 4. Example usage
	ctx := context.Background()
	userID := uuid.Must(uuid.NewV7())
	orgID := uuid.Must(uuid.NewV7())

	// Test permission checking
	err = authService.CheckPermission(ctx, userID, orgID, "contacts:create")
	if err != nil {
		fmt.Printf("Permission check failed: %v\n", err)
	} else {
		fmt.Println("Permission granted: contacts:create")
	}

	// Test resource access checking
	contactID := uuid.Must(uuid.NewV7())
	err = authService.CheckResourceAccess(ctx, userID, orgID, "contacts", contactID)
	if err != nil {
		fmt.Printf("Resource access check failed: %v\n", err)
	} else {
		fmt.Println("Resource access granted: contacts")
	}

	// Test organization access checking
	hasAccess := authService.HasOrganizationAccess(ctx, userID, orgID)
	if hasAccess {
		fmt.Println("Organization access granted")
	} else {
		fmt.Println("Organization access denied")
	}

	// Test getting user roles
	roles, err := authService.GetUserRoles(ctx, userID, orgID)
	if err != nil {
		fmt.Printf("Failed to get user roles: %v\n", err)
	} else {
		fmt.Printf("User roles: %v\n", roles)
	}

	// Test getting user permissions
	permissions, err := authService.GetUserPermissions(ctx, userID, orgID)
	if err != nil {
		fmt.Printf("Failed to get user permissions: %v\n", err)
	} else {
		fmt.Printf("User permissions: %v\n", permissions)
	}

	fmt.Println("\nAuthorization service example completed successfully!")
}
