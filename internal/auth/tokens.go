package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type MyCustomClaims struct {
	Foo string `json:"foo"`
	jwt.RegisteredClaims
}

func MakeJWT(userID uuid.UUID, tokenSeret string, expiresIn time.Duration) (string, error) {
	tokenClaims := MyCustomClaims{
		"bar",
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			Subject:   userID.String(),
		},
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	signedString, err := newToken.SignedString([]byte(tokenSeret))
	if err != nil {
		return "", err
	}
	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	if tokenClaims, ok := parsedToken.Claims.(*MyCustomClaims); ok {
		return uuid.MustParse(tokenClaims.RegisteredClaims.Subject), nil
	} else {
		return uuid.UUID{}, errors.New("auth/tokens.go: failed to parse token fo claims")
	}

}
