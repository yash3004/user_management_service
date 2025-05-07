package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleProvider implements the Provider interface for Google OAuth
type GoogleProvider struct {
	config *oauth2.Config
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider(config ProviderConfig) *GoogleProvider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       config.Scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

// GetAuthURL returns the URL to redirect the user to for authentication
func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange exchanges the auth code for tokens
func (p *GoogleProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// GetUserInfo gets user info from Google using the token
func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	var googleUser struct {
		Sub      string `json:"sub"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		GivenName string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture  string `json:"picture"`
	}
	
	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %v", err)
	}
	
	return &UserInfo{
		ID:        googleUser.Sub,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		FirstName: googleUser.GivenName,
		LastName:  googleUser.FamilyName,
		Picture:   googleUser.Picture,
	}, nil
}

// GetName returns the name of the provider
func (p *GoogleProvider) GetName() string {
	return "google"
}