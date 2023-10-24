package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type JSONMap map[string]interface{}

type Admin struct {
	client     *Client
	serviceKey string
}

type Factor struct {
	ID           string    `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Status       string    `json:"status" db:"status"`
	FriendlyName string    `json:"friendly_name,omitempty" db:"friendly_name"`
	FactorType   string    `json:"factor_type" db:"factor_type"`
}

type Identity struct {
	ID           string     `json:"id" db:"id"`
	UserID       string     `json:"user_id" db:"user_id"`
	IdentityData JSONMap    `json:"identity_data,omitempty" db:"identity_data"`
	Provider     string     `json:"provider" db:"provider"`
	LastSignInAt *time.Time `json:"last_sign_in_at,omitempty" db:"last_sign_in_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

type AdminUser struct {
	ID string `json:"id" db:"id"`

	Aud   string `json:"aud" db:"aud"`
	Role  string `json:"role" db:"role"`
	Email string `json:"email" db:"email"`

	EmailConfirmedAt *time.Time `json:"email_confirmed_at,omitempty" db:"email_confirmed_at"`
	InvitedAt        *time.Time `json:"invited_at,omitempty" db:"invited_at"`

	Phone            string     `json:"phone" db:"phone"`
	PhoneConfirmedAt *time.Time `json:"phone_confirmed_at,omitempty" db:"phone_confirmed_at"`

	ConfirmationSentAt *time.Time `json:"confirmation_sent_at,omitempty" db:"confirmation_sent_at"`

	RecoverySentAt *time.Time `json:"recovery_sent_at,omitempty" db:"recovery_sent_at"`

	EmailChange       string     `json:"new_email,omitempty" db:"email_change"`
	EmailChangeSentAt *time.Time `json:"email_change_sent_at,omitempty" db:"email_change_sent_at"`

	PhoneChange       string     `json:"new_phone,omitempty" db:"phone_change"`
	PhoneChangeSentAt *time.Time `json:"phone_change_sent_at,omitempty" db:"phone_change_sent_at"`

	ReauthenticationSentAt *time.Time `json:"reauthentication_sent_at,omitempty" db:"reauthentication_sent_at"`

	LastSignInAt *time.Time `json:"last_sign_in_at,omitempty" db:"last_sign_in_at"`

	AppMetaData  JSONMap `json:"app_metadata" db:"raw_app_meta_data"`
	UserMetaData JSONMap `json:"user_metadata" db:"raw_user_meta_data"`

	Factors    []Factor   `json:"factors,omitempty" has_many:"factors"`
	Identities []Identity `json:"identities" has_many:"identities"`

	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	BannedUntil *time.Time `json:"banned_until,omitempty" db:"banned_until"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type AdminUserParams struct {
	Role         string  `json:"role"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	Password     *string `json:"password"`
	EmailConfirm bool    `json:"email_confirm"`
	PhoneConfirm bool    `json:"phone_confirm"`
	UserMetadata JSONMap `json:"user_metadata"`
	AppMetadata  JSONMap `json:"app_metadata"`
	BanDuration  string  `json:"ban_duration"`
}

type GenerateLinkParams struct {
	Type       string                 `json:"type"`
	Email      string                 `json:"email"`
	NewEmail   string                 `json:"new_email"`
	Password   string                 `json:"password"`
	Data       map[string]interface{} `json:"data"`
	RedirectTo string                 `json:"redirect_to"`
}

type GenerateLinkResponse struct {
	AdminUser
	ActionLink       string `json:"action_link"`
	EmailOtp         string `json:"email_otp"`
	HashedToken      string `json:"hashed_token"`
	VerificationType string `json:"verification_type"`
	RedirectTo       string `json:"redirect_to"`
}

// Retrieve the user
func (a *Admin) GetUser(ctx context.Context, userID string) (*AdminUser, error) {
	reqURL := fmt.Sprintf("%s/%s/users/%s", a.client.BaseURL, AdminEndpoint, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.serviceKey))
	res := AdminUser{}
	if err := a.client.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Create a user
func (a *Admin) CreateUser(ctx context.Context, params AdminUserParams) (*AdminUser, error) {
	reqBody, _ := json.Marshal(params)
	reqURL := fmt.Sprintf("%s/%s/users", a.client.BaseURL, AdminEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	injectAuthorizationHeader(req, a.serviceKey)
	res := AdminUser{}
	if err := a.client.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update a user
func (a *Admin) UpdateUser(ctx context.Context, userID string, params AdminUserParams) (*AdminUser, error) {
	reqBody, _ := json.Marshal(params)
	reqURL := fmt.Sprintf("%s/%s/users/%s", a.client.BaseURL, AdminEndpoint, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	injectAuthorizationHeader(req, a.serviceKey)
	res := AdminUser{}
	if err := a.client.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Update a user
func (a *Admin) GenerateLink(ctx context.Context, params GenerateLinkParams) (*GenerateLinkResponse, error) {
	reqBody, _ := json.Marshal(params)
	reqURL := fmt.Sprintf("%s/%s/generate_link", a.client.BaseURL, AdminEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	injectAuthorizationHeader(req, a.serviceKey)
	res := GenerateLinkResponse{}
	if err := a.client.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
