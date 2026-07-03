package parser

import "testing"

// TestTitleCaseHelpersHandleMultibyte verifies the parser title-case helpers
// operate on the leading rune rather than the leading byte, so a multi-byte
// first character (an accented "é") is not corrupted into U+FFFD when the
// byte-exact clause reconstruction re-cases a word.
func TestTitleCaseHelpersHandleMultibyte(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"éomer", "Éomer"},
		{"bolt", "Bolt"},
		{"", ""},
	}
	for _, c := range cases {
		if got := titleCaseWord(c.in); got != c.want {
			t.Errorf("titleCaseWord(%q) = %q, want %q", c.in, got, c.want)
		}
		if got := titleFirstEffectText(c.in); got != c.want {
			t.Errorf("titleFirstEffectText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
