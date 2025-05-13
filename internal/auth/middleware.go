package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/yash3004/user_management_service/internal/schemas"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// UserContextKey is the key for user in context
	UserContextKey ContextKey = "user"
)

// AuthMiddleware authenticates the user and adds user info to the request context
func AuthMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check if the header has the Bearer prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			// Extract the token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			
			// Validate token and get user ID
			userID, err := ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Get user from database
			var user schemas.User
			if err := db.First(&user, "id = ?", userID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					http.Error(w, "User not found", http.StatusUnauthorized)
				} else {
					klog.Errorf("Database error: %v", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
				return
			}

			// Check if user is active
			if !user.Active {
				http.Error(w, "User account is inactive", http.StatusForbidden)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			
			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PolicyMiddleware checks if the user has the required permissions
func PolicyMiddleware(db *gorm.DB, resource string, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user, ok := r.Context().Value(UserContextKey).(schemas.User)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has SuperAdmin role (bypass policy check)
			var role schemas.Role
			if err := db.First(&role, "id = ?", user.RoleId).Error; err != nil {
				klog.Errorf("Error fetching role: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// SuperAdmin role has access to everything
			if role.Name == "SuperAdmin" {
				next.ServeHTTP(w, r)
				return
			}

			// Check policies for the user's role
			var policies []schemas.Policy
			if err := db.Where("roles_id = ? AND resource = ?", user.RoleId, resource).Find(&policies).Error; err != nil {
				klog.Errorf("Error fetching policies: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Check if any policy allows the action
			allowed := false
			for _, policy := range policies {
				if (policy.Action == "*" || policy.Action == action) && policy.Effect == "allow" {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Permission denied", http.StatusForbidden)
				return
			}

			// User has permission, proceed to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateToken validates the JWT token and returns the user ID
func ValidateToken(tokenString string) (uuid.UUID, error) {
	// This is a placeholder for actual JWT validation
	// In a real implementation, you would:
	// 1. Parse the JWT token
	// 2. Verify the signature
	// 3. Check if the token is expired
	// 4. Extract and return the user ID from the token claims
	
	// For now, we'll return an error to indicate this needs to be implemented
	return uuid.Nil, errors.New("token validation not implemented")
}