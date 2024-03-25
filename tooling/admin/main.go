package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"io"
	"os"
	"service/domain/data/schema"
	"service/domain/sys/database"
	"time"
)

func main() {
	err := genKey()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func genKey() error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("error while generating key %w", err)
	}

	//%w verb used only for errors and used to wrap/unwrap value
	privateFile, err := os.Create("private.pem")
	if err != nil {
		return fmt.Errorf("error while creating pem file %w", err)
	}
	defer privateFile.Close()

	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return fmt.Errorf("error encoding to pem file %w", err)
	}

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("error marshaling public key %w", err)
	}

	publicFile, err := os.Create("public.pem")
	if err != nil {
		return fmt.Errorf("error while creating pem file %w", err)
	}
	defer publicFile.Close()

	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	if err := pem.Encode(privateFile, &publicBlock); err != nil {
		return fmt.Errorf("error encoding to pem public file %w", err)
	}
	return nil
}

// TODO test genToken by makefile
func genToken() error {
	keyName := "private.pem"
	name := "zard/keys/" + keyName

	file, err := os.Open(name)
	if err != nil {
		return err
	}

	privatePEM, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return fmt.Errorf("reading auth private key %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return fmt.Errorf("parsing auth private key %w", err)
	}

	claims := struct {
		//we embedded the type here, if we do not embed it
		//and act as c style, our struct won't satisfy jwt.StandardClaims Interface
		jwt.StandardClaims
		Roles []string
	}{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "service project",
			Subject:   "123456789",
			ExpiresAt: time.Now().Add(7860 * time.Hour).Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod("RS256")
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = keyName

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return err
	}

	fmt.Println("TOKEN", tokenStr)

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("error marshaling public key %w", err)
	}

	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("error encoding to pem public file %w", err)
	}

	parser := jwt.Parser{
		ValidMethods: []string{"RS256"},
	}
	//jwt.NewParser()
	var ParsedClaims struct {
		jwt.StandardClaims
		Roles []string
	}

	keyFunc := func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id in header")
		}
		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("key id must be string")
		}
		fmt.Println("kid ", kidID)
		return &privateKey.PublicKey, nil
	}

	parsedToken, err := parser.ParseWithClaims(tokenStr, &ParsedClaims, keyFunc)
	if err != nil {
		return fmt.Errorf("parsing token %w", err)
	}

	if !parsedToken.Valid {
		return fmt.Errorf("token is not valid %w", err)
	}

	fmt.Println("TOKEN", tokenStr)
	return nil
}

func seed() error {
	cfg := database.Config{
		User:         "postgres",
		Password:     "postgres",
		Host:         "localhost",
		Name:         "postgres",
		MaxIdleConns: 0,
		MaxOpenConns: 0,
		DisableTLS:   true,
	}
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Seed(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}
	fmt.Println("seed data complete")
	return nil
}

func migrate() error {
	cfg := database.Config{
		User:         "postgres",
		Password:     "postgres",
		Host:         "localhost",
		Name:         "postgres",
		MaxIdleConns: 0,
		MaxOpenConns: 0,
		DisableTLS:   true,
	}
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := schema.Migrate(ctx, db); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}
	fmt.Println("migrate complete")
	return seed()
}
