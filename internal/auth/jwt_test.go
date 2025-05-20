package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTs(t *testing.T) {
	secret := "test_secret"
	newID, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("test-NULL - failed generating new uuid: %s", err)
	}

	t.Logf("preTest, newID: %s\n", newID.String())

	tokenString, err := MakeJWT(newID, secret, time.Minute)
	if err != nil {
		t.Fatalf("test-FAIL: MakeJWT error: %s", err)
	} else {
		t.Logf("test-PASS: MakeJWT returned a token, token length: %v\n", len(tokenString))
	}

	//test correct secret
	returnedID, err2 := ValidateJWT(tokenString, secret)
	if err2 != nil {
		t.Fatalf("test-FAIL: ValidateJWT error: %s", err2)
	} else {
		t.Logf("test-PASS: validateJWT returned a uuid: %s", returnedID.String())
	}

	//test bad secret
	_, err3 := ValidateJWT(tokenString, "bad_secret")
	if err3 != nil {
		t.Logf("test-PASS: validateJWT rejected bad secret: %s", err3)
	} else {
		t.Fatalf("test-FAIL: ValidateJWT error")
	}
}

func TestGetBearerToken(t *testing.T) {
	//test1 - good header + value
	headers1 := http.Header{
		"Authorization": []string{"hello", "Bearer TOKENSTRING", "goodbye"},
		"Content-Type":  []string{"application/json"},
	}
	result1, err := GetBearerToken(headers1)
	if err != nil {
		t.Fatalf("test-FAIL: GetBearerToken returned err on valid header")
	} else {
		t.Logf("test-PASS: GetBearerToken returned this: %s", result1)
	}

	//test2 - good header but no val
	headers2 := http.Header{
		"Authorization": []string{"hello", "nobearer", "goodbye"},
		"Content-Type":  []string{"application/json"},
	}
	result2, err := GetBearerToken(headers2)
	if err != nil {
		t.Logf("test-PASS: GetBearerToken rejected good header with no valid token")
	} else {
		t.Fatalf("test-FAIL: GetBearerToken returned a str with no valid token value: %s", result2)
	}

	//test3 - bad header
	headers3 := http.Header{
		"Content-Type": []string{"application/json"},
	}
	result3, err := GetBearerToken(headers3)
	if err != nil {
		t.Logf("test-PASS: GetBearerToken rejected bad header")
	} else {
		t.Fatalf("test-FAIL: GetBearerToken returned a str with no valid header: %s", result3)
	}
}

func TestRefreshTokens(t *testing.T) {
	t.Logf("Token1: %s", MakeRefreshToken())
	t.Logf("Token2: %s", MakeRefreshToken())
	t.Logf("Token3: %s", MakeRefreshToken())
}
