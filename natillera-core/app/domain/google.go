package domain

type GoogleQuery struct {
	State    string `url:"state"`
	Iss      string `url:"iss"`
	Code     string `url:"code"`
	Scope    string `url:"scope"`
	AuthUser string `url:"authuser"`
	Prompt   string `url:"prompt"`
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type GoogleTokenRequest struct {
	Token string `json:"token" validate:"required"`
}
