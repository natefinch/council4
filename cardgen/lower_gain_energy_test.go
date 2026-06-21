package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateGainEnergySource(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		oracle string
		amount string
	}{
		{"You get {E}{E} (two energy counters).", "game.Fixed(2)"},
		{"You get {E} (an energy counter).", "game.Fixed(1)"},
		{"You get {E}{E}{E}{E} (four energy counters).", "game.Fixed(4)"},
	} {
		t.Run(tc.oracle, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Energy Maker",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			}, "e")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range []string{
				"game.AddPlayerCounter",
				"counter.Energy",
				"game.ControllerReference()",
				tc.amount,
			} {
				if !strings.Contains(source, want) {
					t.Fatalf("generated source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

func TestGainEnergyDoesNotShadowPowerToughness(t *testing.T) {
	t.Parallel()
	// A normal "gets +1/+1" power/toughness modification must still lower as a
	// P/T change, not as an energy gain.
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Buff Maker",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +1/+1 until end of turn.",
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(source, "counter.Energy") {
		t.Fatalf("power/toughness buff wrongly lowered to energy:\n%s", source)
	}
}
