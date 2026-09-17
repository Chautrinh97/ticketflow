package service

import "testing"

// slugify is a pure, zero-dependency function (regex + string trimming, no
// DB access) — per docs/01-architecture/backend-conventions.md's "Test logic
// thuần" rule, a plain table-driven test is enough, no suite/mock needed.
func TestSlugify(t *testing.T) {
	cases := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "[Success] Tiêu đề thường, có dấu cách -> nối bằng gạch ngang",
			title: "Rock Concert 2024",
			want:  "rock-concert-2024",
		},
		{
			name:  "[Success] Chữ hoa, dấu câu và khoảng trắng thừa bị chuẩn hoá",
			title: "  Hello, World!  ",
			want:  "hello-world",
		},
		{
			name:  "[Success] Chuỗi rỗng -> fallback \"event\"",
			title: "",
			want:  "event",
		},
		{
			name:  "[Success] Chỉ toàn ký tự đặc biệt -> fallback \"event\"",
			title: "!!!",
			want:  "event",
		},
		{
			name:  "[Success] Chỉ toàn khoảng trắng -> fallback \"event\"",
			title: "   ",
			want:  "event",
		},
		{
			name:  "[Success] Gạch ngang lặp lại (đầu/cuối/giữa) bị gộp còn 1 và trim 2 đầu",
			title: "---already-slugged---",
			want:  "already-slugged",
		},
		{
			name:  "[Success] Ký tự Unicode có dấu bị coi là non-alnum và loại bỏ",
			title: "Café 2024",
			want:  "caf-2024",
		},
		{
			name:  "[Success] Số và chữ trộn lẫn, dấu chấm cuối bị trim",
			title: "123 Main St.",
			want:  "123-main-st",
		},
		{
			name:  "[Success] Emoji/ký tự biểu tượng -> fallback \"event\"",
			title: "★★★",
			want:  "event",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := slugify(tc.title)
			if got != tc.want {
				t.Fatalf("slugify(%q) = %q, want %q", tc.title, got, tc.want)
			}
		})
	}
}
