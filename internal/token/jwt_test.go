package token

import (
	"errors"
	"testing"
	"time"

	"github.com/aria3ppp/watchlist-server/internal/models"
	"github.com/aria3ppp/watchlist-server/internal/token/mock_jwtinner"
	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type JWTTokenServiceSuite struct {
	suite.Suite
	mockController *gomock.Controller
}

func (suite *JWTTokenServiceSuite) SetupTest() {
	suite.mockController = gomock.NewController(suite.T())
}

func (suite *JWTTokenServiceSuite) TearDownTest() {
	suite.mockController.Finish()
}

func TestJWTTokenServiceSuite(t *testing.T) {
	suite.Run(t, &JWTTokenServiceSuite{})
}

func (suite *JWTTokenServiceSuite) TestOK() {
	expUser := &models.User{ID: 1}
	expPayload := &Payload{UserID: expUser.ID}
	expKey := []byte("expected_key")
	expSigningMethod := jwt.SigningMethodHS256
	expAccessDuration := time.Hour
	expRefreshDuration := 30 * 24 * time.Hour

	expAccessClaims := jwtClaims{
		Payload: expPayload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expAccessDuration)),
		},
	}
	expAccessToken := jwt.NewWithClaims(
		expSigningMethod,
		expAccessClaims,
	)
	expAccessTokenString, err := expAccessToken.SignedString(
		expKey,
	)
	suite.Require().NoError(err)

	expRefreshClaims := jwtClaims{
		Payload: &Payload{UserID: expUser.ID},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expRefreshDuration)),
		},
	}
	expRefreshToken := jwt.NewWithClaims(
		expSigningMethod,
		expRefreshClaims,
	)
	expRefreshTokenString, err := expRefreshToken.SignedString(
		expKey,
	)
	suite.Require().NoError(err)

	mockImpl := mock_jwtinner.NewMockJWTInner(suite.mockController)

	signAcessTokenCall := mockImpl.EXPECT().
		SignedString(expAccessToken, expKey).
		Return(expAccessTokenString, nil).
		Times(1)

	signRefreshTokenCall := mockImpl.EXPECT().
		SignedString(expRefreshToken, expKey).
		Return(expRefreshTokenString, nil).
		Times(1).
		After(signAcessTokenCall)

	var expEmptyAccessClaims jwtClaims

	parseAccessTokenCall := mockImpl.EXPECT().
		ParseWithClaims(expAccessTokenString, &expEmptyAccessClaims, gomock.Any()).
		Do(func(_ string, claims jwt.Claims, _ jwt.Keyfunc, _ ...jwt.ParserOption) {
			expAccessToken.Valid = true
			claims.(*jwtClaims).Payload = expPayload
		}).
		SetArg(1, expAccessClaims).
		Return(expAccessToken, nil).
		Times(1).
		After(signRefreshTokenCall)

	var expEmptyRefreshClaims jwtClaims

	mockImpl.EXPECT().
		ParseWithClaims(expRefreshTokenString, &expEmptyRefreshClaims, gomock.Any()).
		Do(func(_ string, claims jwt.Claims, _ jwt.Keyfunc, _ ...jwt.ParserOption) {
			expRefreshToken.Valid = true
			claims.(*jwtClaims).Payload = expPayload
		}).
		Return(expRefreshToken, nil).
		Times(1).
		After(parseAccessTokenCall)

	tokenService := NewJWT(
		JWTConfig{
			Key:             expKey,
			SigningMethod:   expSigningMethod,
			AccessDuration:  expAccessDuration,
			RefreshDuration: expRefreshDuration,
		},
		mockImpl,
	)

	// Invoke GenerateAccessToken method on service with mocked jwt
	accessToken, err := tokenService.GenerateAccessToken(expPayload)

	suite.Require().NoError(err)
	suite.Require().Equal(expAccessTokenString, accessToken)

	// Invoke GenerateRefreshToken method on service with mocked jwt
	refreshToken, err := tokenService.GenerateRefreshToken(expPayload)

	suite.Require().NoError(err)
	suite.Require().Equal(expRefreshTokenString, refreshToken)

	// Invoke ValidateToken method on service with mocked jwt
	payload, err := tokenService.ValidateToken(accessToken)

	suite.Require().NoError(err)
	suite.Require().Equal(expPayload, payload)

	// Invoke ValidateToken method on service with mocked jwt
	payload, err = tokenService.ValidateToken(refreshToken)

	suite.Require().NoError(err)
	suite.Require().Equal(expPayload, payload)
}

func (suite *JWTTokenServiceSuite) TestGenerateAccessTokenError() {
	expUser := &models.User{ID: 1}
	expPayload := &Payload{UserID: expUser.ID}
	expKey := []byte("expected_key")
	expSigningMethod := jwt.SigningMethodHS256
	expAccessDuration := time.Hour
	expRefreshDuration := 7 * 24 * time.Hour

	expAccessClaims := jwtClaims{
		Payload: expPayload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expAccessDuration)),
		},
	}
	expAccessToken := jwt.NewWithClaims(
		expSigningMethod,
		expAccessClaims,
	)

	expSignedStringError := errors.New("expected_signed_string_error")

	mockImpl := mock_jwtinner.NewMockJWTInner(suite.mockController)

	mockImpl.EXPECT().
		SignedString(expAccessToken, expKey).
		Return("", expSignedStringError).
		Times(1)

	tokenService := NewJWT(
		JWTConfig{
			Key:             expKey,
			SigningMethod:   expSigningMethod,
			AccessDuration:  expAccessDuration,
			RefreshDuration: expRefreshDuration,
		},
		mockImpl,
	)

	// Invoke GenerateAccessToken method on service with mocked jwt
	accessToken, err := tokenService.GenerateAccessToken(expPayload)

	suite.Require().Equal(expSignedStringError, err)
	suite.Require().Equal("", accessToken)
}

func (suite *JWTTokenServiceSuite) TestGenerateRefreshTokenError() {
	expUser := &models.User{ID: 1}
	expPayload := &Payload{UserID: expUser.ID}
	expKey := []byte("expected_key")
	expSigningMethod := jwt.SigningMethodHS256
	expAccessDuration := time.Hour
	expRefreshDuration := 7 * 24 * time.Hour

	expRefreshClaims := jwtClaims{
		Payload: expPayload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expRefreshDuration)),
		},
	}
	expRefreshToken := jwt.NewWithClaims(
		expSigningMethod,
		expRefreshClaims,
	)

	expSignedStringError := errors.New("expected_signed_string_error")

	mockImpl := mock_jwtinner.NewMockJWTInner(suite.mockController)

	mockImpl.EXPECT().
		SignedString(expRefreshToken, expKey).
		Return("", expSignedStringError).
		Times(1)

	tokenService := NewJWT(
		JWTConfig{
			Key:             expKey,
			SigningMethod:   expSigningMethod,
			AccessDuration:  expAccessDuration,
			RefreshDuration: expRefreshDuration,
		},
		mockImpl,
	)

	// Invoke GenerateRefreshToken method on service with mocked jwt
	refreshToken, err := tokenService.GenerateRefreshToken(expPayload)

	suite.Require().Equal(expSignedStringError, err)
	suite.Require().Equal("", refreshToken)
}

func (suite *JWTTokenServiceSuite) TestValidateTokenError() {
	expUser := &models.User{ID: 1}
	expPayload := &Payload{UserID: expUser.ID}
	expKey := []byte("expected_key")
	expSigningMethod := jwt.SigningMethodHS256
	expExpiresAt := jwt.NewNumericDate(
		time.Now().Add(-1 * time.Hour),
	)

	expClaims := jwtClaims{
		Payload: expPayload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expExpiresAt,
		},
	}
	expToken := jwt.NewWithClaims(
		expSigningMethod,
		expClaims,
	)
	expTokenString, err := expToken.SignedString(
		expKey,
	)
	suite.Require().NoError(err)

	mockImpl := mock_jwtinner.NewMockJWTInner(suite.mockController)

	var expEmptyAccessClaims jwtClaims

	expParseWithClaimError := jwt.NewValidationError(
		"validation_error",
		jwt.ValidationErrorExpired,
	)
	expTokenServiceError := ErrInvalidToken

	mockImpl.EXPECT().
		ParseWithClaims(expTokenString, &expEmptyAccessClaims, gomock.Any()).
		Return(nil, expParseWithClaimError).
		Times(1)

	tokenService := NewJWT(
		JWTConfig{
			Key:             expKey,
			SigningMethod:   expSigningMethod,
			AccessDuration:  0,
			RefreshDuration: 0,
		},
		mockImpl,
	)

	// Invoke ValidateToken method on service with mocked jwt
	payload, err := tokenService.ValidateToken(expTokenString)

	suite.Require().Equal(expTokenServiceError, err)
	suite.Require().Nil(payload)
}
