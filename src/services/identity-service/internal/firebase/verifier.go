// Package firebase wraps Firebase ID token verification behind an
// interface with two implementations: MockVerifier (default, for local
// dev/docker-compose without a provisioned Firebase project) and
// RealVerifier (Firebase Admin SDK), selected via AUTH_FIREBASE_MODE.
package firebase

import "context"

type VerifiedToken struct {
	UID   string
	Email string
	Name  string
}

type Verifier interface {
	Verify(ctx context.Context, idToken string) (*VerifiedToken, error)
}
