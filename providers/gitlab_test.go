package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/oauth2-proxy/oauth2-proxy/pkg/apis/sessions"
)

func testGitLabProvider(hostname string) *GitLabProvider {
	p := NewGitLabProvider(
		&ProviderData{
			ProviderName: "",
			LoginURL:     &url.URL{},
			RedeemURL:    &url.URL{},
			ProfileURL:   &url.URL{},
			ValidateURL:  &url.URL{},
			Scope:        ""})
	if hostname != "" {
		updateURL(p.Data().LoginURL, hostname)
		updateURL(p.Data().RedeemURL, hostname)
		updateURL(p.Data().ProfileURL, hostname)
		updateURL(p.Data().ValidateURL, hostname)
	}

	return p
}

func testGitLabBackend() *httptest.Server {
	userInfo := `
		{
			"nickname": "FooBar",
			"email": "foo@bar.com",
			"email_verified": false,
			"groups": ["foo", "bar"]
		}
	`
	authHeader := "Bearer gitlab_access_token"

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth/userinfo" {
				if r.Header["Authorization"][0] == authHeader {
					w.WriteHeader(200)
					w.Write([]byte(userInfo))
				} else {
					w.WriteHeader(401)
				}
			} else {
				w.WriteHeader(404)
			}
		}))
}

func TestGitLabProviderBadToken(t *testing.T) {
	b := testGitLabBackend()
	defer b.Close()

	bURL, _ := url.Parse(b.URL)
	p := testGitLabProvider(bURL.Host)

	ctx := context.Background()
	session := &sessions.SessionState{AccessToken: "unexpected_gitlab_access_token"}
	_, err := p.GetEmailAddress(ctx, session)
	assert.NotEqual(t, nil, err)
}

func TestGitLabProviderUnverifiedEmailDenied(t *testing.T) {
	b := testGitLabBackend()
	defer b.Close()

	bURL, _ := url.Parse(b.URL)
	p := testGitLabProvider(bURL.Host)

	ctx := context.Background()
	session := &sessions.SessionState{AccessToken: "gitlab_access_token"}
	_, err := p.GetEmailAddress(ctx, session)
	assert.NotEqual(t, nil, err)
}

func TestGitLabProviderUnverifiedEmailAllowed(t *testing.T) {
	b := testGitLabBackend()
	defer b.Close()

	bURL, _ := url.Parse(b.URL)
	p := testGitLabProvider(bURL.Host)
	p.AllowUnverifiedEmail = true

	ctx := context.Background()
	session := &sessions.SessionState{AccessToken: "gitlab_access_token"}
	email, err := p.GetEmailAddress(ctx, session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "foo@bar.com", email)
}

func TestGitLabProviderUsername(t *testing.T) {
	b := testGitLabBackend()
	defer b.Close()

	bURL, _ := url.Parse(b.URL)
	p := testGitLabProvider(bURL.Host)
	p.AllowUnverifiedEmail = true

	ctx := context.Background()
	session := &sessions.SessionState{AccessToken: "gitlab_access_token"}
	username, err := p.GetUserName(ctx, session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "FooBar", username)
}

func TestGitLabProviderGroupMembershipValid(t *testing.T) {
	b := testGitLabBackend()
	defer b.Close()

	bURL, _ := url.Parse(b.URL)
	p := testGitLabProvider(bURL.Host)
	p.AllowUnverifiedEmail = true
	p.Group = "foo"

	ctx := context.Background()
	session := &sessions.SessionState{AccessToken: "gitlab_access_token"}
	email, err := p.GetEmailAddress(ctx, session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "foo@bar.com", email)
}

func TestGitLabProviderGroupMembershipMissing(t *testing.T) {
	b := testGitLabBackend()
	defer b.Close()

	bURL, _ := url.Parse(b.URL)
	p := testGitLabProvider(bURL.Host)
	p.AllowUnverifiedEmail = true
	p.Group = "baz"

	ctx := context.Background()
	session := &sessions.SessionState{AccessToken: "gitlab_access_token"}
	_, err := p.GetEmailAddress(ctx, session)
	assert.NotEqual(t, nil, err)
}

func TestGitLabProviderEmailDomainValid(t *testing.T) {
	b := testGitLabBackend()
	defer b.Close()

	bURL, _ := url.Parse(b.URL)
	p := testGitLabProvider(bURL.Host)
	p.AllowUnverifiedEmail = true
	p.EmailDomains = []string{"bar.com"}

	ctx := context.Background()
	session := &sessions.SessionState{AccessToken: "gitlab_access_token"}
	email, err := p.GetEmailAddress(ctx, session)
	assert.Equal(t, nil, err)
	assert.Equal(t, "foo@bar.com", email)
}

func TestGitLabProviderEmailDomainInvalid(t *testing.T) {
	b := testGitLabBackend()
	defer b.Close()

	bURL, _ := url.Parse(b.URL)
	p := testGitLabProvider(bURL.Host)
	p.AllowUnverifiedEmail = true
	p.EmailDomains = []string{"baz.com"}

	ctx := context.Background()
	session := &sessions.SessionState{AccessToken: "gitlab_access_token"}
	_, err := p.GetEmailAddress(ctx, session)
	assert.NotEqual(t, nil, err)
}
