package templater

import (
	"testing"
)

func TestExtractVarNames(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"{{.NAME}}", []string{"NAME"}},
		{"{{.FOO}} {{.BAR}}", []string{"FOO", "BAR"}},
		{"{{.FOO}} {{.FOO}}", []string{"FOO"}}, // dedup
		{"no vars here", nil},
		{"{{printf \"%s\" .NAME}}", []string{"NAME"}},
		{"{{.NAME | trim | upper}}", []string{"NAME"}},
		{"{{.A_B_C}}", []string{"A_B_C"}},
		{"", nil},
		{".lowercase", nil}, // lowercase not matched
	}
	for _, tt := range tests {
		got := extractVarNames(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("extractVarNames(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("extractVarNames(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
