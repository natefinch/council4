package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceDividedDamage asserts that "deals N damage
// divided as you choose among <cardinality> <targets>" with a fixed total lowers
// to a single multi-target spec and a Divided Damage instruction whose maximum
// target count is capped at the total.
func TestGenerateExecutableCardSourceDividedDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "any number of targets",
			oracleText: "Test Bolt deals 3 damage divided as you choose among any number of targets.",
			wantedSnips: []string{
				"MinTargets: 1",
				"MaxTargets: 3",
				"Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer",
				"Amount:    game.Fixed(3)",
				"Recipient: game.AnyTargetDamageRecipient(0)",
				"Divided:   true",
			},
		},
		{
			name:       "one or two targets",
			oracleText: "Test Bolt deals 2 damage divided as you choose among one or two targets.",
			wantedSnips: []string{
				"MinTargets: 1",
				"MaxTargets: 2",
				"Divided:   true",
			},
		},
		{
			name:       "one two or three target creatures",
			oracleText: "Test Bolt deals 4 damage divided as you choose among one, two, or three target creatures.",
			wantedSnips: []string{
				"MinTargets: 1",
				"MaxTargets: 3",
				"PermanentTypes: []types.Card{types.Creature}",
				"Divided:   true",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
				Colors:     []string{"R"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range append([]string{"Primitive: game.Damage"}, test.wantedSnips...) {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceDividedDamageFailsClosed asserts that
// divided-damage wordings the executable backend cannot represent exactly stay
// rejected rather than being silently approximated.
func TestGenerateExecutableCardSourceDividedDamageFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "variable total",
			oracleText: "Test Bolt deals X damage divided as you choose among any number of targets.",
		},
		{
			name:       "keyword filtered creatures",
			oracleText: "Test Bolt deals 4 damage divided as you choose among any number of target creatures with flying.",
		},
		{
			name:       "combat state filtered creatures",
			oracleText: "Test Bolt deals 5 damage divided as you choose among any number of target attacking creatures.",
		},
		{
			name:       "type union",
			oracleText: "Test Bolt deals 5 damage divided as you choose among any number of target creatures and/or planeswalkers.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
				Colors:     []string{"R"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want fail-closed rejection", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected a diagnostic for unsupported divided damage")
			}
		})
	}
}
