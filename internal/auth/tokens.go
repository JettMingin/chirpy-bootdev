package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
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
		return uuid.UUID{}, errors.New("auth/tokens.go: failed to parse token for claims")
	}

}

func GetBearerToken(headers http.Header) (string, error) {
	authHeaders, ok := headers["Authorization"]
	if !ok {
		return "", errors.New("auth/tokens.go: failed to get 'Authorization' Header Value")
	}
	authToken := ""
	for _, el := range authHeaders {
		if strings.HasPrefix(el, "Bearer") {
			el = strings.TrimPrefix(el, "Bearer")
			authToken = strings.Trim(el, " ")
			break
		}
	}
	if authToken == "" {
		return "", errors.New("auth/tokens.go: failed to extract token from Authorization Header")
	}
	return authToken, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}
