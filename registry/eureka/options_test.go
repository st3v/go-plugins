package eureka

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/micro/go-micro/registry"
)

func TestOAuth2ClientCredentials(t *testing.T) {
	clientID := "client-id"
	clientSecret := "client-secret"
	tokenURL := "token-url"

	options := new(registry.Options)
	options.Context = context.WithValue(context.Background(), "foo", "bar")

	OAuth2ClientCredentials(clientID, clientSecret, tokenURL)(options)

	creds, ok := options.Context.Value(contextOauth2Credentials{}).(oauth2Credentials)
	if !ok {
		t.Errorf("oauth2Credentials missing in options.Context")
	}

	tests := []struct {
		subject string
		want    string
		have    string
	}{
		{"ClientID", clientID, creds.ClientID},
		{"ClientSecret", clientSecret, creds.ClientSecret},
		{"TokenURL", tokenURL, creds.TokenURL},
		{"OriginalContext", "bar", options.Context.Value("foo").(string)},
	}

	for _, tc := range tests {
		if tc.want != tc.have {
			t.Errorf("%s: want %q, got %q", tc.subject, tc.want, tc.have)
		}
	}
}
