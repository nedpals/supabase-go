package supabase

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
)

type authError struct {
	Message string `json:"message"`
}

type Auth struct {
	client *Client
}

type UserCredentials struct {
	Email    string
	Password string
	Data     interface{}
}

type User struct {
	ID                 string                    `json:"id"`
	Aud                string                    `json:"aud"`
	Role               string                    `json:"role"`
	Email              string                    `json:"email"`
	InvitedAt          time.Time                 `json:"invited_at"`
	ConfirmedAt        time.Time                 `json:"confirmed_at"`
	ConfirmationSentAt time.Time                 `json:"confirmation_sent_at"`
	AppMetadata        struct{ provider string } `json:"app_metadata"`
	UserMetadata       map[string]interface{}    `json:"user_metadata"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}

// SignUp registers the user's email and password to the database.
func (a *Auth) SignUp(ctx context.Context, credentials UserCredentials) (*User, error) {
	reqBody, _ := json.Marshal(credentials)
	reqURL := fmt.Sprintf("%s/%s/signup", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	res := User{}
	if err := a.client.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

type AuthenticatedDetails struct {
	AccessToken          string `json:"access_token"`
	TokenType            string `json:"token_type"`
	ExpiresIn            int    `json:"expires_in"`
	RefreshToken         string `json:"refresh_token"`
	User                 User   `json:"user"`
	ProviderToken        string `json:"provider_token"`
	ProviderRefreshToken string `json:"provider_refresh_token"`
}

type authenticationError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type exchangeError struct {
	Message string `json:"msg"`
}

// SignIn enters the user credentials and returns the current user if succeeded.
func (a *Auth) SignIn(ctx context.Context, credentials UserCredentials) (*AuthenticatedDetails, error) {
	reqBody, _ := json.Marshal(credentials)
	reqURL := fmt.Sprintf("%s/%s/token?grant_type=password", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	res := AuthenticatedDetails{}
	errRes := authenticationError{}
	hasCustomError, err := a.client.sendCustomRequest(req, &res, &errRes)
	if err != nil {
		return nil, err
	} else if hasCustomError {
		return nil, errors.New(fmt.Sprintf("%s: %s", errRes.Error, errRes.ErrorDescription))
	}

	return &res, nil
}

// SignIn enters the user credentials and returns the current user if succeeded.
func (a *Auth) RefreshUser(ctx context.Context, userToken string, refreshToken string) (*AuthenticatedDetails, error) {
	reqBody, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	reqURL := fmt.Sprintf("%s/%s/token?grant_type=refresh_token", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	injectAuthorizationHeader(req, userToken)
	req.Header.Set("Content-Type", "application/json")
	res := AuthenticatedDetails{}
	errRes := authenticationError{}
	hasCustomError, err := a.client.sendCustomRequest(req, &res, &errRes)
	if err != nil {
		return nil, err
	} else if hasCustomError {
		return nil, errors.New(fmt.Sprintf("%s: %s", errRes.Error, errRes.ErrorDescription))
	}

	return &res, nil
}

type ExchangeCodeOpts struct {
	AuthCode     string `json:"auth_code"`
	CodeVerifier string `json:"code_verifier"`
}

// ExchangeCode takes an auth code and PCKE verifier and returns the current user if succeeded.
func (a *Auth) ExchangeCode(ctx context.Context, opts ExchangeCodeOpts) (*AuthenticatedDetails, error) {
	reqBody, _ := json.Marshal(opts)
	reqURL := fmt.Sprintf("%s/%s/token?grant_type=pkce", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	res := AuthenticatedDetails{}
	errRes := exchangeError{}
	hasCustomError, err := a.client.sendCustomRequest(req, &res, &errRes)
	if err != nil {
		return nil, err
	} else if hasCustomError {
		return nil, errors.New(errRes.Message)
	}

	return &res, err
}

// SendMagicLink sends a link to a specific e-mail address for passwordless auth.
func (a *Auth) SendMagicLink(ctx context.Context, email string) error {
	reqBody, _ := json.Marshal(map[string]string{"email": email})
	reqURL := fmt.Sprintf("%s/%s/magiclink", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	errRes := authError{}
	hasCustomError, err := a.client.sendCustomRequest(req, nil, &errRes)
	if err != nil {
		return err
	} else if hasCustomError {
		return errors.New(fmt.Sprintf("%s", errRes.Message))
	}

	return nil
}

type ProviderSignInOptions struct {
	Provider   string   `url:"provider"`
	RedirectTo string   `url:"redirect_to"`
	Scopes     []string `url:"scopes"`
	FlowType   FlowType
}

type FlowType string

const (
	Implicit FlowType = "implicit"
	PKCE     FlowType = "pkce"
)

type ProviderSignInDetails struct {
	URL          string `json:"url"`
	Provider     string `json:"provider"`
	CodeVerifier string `json:"code_verifier"`
}

// SignInWithProvider returns a URL for signing in via OAuth
func (a *Auth) SignInWithProvider(opts ProviderSignInOptions) (*ProviderSignInDetails, error) {
	params, err := query.Values(opts)
	if err != nil {
		return nil, err
	}

	params.Set("scopes", strings.Join(opts.Scopes, " "))

	if opts.FlowType == PKCE {
		p, err := generatePKCEParams()
		if err != nil {
			return nil, err
		}

		params.Add("code_challenge", p.Challenge)
		params.Add("code_challenge_method", p.ChallengeMethod)

		details := ProviderSignInDetails{
			URL:          fmt.Sprintf("%s/%s/authorize?%s", a.client.BaseURL, AuthEndpoint, params.Encode()),
			Provider:     opts.Provider,
			CodeVerifier: p.Verifier,
		}

		return &details, nil
	}

	// Implicit flow
	details := ProviderSignInDetails{
		URL:      fmt.Sprintf("%s/%s/authorize?%s", a.client.BaseURL, AuthEndpoint, params.Encode()),
		Provider: opts.Provider,
	}

	return &details, nil
}

// User retrieves the user information based on the given token
func (a *Auth) User(ctx context.Context, userToken string) (*User, error) {
	reqURL := fmt.Sprintf("%s/%s/user", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	injectAuthorizationHeader(req, userToken)
	res := User{}
	errRes := authError{}
	hasCustomError, err := a.client.sendCustomRequest(req, &res, &errRes)
	if err != nil {
		return nil, err
	} else if hasCustomError {
		return nil, errors.New(fmt.Sprintf("%s", errRes.Message))
	}

	return &res, nil
}

// UpdateUser updates the user information
func (a *Auth) UpdateUser(ctx context.Context, userToken string, updateData map[string]interface{}) (*User, error) {
	reqBody, _ := json.Marshal(updateData)
	reqURL := fmt.Sprintf("%s/%s/user", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	injectAuthorizationHeader(req, userToken)

	res := User{}
	errRes := authError{}
	hasCustomError, err := a.client.sendCustomRequest(req, &res, &errRes)
	if err != nil {
		return nil, err
	} else if hasCustomError {
		return nil, errors.New(fmt.Sprintf("%s", errRes.Message))
	}

	return &res, nil
}

// ResetPasswordForEmail sends a password recovery link to the given e-mail address.
func (a *Auth) ResetPasswordForEmail(ctx context.Context, email string) error {
	reqBody, _ := json.Marshal(map[string]string{"email": email})
	reqURL := fmt.Sprintf("%s/%s/recover", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	if err = a.client.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}

// SignOut revokes the users token and session.
func (a *Auth) SignOut(ctx context.Context, userToken string) error {
	reqURL := fmt.Sprintf("%s/%s/logout", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return err
	}

	injectAuthorizationHeader(req, userToken)
	req.Header.Set("Content-Type", "application/json")
	if err = a.client.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}

// InviteUserByEmail sends an invite link to the given email. Returns a user.
func (a *Auth) InviteUserByEmail(ctx context.Context, email string) (*User, error) {
	reqBody, _ := json.Marshal(map[string]string{"email": email})
	reqURL := fmt.Sprintf("%s/%s/invite", a.client.BaseURL, AuthEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	injectAuthorizationHeader(req, a.client.apiKey)
	req.Header.Set("Content-Type", "application/json")
	res := User{}
	if err := a.client.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// adapted from https://go-review.googlesource.com/c/oauth2/+/463979/9/pkce.go#64
type PKCEParams struct {
	Challenge       string
	ChallengeMethod string
	Verifier        string
}

func generatePKCEParams() (*PKCEParams, error) {
	data := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return nil, err
	}

	// RawURLEncoding since "code challenge can only contain alphanumeric characters, hyphens, periods, underscores and tildes"
	verifier := base64.RawURLEncoding.EncodeToString(data)
	sha := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sha[:])
	return &PKCEParams{
		Challenge:       challenge,
		ChallengeMethod: "S256",
		Verifier:        verifier,
	}, nil
}
