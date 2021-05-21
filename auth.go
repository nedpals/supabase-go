package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-querystring/query"
)

type authError struct {
	Message string `json:"message"`
}

type auth struct {
	baseEndpoint string
	client       *Client
}

type UserCredentials struct {
	email    string
	password string
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
func (a *auth) SignUp(ctx context.Context, credentials UserCredentials) (*User, error) {
	reqBody, _ := json.Marshal(credentials)
	reqURL := fmt.Sprintf("%s/%s/signup", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(reqBody))
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
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type authenticationError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// SignIn enters the user credentials and returns the current user if succeeded.
func (a *auth) SignIn(ctx context.Context, credentials UserCredentials) (*AuthenticatedDetails, error) {
	reqBody, _ := json.Marshal(credentials)
	reqURL := fmt.Sprintf("%s/%s/token?grant_type=password", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(reqBody))
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
func (a *auth) RefreshUser(ctx context.Context, userToken string, refreshToken string) (*AuthenticatedDetails, error) {
	reqBody, _ := json.Marshal(map[string]string{"refresh_token": refreshToken})
	reqURL := fmt.Sprintf("%s/%s/token?grant_type=refresh_token", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(reqBody))
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

// SendMagicLink sends a link to a specific e-mail address for passwordless auth.
func (a *auth) SendMagicLink(ctx context.Context, email string) error {
	reqBody, _ := json.Marshal(map[string]string{"email": email})
	reqURL := fmt.Sprintf("%s/%s/magiclink", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, bytes.NewBuffer(reqBody))
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
}

type ProviderSignInDetails struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

// SignInWithProvider returns a URL for signing in via OAuth
func (a *auth) SignInWithProvider(opts ProviderSignInOptions) (*ProviderSignInDetails, error) {
	params, err := query.Values(opts)
	if err != nil {
		return nil, err
	}

	details := ProviderSignInDetails{
		URL:      fmt.Sprintf("%s/%s/authorize?%s", a.client.BaseURL, a.baseEndpoint, params.Encode()),
		Provider: opts.Provider,
	}
	return &details, nil
}

// User retrieves the user information based on the given token
func (a *auth) User(ctx context.Context, userToken string) (*User, error) {
	reqURL := fmt.Sprintf("%s/%s/user", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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
func (a *auth) UpdateUser(ctx context.Context, userToken string, updateData map[string]interface{}) (*User, error) {
	reqBody, _ := json.Marshal(updateData)
	reqURL := fmt.Sprintf("%s/%s/user", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "PUT", reqURL, bytes.NewBuffer(reqBody))
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
func (a *auth) ResetPasswordForEmail(ctx context.Context, email string) error {
	reqBody, _ := json.Marshal(map[string]string{"email": email})
	reqURL := fmt.Sprintf("%s/%s/recover", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	if err = a.client.sendRequest(req, nil); err != nil {
		return err
	}

	return nil
}

// SignOut revokes the users token and session.
func (a *auth) SignOut(ctx context.Context, userToken string) error {
	reqURL := fmt.Sprintf("%s/%s/logout", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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
func (a *auth) InviteUserByEmail(ctx context.Context, email string) (*User, error) {
	reqBody, _ := json.Marshal(map[string]string{"email": email})
	reqURL := fmt.Sprintf("%s/%s/invite", a.client.BaseURL, a.baseEndpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(reqBody))
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
