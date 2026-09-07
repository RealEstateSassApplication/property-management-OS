package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

type OIDCVerifier struct {
	issuer   string
	verifier *oidc.IDTokenVerifier
}

func NewOIDCVerifier(ctx context.Context, issuer, audience string) (*OIDCVerifier, error) {
	issuer = strings.TrimSpace(issuer)
	audience = strings.TrimSpace(audience)
	if issuer == "" {
		return nil, fmt.Errorf("OIDC issuer is required")
	}
	if audience == "" {
		return nil, fmt.Errorf("OIDC audience is required")
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}
	return &OIDCVerifier{
		issuer:   issuer,
		verifier: provider.Verifier(&oidc.Config{ClientID: audience}),
	}, nil
}

func (v *OIDCVerifier) Verify(ctx context.Context, rawToken string) (Principal, error) {
	if v == nil || v.verifier == nil {
		return Principal{}, ErrAuthenticationNotConfigured
	}
	idToken, err := v.verifier.Verify(ctx, strings.TrimSpace(rawToken))
	if err != nil {
		return Principal{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return Principal{}, fmt.Errorf("%w: decode claims: %v", ErrInvalidToken, err)
	}
	if strings.TrimSpace(idToken.Subject) == "" {
		return Principal{}, fmt.Errorf("%w: subject claim is required", ErrInvalidToken)
	}

	return Principal{
		Issuer:  v.issuer,
		Subject: idToken.Subject,
		Email:   claims.Email,
	}, nil
}
