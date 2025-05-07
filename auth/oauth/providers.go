package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/github"
)

// Provider represents an OAuth provider
type Provider interface {
	GetAuthURL(state string) string
	
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
	
	GetName() string
}

type UserInfo struct {
	ID        string
	Email     string
	Name      string
	FirstName string
	LastName  string
	Picture   string
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type ProviderFactory struct {
	providers map[string]Provider
}

func NewProviderFactory(configs map[string]ProviderConfig) *ProviderFactory {
	factory := &ProviderFactory{
		providers: make(map[string]Provider),
	}
	
	for name, config := range configs {
		switch name {
		case "google":
			factory.providers[name] = NewGoogleProvider(config)
		case "github":
			factory.providers[name] = NewGithubProvider(config)
		// Add more providers as needed
		}
	}
	
	return factory
}

// GetProvider returns a provider by name
func (f *ProviderFactory) GetProvider(name string) (Provider, error) {
	provider, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetAllProviders returns all configured providers
func (f *ProviderFactory) GetAllProviders() map[string]Provider {
	return f.providers
}