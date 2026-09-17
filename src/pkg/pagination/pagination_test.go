package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseParams(t *testing.T) {
	cases := []struct {
		name         string
		pageStr      string
		pageSizeStr  string
		wantPage     int
		wantPageSize int
	}{
		{"[Success] empty strings use defaults", "", "", DefaultPage, DefaultPageSize},
		{"[Success] valid values passed through", "3", "50", 3, 50},
		{"[Error] non-numeric page falls back to default", "abc", "20", DefaultPage, 20},
		{"[Error] non-numeric page_size falls back to default", "2", "xyz", 2, DefaultPageSize},
		{"[Error] page zero falls back to default", "0", "20", DefaultPage, 20},
		{"[Error] negative page falls back to default", "-5", "20", DefaultPage, 20},
		{"[Error] page_size zero falls back to default", "1", "0", 1, DefaultPageSize},
		{"[Error] negative page_size falls back to default", "1", "-10", 1, DefaultPageSize},
		{"[Success] page_size at max boundary kept as-is", "1", "100", 1, MaxPageSize},
		{"[Error] page_size above max clamped to max", "1", "101", 1, MaxPageSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseParams(tc.pageStr, tc.pageSizeStr)
			assert.Equal(t, tc.wantPage, got.Page)
			assert.Equal(t, tc.wantPageSize, got.PageSize)
		})
	}
}

func TestParams_OffsetAndLimit(t *testing.T) {
	p := Params{Page: 1, PageSize: 20}
	assert.Equal(t, 0, p.Offset())
	assert.Equal(t, 20, p.Limit())

	p = Params{Page: 3, PageSize: 20}
	assert.Equal(t, 40, p.Offset())
	assert.Equal(t, 20, p.Limit())
}

func TestNew(t *testing.T) {
	t.Run("[Success] nil items coerced to empty slice, not nil", func(t *testing.T) {
		env := New[string](nil, 0)
		assert.NotNil(t, env.Items)
		assert.Empty(t, env.Items)
		assert.Equal(t, int64(0), env.Total)
	})

	t.Run("[Success] non-nil items preserved with total", func(t *testing.T) {
		env := New([]int{1, 2, 3}, 3)
		assert.Equal(t, []int{1, 2, 3}, env.Items)
		assert.Equal(t, int64(3), env.Total)
	})
}
