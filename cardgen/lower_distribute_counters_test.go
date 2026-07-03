package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceDistributeCounters asserts that "Distribute N
// <kind> counters among <cardinality> target creatures[ you control]" with a
// fixed or variable X total lowers to a single multi-target spec and a Distribute
// AddCounter instruction whose maximum target count is capped at a fixed total
// and left at the wording's bound for a variable X.
func TestGenerateExecutableCardSourceDistributeCounters(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		wantedSnips []string
	}{
		{
			name:       "one or two target creatures",
			oracleText: "Distribute two +1/+1 counters among one or two target creatures.",
			wantedSnips: []string{
				"MinTargets: 1",
				"MaxTargets: 2",
				"Amount:      game.Fixed(2)",
				"Object:      game.AllTargetPermanentsReference(0)",
				"CounterKind: counter.PlusOnePlusOne",
				"Distribute:  true",
			},
		},
		{
			name:       "up to two target creatures",
			oracleText: "Distribute two +1/+1 counters among up to two target creatures.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 2",
				"Amount:      game.Fixed(2)",
				"Distribute:  true",
			},
		},
		{
			name:       "up to three target creatures you control",
			oracleText: "Distribute three +1/+1 counters among up to three target creatures you control.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 3",
				"Amount:      game.Fixed(3)",
				"Distribute:  true",
			},
		},
		{
			name:       "one two or three target creatures",
			oracleText: "Distribute three +1/+1 counters among one, two, or three target creatures.",
			wantedSnips: []string{
				"MinTargets: 1",
				"MaxTargets: 3",
				"Amount:      game.Fixed(3)",
				"Distribute:  true",
			},
		},
		{
			name:       "one two or three target creatures you control",
			oracleText: "Distribute three +1/+1 counters among one, two, or three target creatures you control.",
			wantedSnips: []string{
				"MinTargets: 1",
				"MaxTargets: 3",
				"Controller: game.ControllerYou",
				"Distribute:  true",
			},
		},
		{
			name:       "any number of target creatures",
			oracleText: "Distribute four +1/+1 counters among any number of target creatures.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 4",
				"Amount:      game.Fixed(4)",
				"Distribute:  true",
			},
		},
		{
			name:       "variable X among any number of target creatures",
			oracleText: "Distribute X +1/+1 counters among any number of target creatures.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 99",
				"Kind: game.DynamicAmountX",
				"Distribute:  true",
			},
		},
		{
			name:       "minus counters",
			oracleText: "Distribute two -1/-1 counters among one or two target creatures.",
			wantedSnips: []string{
				"MinTargets: 1",
				"MaxTargets: 2",
				"CounterKind: counter.MinusOneMinusOne",
				"Distribute:  true",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Rite",
				Layout:     "normal",
				ManaCost:   "{1}{G}",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
				Colors:     []string{"G"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range append([]string{"Primitive: game.AddCounter"}, test.wantedSnips...) {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceDistributeCountersFailsClosed asserts that
// distribute-counters wordings the executable backend cannot represent exactly
// stay rejected rather than being silently approximated.
func TestGenerateExecutableCardSourceDistributeCountersFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "trailing rider clause",
			oracleText: "Distribute three +1/+1 counters among one, two, or three target creatures, then double the number of +1/+1 counters on each of those creatures.",
		},
		{
			name:       "dynamic total",
			oracleText: "Distribute X +1/+1 counters among any number of target creatures, where X is the number of creatures you control.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Test Rite",
				Layout:     "normal",
				ManaCost:   "{1}{G}",
				TypeLine:   "Sorcery",
				OracleText: test.oracleText,
				Colors:     []string{"G"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "t")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want fail-closed rejection", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("expected a diagnostic for unsupported distribute counters")
			}
		})
	}
}
