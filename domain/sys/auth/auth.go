package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
)

type KeyLookup interface {
	PrivateKey(kid string) (*rsa.PrivateKey, error)
	PublicKey(kid string) (*rsa.PublicKey, error)
}

type Auth struct {
	activeKID string
	keyLookup KeyLookup
	method    jwt.SigningMethod
	keyFunc   func(t *jwt.Token) (any, error)
	parser    jwt.Parser
}

func New(activeKID string, lookup KeyLookup) (*Auth, error) {
	_, err := lookup.PrivateKey(activeKID)
	if err != nil {
		return nil, errors.New("active kid doesn't exist in store")
	}

	method := jwt.GetSigningMethod("RS256")
	if method == nil {
		return nil, errors.New("error while getting signing method")
	}

	keyFunc := func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing kid error")
		}

		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("kid must be string")
		}
		return lookup.PublicKey(kidID)
	}

	jwtParser := jwt.Parser{
		ValidMethods: []string{"RS256"},
	}

	a := Auth{
		activeKID: activeKID,
		keyLookup: lookup,
		method:    method,
		keyFunc:   keyFunc,
		parser:    jwtParser,
	}
	return &a, nil
}

func (a *Auth) GenerateToken(claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = a.activeKID

	privateKey, err := a.keyLookup.PrivateKey(a.activeKID)
	if err != nil {
		return "", errors.New("")
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.New("")
	}
	return str, nil
}

func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)

	if err != nil {
		return Claims{}, fmt.Errorf("%w", "whats this")
	}

	if !token.Valid {
		return Claims{}, errors.New("")
	}
	return claims, nil
}
