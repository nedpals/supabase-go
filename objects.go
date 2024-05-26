package supabase

type VerificationType string

const (
	VerificationTypeSignup      = "signup"
	VerificationTypeRecovery    = "recovery"
	VerificationTypeInvite      = "invite"
	VerificationTypeMagiclink   = "magiclink"
	VerificationTypeEmailChange = "email_change"
	VerificationTypeSMS         = "sms"
	VerificationTypePhoneChange = "phone_change"
)

type VerifyResponse struct {
	URL string

	// The fields below are returned only for a successful response.
	AccessToken  string
	TokenType    string
	ExpiresIn    int
	RefreshToken string
	Type         VerificationType

	// The fields below are returned if there was an error verifying.
	Error            string
	ErrorCode        string
	ErrorDescription string
}

type VerifyRequest struct {
	Type       VerificationType
	Token      string
	RedirectTo string
}

type OTPRequest struct {
	Email      string                 `json:"email"`
	Phone      string                 `json:"phone"`
	CreateUser bool                   `json:"create_user"`
	Data       map[string]interface{} `json:"data"`
}

type OTPResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	User         User   `json:"user"`
}
