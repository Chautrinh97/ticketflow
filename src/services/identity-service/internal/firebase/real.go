package firebase

import (
	"context"

	firebaseadmin "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// RealVerifier calls the Firebase Admin SDK. Used when AUTH_FIREBASE_MODE=real
// and FIREBASE_CREDENTIALS_FILE points at a service-account JSON key.
type RealVerifier struct {
	client *auth.Client
}

func NewRealVerifier(ctx context.Context, credentialsFile string) (*RealVerifier, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	app, err := firebaseadmin.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return &RealVerifier{client: client}, nil
}

func (r *RealVerifier) Verify(ctx context.Context, idToken string) (*VerifiedToken, error) {
	token, err := r.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	email, _ := token.Claims["email"].(string)
	name, _ := token.Claims["name"].(string)
	return &VerifiedToken{UID: token.UID, Email: email, Name: name}, nil
}
