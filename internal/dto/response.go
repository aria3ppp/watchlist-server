package dto

type UserRefreshResponse struct {
	JwtToken     string `json:"jwt_token"`
	JwtExpiresAt int64  `json:"jwt_expires_at"`
}

type UserLoginResponse struct {
	UserRefreshResponse
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
	UserID           int    `json:"user_id"`
}
