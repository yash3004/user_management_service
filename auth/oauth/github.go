package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GithubProvider implements the Provider interface for GitHub OAuth
type GithubProvider struct {
	config *oauth2.Config
}

// NewGithubProvider creates a new GitHub OAuth provider
func NewGithubProvider(config ProviderConfig) *GithubProvider {
	return &GithubProvider{
		config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			RedirectURL:  config.RedirectURL,
			Scopes:       config.Scopes,
			Endpoint:     github.Endpoint,
		},
	}
}

func (p *GithubProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (p *GithubProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

func (p *GithubProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(ctx, token)
	
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	
	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %v", err)
	}
	
	var email string
	if githubUser.Email == "" {
		email, err = p.getEmail(client)
		if err != nil {
			return nil, err
		}
	} else {
		email = githubUser.Email
	}
	
	firstName, lastName := splitName(githubUser.Name)
	
	return &UserInfo{
		ID:        fmt.Sprintf("%d", githubUser.ID),
		Email:     email,
		Name:      githubUser.Name,
		FirstName: firstName,
		LastName:  lastName,
		Picture:   githubUser.AvatarURL,
	}, nil
}

func (p *GithubProvider) getEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", fmt.Errorf("failed to get user emails: %v", err)
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read emails response body: %v", err)
	}
	
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", fmt.Errorf("failed to parse emails: %v", err)
	}
	
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	
	return "", fmt.Errorf("no verified email found")
}

func (p *GithubProvider) GetName() string {
	return "github"
}

func splitName(fullName string) (string, string) {
	if fullName == "" {
		return "", ""
	}
	
	parts := strings.Fields(fullName)
	if len(parts) == 1 {
		return parts[0], ""
	}
	
	return parts[0], strings.Join(parts[1:], " ")
}