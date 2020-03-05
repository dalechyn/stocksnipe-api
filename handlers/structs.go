package handlers

type registerRequest struct {
	Email    string `json:"email" validate:"required"`
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"min=8,max=32"`
}

type loginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"min=8,max=32"`
}

type loginResponse struct {
	Login         string `json:"login"`
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken"`
}

type tokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type tokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

