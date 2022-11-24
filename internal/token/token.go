package token

//go:generate mockgen -destination mock_token/mock_service.go . Service

type Service interface {
	GenerateAccessToken(*Payload) (string, error)
	GenerateRefreshToken(*Payload) (string, error)
	ValidateToken(tokenString string) (*Payload, error)
}

type Payload struct {
	UserID int
	// more fields?
}
