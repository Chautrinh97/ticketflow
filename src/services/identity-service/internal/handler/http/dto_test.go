package http

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"ticketflow/services/identity-service/internal/model"
)

// Pure-unit test: toUserDTO has no external dependencies — plain
// testify/assert, no suite/mock needed, per backend-conventions.md's
// "Test logic thuần" carve-out.

const (
	testCaseSuccess_AllFieldsPresent = "[Success] non-nil FullName/AvatarURL map through as non-nil pointers"
	testCaseSuccess_NilPointerFields = "[Success] nil FullName/AvatarURL map through as nil (not empty string)"
)

func TestToUserDTO(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run(testCaseSuccess_AllFieldsPresent, func(t *testing.T) {
		fullName := "Nguyen Van A"
		avatarURL := "https://example.com/a.png"
		u := &model.User{
			ID:        "user-1",
			Email:     "a@example.com",
			FullName:  &fullName,
			AvatarURL: &avatarURL,
			Role:      model.RoleUser,
			Status:    model.StatusActive,
			CreatedAt: createdAt,
		}

		dto := toUserDTO(u)

		assert.Equal(t, "user-1", dto.ID)
		assert.Equal(t, "a@example.com", dto.Email)
		if assert.NotNil(t, dto.FullName) {
			assert.Equal(t, fullName, *dto.FullName)
		}
		if assert.NotNil(t, dto.AvatarURL) {
			assert.Equal(t, avatarURL, *dto.AvatarURL)
		}
		assert.Equal(t, model.RoleUser, dto.Role)
		assert.Equal(t, model.StatusActive, dto.Status)
		assert.Equal(t, createdAt.Format(time.RFC3339), dto.CreatedAt)
	})

	t.Run(testCaseSuccess_NilPointerFields, func(t *testing.T) {
		u := &model.User{
			ID:        "user-2",
			Email:     "b@example.com",
			FullName:  nil,
			AvatarURL: nil,
			Role:      model.RoleOrganizer,
			Status:    model.StatusPending,
			CreatedAt: createdAt,
		}

		dto := toUserDTO(u)

		assert.Nil(t, dto.FullName)
		assert.Nil(t, dto.AvatarURL)
		assert.Equal(t, model.RoleOrganizer, dto.Role)
		assert.Equal(t, model.StatusPending, dto.Status)
	})
}
