package eureka

import (
	"golang.org/x/net/context"

	"github.com/micro/go-micro/registry"
)

type contextOauth2Credentials struct{}

type oauth2Credentials struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
}

// Enable OAuth 2.0 Client Credentials Grant Flow
func OAuth2ClientCredentials(clientID, clientSecret, tokenURL string) registry.Option {
	return func(o *registry.Options) {
		o.Context = context.WithValue(o.Context, contextOauth2Credentials{}, oauth2Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     tokenURL,
		})
	}
}
