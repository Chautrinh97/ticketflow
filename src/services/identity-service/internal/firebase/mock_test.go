package firebase

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pure-unit test: MockVerifier has no external dependencies (no Firebase,
// no DB, no Redis), so this runs with plain testify/assert — no
// suite/mock needed, per backend-conventions.md's "Test logic thuần" carve-out.

const (
	testCaseError_InvalidBase64  = "[Error] idToken not valid base64 is rejected"
	testCaseError_InvalidJSON    = "[Error] valid base64 but invalid JSON payload is rejected"
	testCaseError_EmptyUID       = "[Error] valid JSON but empty uid is rejected"
	testCaseError_EmptyEmail     = "[Error] valid JSON but empty email is rejected"
	testCaseSuccess_ValidPayload = "[Success] valid base64 JSON with uid+email+name decodes"
	testCaseSuccess_NoNameField  = "[Success] valid payload without name still succeeds (name is optional)"
)

func encodeMockPayload(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func TestMockVerifier_Verify(t *testing.T) {
	v := NewMockVerifier()

	cases := []struct {
		name      string
		idToken   string
		wantErr   bool
		wantToken *VerifiedToken
	}{
		{
			name:    testCaseError_InvalidBase64,
			idToken: "not-valid-base64!!!",
			wantErr: true,
		},
		{
			name:    testCaseError_InvalidJSON,
			idToken: encodeMockPayload("{not valid json"),
			wantErr: true,
		},
		{
			name:    testCaseError_EmptyUID,
			idToken: encodeMockPayload(`{"uid":"","email":"a@example.com","name":"A"}`),
			wantErr: true,
		},
		{
			name:    testCaseError_EmptyEmail,
			idToken: encodeMockPayload(`{"uid":"uid-1","email":"","name":"A"}`),
			wantErr: true,
		},
		{
			name:      testCaseSuccess_ValidPayload,
			idToken:   encodeMockPayload(`{"uid":"uid-1","email":"a@example.com","name":"Nguyen Van A"}`),
			wantErr:   false,
			wantToken: &VerifiedToken{UID: "uid-1", Email: "a@example.com", Name: "Nguyen Van A"},
		},
		{
			name:      testCaseSuccess_NoNameField,
			idToken:   encodeMockPayload(`{"uid":"uid-2","email":"b@example.com"}`),
			wantErr:   false,
			wantToken: &VerifiedToken{UID: "uid-2", Email: "b@example.com", Name: ""},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := v.Verify(context.Background(), tc.idToken)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantToken, got)
		})
	}
}
