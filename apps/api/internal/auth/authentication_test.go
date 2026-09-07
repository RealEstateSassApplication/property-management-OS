package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeTokenVerifier struct {
	principal Principal
	err       error
}

func (f fakeTokenVerifier) Verify(context.Context, string) (Principal, error) {
	return f.principal, f.err
}

type fakeIdentityRepository struct {
	userID  string
	issuer  string
	subject string
	email   string
	err     error
}

func (f *fakeIdentityRepository) ResolveIdentity(_ context.Context, issuer, subject, email string) (string, error) {
	f.issuer = issuer
	f.subject = subject
	f.email = email
	return f.userID, f.err
}

func TestAuthenticateResolvesVerifiedSubjectToInternalUser(t *testing.T) {
	identities := &fakeIdentityRepository{userID: "user-1"}
	authenticator := NewAuthenticator(fakeTokenVerifier{principal: Principal{
		Issuer: " https://issuer.example ", Subject: " subject-1 ", Email: " USER@EXAMPLE.COM ",
	}}, identities)

	identity, err := authenticator.Authenticate(context.Background(), "signed-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if identity.UserID != "user-1" || identities.issuer != "https://issuer.example" || identities.subject != "subject-1" || identities.email != "user@example.com" {
		t.Fatalf("unexpected resolved identity: %#v repo=%#v", identity, identities)
	}
}

func TestAuthenticateRejectsInvalidToken(t *testing.T) {
	authenticator := NewAuthenticator(fakeTokenVerifier{err: errors.New("signature failed")}, &fakeIdentityRepository{})
	_, err := authenticator.Authenticate(context.Background(), "bad-token")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}

func TestAuthenticatePropagatesUnlinkedIdentity(t *testing.T) {
	authenticator := NewAuthenticator(
		fakeTokenVerifier{principal: Principal{Issuer: "https://issuer.example", Subject: "subject-1"}},
		&fakeIdentityRepository{err: ErrIdentityNotLinked},
	)
	_, err := authenticator.Authenticate(context.Background(), "signed-token")
	if !errors.Is(err, ErrIdentityNotLinked) {
		t.Fatalf("expected identity-not-linked, got %v", err)
	}
}
