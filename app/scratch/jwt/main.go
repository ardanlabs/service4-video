package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {

	// =========================================================================
	// Generate Private / Public RSA Key

	const keyFile = "zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem"
	file, err := os.Open(keyFile)
	if err != nil {
		return fmt.Errorf("opening key file: %w", err)
	}
	defer file.Close()

	privatePEM, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return fmt.Errorf("reading auth private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privatePEM))
	if err != nil {
		return fmt.Errorf("parsing public pem: %w", err)
	}

	// Create a file for the private key information in PEM form.
	// privateFile, err := os.Create("private.pem")
	// if err != nil {
	// 	return fmt.Errorf("creating private file: %w", err)
	// }
	// defer privateFile.Close()

	// Construct a PEM block for the private key.
	// privateBlock := pem.Block{
	// 	Type:  "PRIVATE KEY",
	// 	Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	// }

	// Write the private key to the private key file.
	// if err := pem.Encode(privateFile, &privateBlock); err != nil {
	// 	return fmt.Errorf("encoding to private file: %w", err)
	// }

	// Marshal the public key from the private key to PKIX.
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	fmt.Print("========================================\n\n")

	// Write the public key to the public key file.
	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	// =========================================================================
	// Generate JWT with Signature

	fmt.Print("\n========================================\n\n")

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "123456789",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"USER"},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), claims)
	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	fmt.Println(tokenString)

	// =========================================================================
	// Validate JWT with Public Key

	fmt.Print("\n========================================\n\n")

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		switch token.Header["kid"] {
		case "kid1":
			return &privateKey.PublicKey, nil
		default:
			return nil, errors.New("unknown key")
		}
	}

	var clm struct {
		jwt.RegisteredClaims
		Roles []string
	}
	if _, err := parser.ParseWithClaims(tokenString, &clm, keyFunc); err != nil {
		return fmt.Errorf("parse with claims: %w", err)
	}

	fmt.Print("signature validated\n\n")
	fmt.Printf("%#v", clm)

	return nil
}
