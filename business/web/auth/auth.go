// Package auth provides authentication and authorization support.
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// ErrForbidden is returned when a auth issue is identified.
var ErrForbidden = errors.New("attempted action is not allowed")

// KeyLookup declares a method set of behavior for looking up
// private and public keys for JWT use.
type KeyLookup interface {
	PrivateKeyPEM(kid string) (pem string, err error)
	PublicKeyPEM(kid string) (pem string, err error)
}

// Config represents information required to initialize auth.
type Config struct {
	Log       *zap.SugaredLogger
	KeyLookup KeyLookup
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Auth struct {
	log       *zap.SugaredLogger
	keyLookup KeyLookup
	method    jwt.SigningMethod
	parser    *jwt.Parser
	mu        sync.RWMutex
	cache     map[string]string
}

// New creates an Auth to support authentication/authorization.
func New(cfg Config) (*Auth, error) {
	a := Auth{
		log:       cfg.Log,
		keyLookup: cfg.KeyLookup,
		method:    jwt.GetSigningMethod("RS256"),
		parser:    jwt.NewParser(jwt.WithValidMethods([]string{"RS256"})),
		cache:     make(map[string]string),
	}

	return &a, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(kid string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = kid

	privateKeyPEM, err := a.keyLookup.PrivateKeyPEM(kid)
	if err != nil {
		return "", fmt.Errorf("private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("parsing private pem: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}

// Authenticate processes the token to validate the sender's token is valid.
func (a *Auth) Authenticate(ctx context.Context, bearerToken string) (Claims, error) {
	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return Claims{}, errors.New("expected authorization header format: Bearer <token>")
	}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		kid, exists := token.Header["kid"]
		if !exists {
			return nil, errors.New("kid not in header")
		}

		pem, err := a.keyLookup.PublicKeyPEM(kid.(string))
		if err != nil {
			return nil, err
		}

		publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pem))
		if err != nil {
			return "", fmt.Errorf("parsing public pem: %w", err)
		}

		return publicKey, nil
	}

	var claims Claims
	if _, err := a.parser.ParseWithClaims(parts[1], &claims, keyFunc); err != nil {
		return Claims{}, fmt.Errorf("parse with claims: %w", err)
	}

	return claims, nil
}
