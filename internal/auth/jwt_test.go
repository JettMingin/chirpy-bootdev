package auth

import (
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
	}
	t.Logf("test-PASS: MakeJWT returned string: %s\n", tokenString)

	returnId, err := ValidateJWT(tokenString, secret)
	if err == nil {
		t.Fatalf("test-FAIL: ValidateJWT error: %s", err)
	} else {
		t.Logf("test-PASS: validateJWT returned string: %s\n", returnId.String())
	}
}
