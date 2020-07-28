package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/oauth2-proxy/oauth2-proxy/pkg/apis/sessions"
	"github.com/oauth2-proxy/oauth2-proxy/pkg/requests"
)

// FacebookProvider represents an Facebook based Identity Provider
type FacebookProvider struct {
	*ProviderData
}

var _ Provider = (*FacebookProvider)(nil)

const (
	facebookProviderName = "Facebook"
	facebookDefaultScope = "public_profile email"
)

var (
	// Default Login URL for Facebook.
	// Pre-parsed URL of https://www.facebook.com/v2.5/dialog/oauth.
	facebookDefaultLoginURL = &url.URL{
		Scheme: "https",
		Host:   "www.facebook.com",
		Path:   "/v2.5/dialog/oauth",
		// ?granted_scopes=true
	}

	// Default Redeem URL for Facebook.
	// Pre-parsed URL of https://graph.facebook.com/v2.5/oauth/access_token.
	facebookDefaultRedeemURL = &url.URL{
		Scheme: "https",
		Host:   "graph.facebook.com",
		Path:   "/v2.5/oauth/access_token",
	}

	// Default Profile URL for Facebook.
	// Pre-parsed URL of https://graph.facebook.com/v2.5/me.
	facebookDefaultProfileURL = &url.URL{
		Scheme: "https",
		Host:   "graph.facebook.com",
		Path:   "/v2.5/me",
	}
)

// NewFacebookProvider initiates a new FacebookProvider
func NewFacebookProvider(p *ProviderData) *FacebookProvider {
	p.setProviderDefaults(providerDefaults{
		name:        facebookProviderName,
		loginURL:    facebookDefaultLoginURL,
		redeemURL:   facebookDefaultRedeemURL,
		profileURL:  facebookDefaultProfileURL,
		validateURL: facebookDefaultProfileURL,
		scope:       facebookDefaultScope,
	})
	return &FacebookProvider{ProviderData: p}
}

func getFacebookHeader(accessToken string) http.Header {
	header := make(http.Header)
	header.Set("Accept", "application/json")
	header.Set("x-li-format", "json")
	header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	return header
}

// GetEmailAddress returns the Account email address
func (p *FacebookProvider) GetEmailAddress(ctx context.Context, s *sessions.SessionState) (string, error) {
	if s.AccessToken == "" {
		return "", errors.New("missing access token")
	}

	type result struct {
		Email string
	}
	var r result

	requestURL := p.ProfileURL.String() + "?fields=name,email"
	err := requests.New(requestURL).
		WithContext(ctx).
		WithHeaders(getFacebookHeader(s.AccessToken)).
		Do().
		UnmarshalInto(&r)
	if err != nil {
		return "", err
	}

	if r.Email == "" {
		return "", errors.New("no email")
	}
	return r.Email, nil
}

// ValidateSessionState validates the AccessToken
func (p *FacebookProvider) ValidateSessionState(ctx context.Context, s *sessions.SessionState) bool {
	return validateToken(ctx, p, s.AccessToken, getFacebookHeader(s.AccessToken))
}
