package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yash3004/user_management_service/internal/schemas"
	allManager "github.com/yash3004/user_management_service"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// UserKey is the key for the user in the context
	UserKey ContextKey = "user"

	// AuthorizationHeader is the header name for the authorization token
	AuthorizationHeader = "Authorization"

	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
)

// Middleware provides authentication and authorization middleware
type Middleware struct {
	sessionManager *SessionManager
	jwtSecret      []byte
	policyManager  allManager.PoliciesManager
}

// NewMiddleware creates a new auth middleware
func NewMiddleware(sessionManager *SessionManager, jwtSecret []byte, policyManager allManager.PoliciesManager) *Middleware {
	return &Middleware{
		sessionManager: sessionManager,
		jwtSecret:      jwtSecret,
		policyManager:  policyManager,
	}
}

// RequireAuthentication ensures the user is authenticated
func (m *Middleware) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try to get user from session first (for web flows)
		user, err := m.sessionManager.GetCurrentUser(ctx, r)
		if err == nil {
			// User found in session
			ctx = context.WithValue(ctx, UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Try JWT token (for API flows)
		authHeader := r.Header.Get(AuthorizationHeader)
		if authHeader != "" && strings.HasPrefix(authHeader, BearerPrefix) {
			tokenString := strings.TrimPrefix(authHeader, BearerPrefix)

			user, err = m.validateJWT(ctx, tokenString)
			if err == nil {
				// User found from JWT
				ctx = context.WithValue(ctx, UserKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Not authenticated
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// RequirePolicy ensures the user has the required policy permission
func (m *Middleware) RequirePolicy(policyName string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, ok := ctx.Value(UserKey).(*schemas.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user has the required policy
		hasPolicy, err := m.policyManager.UserHasPolicy(ctx, user.ID, policyName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !hasPolicy {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireProjectAccess ensures the user has access to a project
func (m *Middleware) RequireProjectAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		user, ok := ctx.Value(UserKey).(*schemas.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get project ID from URL parameters or request body
		// This is a simplified example - adapt to your routing library
		projectIDStr := r.PathValue("projectID") // Go 1.22+ feature
		if projectIDStr == "" {
			http.Error(w, "Project ID required", http.StatusBadRequest)
			return
		}

		projectID, err := uuid.Parse(projectIDStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// Check if user has access to the project
		hasAccess, err := m.policyManager.UserHasProjectAccess(ctx, user.ID, projectID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !hasAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetCurrentUser gets the current user from the context
func (m *Middleware) GetCurrentUser(ctx context.Context) (*schemas.User, error) {
	user, ok := ctx.Value(UserKey).(*schemas.User)
	if !ok {
		return nil, errors.New("no user in context")
	}
	return user, nil
}

// validateJWT validates a JWT token and returns the user
func (m *Middleware) validateJWT(ctx context.Context, tokenString string) (*schemas.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid subject claim")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	userStore, ok := m.sessionManager.userStore.(UserStore)
	if !ok {
		return nil, errors.New("invalid user store")
	}

	user, err := userStore.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}
