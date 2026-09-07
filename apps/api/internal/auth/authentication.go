package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Principal struct {
	Issuer  string
	Subject string
	Email   string
}

type AuthenticatedIdentity struct {
	UserID  string
	Issuer  string
	Subject string
	Email   string
}

type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (Principal, error)
}

type IdentityRepository interface {
	ResolveIdentity(ctx context.Context, issuer, subject, email string) (string, error)
}

var (
	ErrInvalidToken                = errors.New("invalid bearer token")
	ErrIdentityNotLinked           = errors.New("verified identity is not linked to a local user")
	ErrAuthenticationNotConfigured = errors.New("authentication is not configured")
)

type Authenticator struct {
	verifier   TokenVerifier
	identities IdentityRepository
}

func NewAuthenticator(verifier TokenVerifier, identities IdentityRepository) *Authenticator {
	return &Authenticator{verifier: verifier, identities: identities}
}

func (a *Authenticator) Authenticate(ctx context.Context, rawToken string) (AuthenticatedIdentity, error) {
	if a == nil || a.verifier == nil || a.identities == nil {
		return AuthenticatedIdentity{}, ErrAuthenticationNotConfigured
	}
	if strings.TrimSpace(rawToken) == "" {
		return AuthenticatedIdentity{}, ErrInvalidToken
	}

	principal, err := a.verifier.Verify(ctx, rawToken)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return AuthenticatedIdentity{}, err
		}
		return AuthenticatedIdentity{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	principal.Issuer = strings.TrimSpace(principal.Issuer)
	principal.Subject = strings.TrimSpace(principal.Subject)
	principal.Email = strings.TrimSpace(strings.ToLower(principal.Email))
	if principal.Issuer == "" || principal.Subject == "" {
		return AuthenticatedIdentity{}, ErrInvalidToken
	}

	userID, err := a.identities.ResolveIdentity(ctx, principal.Issuer, principal.Subject, principal.Email)
	if err != nil {
		return AuthenticatedIdentity{}, err
	}
	return AuthenticatedIdentity{
		UserID:  userID,
		Issuer:  principal.Issuer,
		Subject: principal.Subject,
		Email:   principal.Email,
	}, nil
}
