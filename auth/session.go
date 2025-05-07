package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	
	"github.com/yash3004/user_management_service/internal/schemas"
)

const (
	SessionName = "auth-session"
	
	UserIDKey = "user_id"
	
	OAuthTokenKey = "oauth_token"
	
	OAuthStateKey = "oauth_state"
)

type SessionManager struct {
	store     sessions.Store
	userStore UserStore
}

type UserStore interface {
	FindByID(ctx context.Context, id uuid.UUID) (*schemas.User, error)
	FindByEmail(ctx context.Context, email string) (*schemas.User, error)
	FindByOAuth(ctx context.Context, provider, oauthID string) (*schemas.User, error)
	Create(ctx context.Context, user *schemas.User) error
	Update(ctx context.Context, user *schemas.User) error
}

func NewSessionManager(secret []byte, userStore UserStore) *SessionManager {
	return &SessionManager{
		store:     sessions.NewCookieStore(secret),
		userStore: userStore,
	}
}

func (sm *SessionManager) GetSession(r *http.Request) (*sessions.Session, error) {
	return sm.store.Get(r, SessionName)
}

func (sm *SessionManager) Login(ctx context.Context, w http.ResponseWriter, r *http.Request, user *schemas.User) error {
	session, err := sm.GetSession(r)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	session.Values[UserIDKey] = user.ID.String()
	
	return session.Save(r, w)
}

// Logout logs a user out
func (sm *SessionManager) Logout(w http.ResponseWriter, r *http.Request) error {
	session, err := sm.GetSession(r)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
	
	return session.Save(r, w)
}

// GetCurrentUser gets the current logged-in user
func (sm *SessionManager) GetCurrentUser(ctx context.Context, r *http.Request) (*schemas.User, error) {
	session, err := sm.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	userIDStr, ok := session.Values[UserIDKey].(string)
	if !ok {
		return nil, errors.New("user not logged in")
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	
	user, err := sm.userStore.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	
	return user, nil
}

// SetOAuthState stores the OAuth state in the session
func (sm *SessionManager) SetOAuthState(w http.ResponseWriter, r *http.Request, state string) error {
	session, err := sm.GetSession(r)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	session.Values[OAuthStateKey] = state
	
	return session.Save(r, w)
}

// VerifyOAuthState verifies the OAuth state from the session
func (sm *SessionManager) VerifyOAuthState(r *http.Request, state string) (bool, error) {
	session, err := sm.GetSession(r)
	if err != nil {
		return false, fmt.Errorf("failed to get session: %w", err)
	}
	
	sessionState, ok := session.Values[OAuthStateKey].(string)
	if !ok {
		return false, errors.New("no state in session")
	}
	
	// Clear the state from the session
	delete(session.Values, OAuthStateKey)
	
	return sessionState == state, nil
}

// StoreOAuthToken stores the OAuth token in the session
func (sm *SessionManager) StoreOAuthToken(w http.ResponseWriter, r *http.Request, token *oauth2.Token) error {
	session, err := sm.GetSession(r)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	
	// Serialize token to JSON
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to serialize token: %w", err)
	}
	
	session.Values[OAuthTokenKey] = string(tokenBytes)
	
	return session.Save(r, w)
}

// GetOAuthToken gets the OAuth token from the session
func (sm *SessionManager) GetOAuthToken(r *http.Request) (*oauth2.Token, error) {
	session, err := sm.GetSession(r)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	tokenStr, ok := session.Values[OAuthTokenKey].(string)
	if !ok {
		return nil, errors.New("no token in session")
	}
	
	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokenStr), &token); err != nil {
		return nil, fmt.Errorf("failed to deserialize token: %w", err)
	}
	
	return &token, nil
}