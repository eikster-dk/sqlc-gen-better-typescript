package native

import (
	"strings"
	"testing"
)

func TestCleanWhitespace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string produces single newline",
			input: "",
			want:  "\n",
		},
		{
			name:  "single line without trailing newline gets one added",
			input: "hello",
			want:  "hello\n",
		},
		{
			name:  "single line with trailing newline preserved",
			input: "hello\n",
			want:  "hello\n",
		},
		{
			name:  "two consecutive newlines preserved",
			input: "a\n\nb",
			want:  "a\n\nb\n",
		},
		{
			name:  "three consecutive newlines collapsed to two",
			input: "a\n\n\nb",
			want:  "a\n\nb\n",
		},
		{
			name:  "many consecutive newlines collapsed to two",
			input: "a\n\n\n\n\n\nb",
			want:  "a\n\nb\n",
		},
		{
			name:  "trailing spaces on line removed",
			input: "hello   \nworld",
			want:  "hello\nworld\n",
		},
		{
			name:  "trailing tabs on line removed",
			input: "hello\t\t\nworld",
			want:  "hello\nworld\n",
		},
		{
			name:  "leading and trailing whitespace trimmed",
			input: "\n\n  hello  \n\n",
			want:  "hello\n",
		},
		{
			name:  "whitespace-only input produces single newline",
			input: "   \n\n\t\t\n   ",
			want:  "\n",
		},
		{
			name:  "mixed: excessive newlines and trailing spaces",
			input: "line1   \n\n\n\nline2\t\n\n\n\n\nline3",
			want:  "line1\n\nline2\n\nline3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("cleanWhitespace(%q)\ngot:  %q\nwant: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanWhitespace_AlwaysEndsWithNewline(t *testing.T) {
	inputs := []string{"", "a", "a\n", "a\nb", "a\nb\n", "\n\n\n"}
	for _, input := range inputs {
		got := cleanWhitespace(input)
		if !strings.HasSuffix(got, "\n") {
			t.Errorf("cleanWhitespace(%q) = %q, does not end with newline", input, got)
		}
	}
}

func TestCleanWhitespace_NeverHasThreeConsecutiveNewlines(t *testing.T) {
	inputs := []string{
		"a\n\n\n\n\nb",
		"\n\n\n\n\n",
		"a\n\n\nb\n\n\nc",
	}
	for _, input := range inputs {
		got := cleanWhitespace(input)
		if strings.Contains(got, "\n\n\n") {
			t.Errorf("cleanWhitespace(%q) = %q, contains three consecutive newlines", input, got)
		}
	}
}
