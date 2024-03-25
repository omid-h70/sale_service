package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v4"
	"testing"
	"time"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

func TestAuth(t *testing.T) {

	t.Log("Given the Need to be able to authenticate and authorize access")
	{
		testID := 0
		t.Logf("\t Test %d \t When Handeling Single User", testID)
		{
			kID := "private.pem"
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatalf("\t %s \tTest %d \t Failed while Creating Private key %v", failure, testID, err)
			}
			t.Logf("\t %s \t Test %d \t Private Key Created", success, testID)

			a, err := New(kID, &testKeyStore{privateKey})
			if err != nil {
				t.Fatalf("\t %s \tTest %d \t Failed while Creating Authentication %v", failure, testID, err)
			}
			t.Logf("\t %s \t Test %d \t Authtenticator Created", success, testID)

			claims := Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "service project",
					Subject:   "ABCD",
					ExpiresAt: time.Now().Add(7860 * time.Hour).Unix(),
					IssuedAt:  time.Now().UTC().Unix(),
				},
				Roles: []string{RoleAdmin},
			}

			token, err := a.GenerateToken(claims)
			if err != nil {
				t.Fatalf("\t %s \tTest %d \t Failed When Creating Token %v", failure, testID, err)
			}
			t.Logf("\t %s \t Test %d \t Token Created", success, testID)

			parsedClaims, err := a.ValidateToken(token)
			if err != nil {
				t.Fatalf("\t %s \tTest %d \t Failed When Validating Token %v", failure, testID, err)
			}
			t.Logf("\t %s \t Test %d \t Token Validated", success, testID)

			if exp, got := len(claims.Roles), len(parsedClaims.Roles); exp != got {
				t.Logf("\t Test %d \t Exp %d", testID, exp)
				t.Logf("\t Test %d \t Got %d", testID, got)
				t.Fatalf("\t %s \t Test %d \t Failed", failure, err)
			}
			t.Logf("\t %s \t Test %d \t Got Expected number of rows", success, testID)

			if exp, got := claims.Roles[0], parsedClaims.Roles[0]; exp != got {
				t.Logf("\t Test %d \t Exp %d", testID, exp)
				t.Logf("\t Test %d \t Got %d", testID, got)
				t.Fatalf("\t %s \t Test %d \t Failed", failure, err)
			}
			t.Logf("\t %s \t Test %d \t Got Expected roles", success, testID)
		}
	}
}

type testKeyStore struct {
	pk *rsa.PrivateKey
}

func (ks *testKeyStore) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	return ks.pk, nil
}

func (ks *testKeyStore) PublicKey(kid string) (*rsa.PublicKey, error) {
	return &ks.pk.PublicKey, nil
}
