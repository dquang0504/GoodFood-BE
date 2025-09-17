package dto

import "time"

// UserCards struct represents the summary stats for Admin User module
type UserCards struct {
	TotalUsers    int `boil:"totalusers"`
	TotalDisabled int `boil:"totaldisabled"`
}

// UserError struct represents custom errors for api response in Admin User module
type UserError struct {
	ErrName     string `json:"errName"`
	ErrUsername string `json:"errUsername"`
	ErrPhone    string `json:"errPhone"`
	ErrEmail    string `json:"errEmail"`
	ErrPassword string `json:"errPassword"`
}

type OAuthLoginStruct struct {
	AccessToken string `json:"accessToken"`
}

type ForgotPassStruct struct {
	Email   string `json:"email"`
	CodeOTP string `json:"codeOTP"`
}

type ResetPass struct {
	Token       string `json:"token"`
	NewPass     string `json:"newPass"`
	ConfirmPass string `json:"confirmPass"`
	Email       string `json:"email"`
}

type ContactResponse struct {
	Fullname string `json:"name"`
	Email    string `json:"fromEmail"`
	Message  string `json:"content"`
}

// RefreshRecord struct represents the refresh token caching structure.
type RefreshRecord struct {
	Username  string    `json:"username"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	UserAgent string    `json:"user_agent,omitempty"`
	IP        string    `json:"ip,omitempty"`
}
