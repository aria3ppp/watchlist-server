package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TODO: extract token service method into stand-alone functions and delete token service
type JWT struct {
	jwtInner        JWTInner
	key             any
	signingMethod   jwt.SigningMethod
	accessDuration  time.Duration
	refreshDuration time.Duration
}

var _ Service = (*JWT)(nil)

type JWTConfig struct {
	Key             any
	SigningMethod   jwt.SigningMethod
	AccessDuration  time.Duration
	RefreshDuration time.Duration
}

func NewJWT(
	config JWTConfig,
	impl ...JWTInner,
) *JWT {
	var jwtImpl JWTInner
	if len(impl) > 0 {
		jwtImpl = impl[0]
	} else {
		jwtImpl = jwtInnerImpl{}
	}
	return &JWT{
		key:             config.Key,
		signingMethod:   config.SigningMethod,
		accessDuration:  config.AccessDuration,
		refreshDuration: config.RefreshDuration,
		jwtInner:        jwtImpl,
	}
}

type jwtClaims struct {
	*Payload
	jwt.RegisteredClaims
}

func (ts *JWT) GenerateAccessToken(
	payload *Payload,
) (string, error) {
	return generateToken(
		ts.jwtInner,
		ts.key,
		ts.signingMethod,
		payload,
		ts.accessDuration,
	)
}

func (ts *JWT) GenerateRefreshToken(
	payload *Payload,
) (string, error) {
	return generateToken(
		ts.jwtInner,
		ts.key,
		ts.signingMethod,
		payload,
		ts.refreshDuration,
	)
}

func generateToken(
	jwtImpl JWTInner,
	key any,
	signingMethod jwt.SigningMethod,
	payload *Payload,
	duration time.Duration,
) (string, error) {
	claims := jwtClaims{
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}

	tokenWithClaim := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := jwtImpl.SignedString(
		tokenWithClaim,
		key,
	)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (ts *JWT) ValidateToken(
	tokenString string,
) (*Payload, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != ts.signingMethod.Alg() {
			return nil, fmt.Errorf(
				"unexpected jwt signing method: %s, expected: %v",
				t.Method.Alg(),
				ts.signingMethod.Alg(),
			)
		}
		return ts.key, nil
	}
	var claims jwtClaims
	tkn, err := ts.jwtInner.ParseWithClaims(
		tokenString,
		&claims,
		keyFunc,
	)
	if err == nil && tkn.Valid {
		return claims.Payload, nil
	}
	if err != nil {
		if _, ValidationError := err.(*jwt.ValidationError); ValidationError {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	// highly possible unreachable code b/c of jwt library bad design
	return nil, ErrInvalidToken
}

//go:generate mockgen -package mock_jwtinner -destination mock_jwtinner/mock_jwtinner.go . JWTInner

// jwt inner interface api
type JWTInner interface {
	SignedString(token *jwt.Token, key any) (string, error)
	ParseWithClaims(
		tokenString string,
		claims jwt.Claims,
		keyFunc jwt.Keyfunc,
		options ...jwt.ParserOption,
	) (*jwt.Token, error)
}

type jwtInnerImpl struct{}

var _ JWTInner = jwtInnerImpl{}

func (impl jwtInnerImpl) SignedString(
	token *jwt.Token,
	key any,
) (string, error) {
	return token.SignedString(key)
}

func (impl jwtInnerImpl) ParseWithClaims(
	tokenString string,
	claims jwt.Claims,
	keyFunc jwt.Keyfunc,
	options ...jwt.ParserOption,
) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, keyFunc, options...)
}
