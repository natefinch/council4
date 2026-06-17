package cluster

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"if you do", "if you do"},
		{"  If You  Do ", "if you do"},
		{"deal 3 damage", "deal N damage"},
		{"crew 2", "crew N"},
		{"crew 12", "crew N"},
		{"draw 1 card, then draw 2", "draw N card, then draw N"},
		{"", ""},
		{"   ", ""},
		{"R2D2", "r2d2"},
	}
	for _, tc := range cases {
		if got := Normalize(tc.in); got != tc.want {
			t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
