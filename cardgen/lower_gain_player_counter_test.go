package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateGainEnergySource(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		oracle string
		kind   string
		amount string
	}{
		{"You get {E}{E} (two energy counters).", "counter.Energy", "game.Fixed(2)"},
		{"You get {E} (an energy counter).", "counter.Energy", "game.Fixed(1)"},
		{"You get {E}{E}{E}{E} (four energy counters).", "counter.Energy", "game.Fixed(4)"},
		{"You get an experience counter.", "counter.Experience", "game.Fixed(1)"},
		{"You get two experience counters.", "counter.Experience", "game.Fixed(2)"},
	} {
		t.Run(tc.oracle, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Counter Maker",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: tc.oracle,
			}, "c")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range []string{
				"game.AddPlayerCounter",
				tc.kind,
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

func TestGenerateGainPlayerCounterRecipients(t *testing.T) {
	t.Parallel()
	pt := "1"
	for _, tc := range []struct {
		name      string
		typeLine  string
		oracle    string
		reference string
		amount    string
	}{
		{
			name:      "Crypt Cobra",
			typeLine:  "Creature — Snake",
			oracle:    "Whenever this creature attacks and isn't blocked, defending player gets a poison counter.",
			reference: "game.DefendingPlayerReference()",
			amount:    "game.Fixed(1)",
		},
		{
			name:      "Pit Scorpion",
			typeLine:  "Creature — Scorpion",
			oracle:    "Whenever this creature deals damage to a player, that player gets a poison counter.",
			reference: "game.EventPlayerReference()",
			amount:    "game.Fixed(1)",
		},
		{
			name:      "Marsh Viper",
			typeLine:  "Creature — Snake",
			oracle:    "Whenever this creature deals damage to a player, that player gets two poison counters.",
			reference: "game.EventPlayerReference()",
			amount:    "game.Fixed(2)",
		},
		{
			name:      "Venerated Rotpriest",
			typeLine:  "Creature — Plant Warrior",
			oracle:    "Whenever a creature you control becomes the target of a spell, target opponent gets a poison counter.",
			reference: "game.TargetPlayerReference(0)",
			amount:    "game.Fixed(1)",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       tc.name,
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				OracleText: tc.oracle,
				ManaCost:   "{1}{B}",
				Power:      &pt,
				Toughness:  &pt,
			}, "p")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range []string{
				"game.AddPlayerCounter",
				"counter.Poison",
				tc.reference,
				tc.amount,
			} {
				if !strings.Contains(source, want) {
					t.Fatalf("generated source missing %q:\n%s", want, source)
				}
			}
			if strings.Contains(source, "game.ControllerReference()") {
				t.Fatalf("recipient wrongly lowered to controller:\n%s", source)
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
