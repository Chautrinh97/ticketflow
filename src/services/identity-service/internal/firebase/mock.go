package firebase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
)

// MockVerifier decodes a locally-crafted "token" — base64(JSON{uid,email,name})
// — instead of calling real Firebase, so `docker compose up` and dev/test
// flows work without a provisioned Firebase project. Never used when
// AUTH_FIREBASE_MODE=real.
type MockVerifier struct{}

func NewMockVerifier() *MockVerifier { return &MockVerifier{} }

func (m *MockVerifier) Verify(ctx context.Context, idToken string) (*VerifiedToken, error) {
	raw, err := base64.StdEncoding.DecodeString(idToken)
	if err != nil {
		return nil, errors.New("invalid mock firebase token")
	}
	var t VerifiedToken
	if err := json.Unmarshal(raw, &t); err != nil || t.UID == "" || t.Email == "" {
		return nil, errors.New("invalid mock firebase token payload")
	}
	return &t, nil
}
