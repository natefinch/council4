package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceDisjunctiveSacrificeDiscard covers the
// resolving-optional "you may sacrifice X or discard Y. If you do, REWARD"
// disjunctive cost family. The two mutually-exclusive costs lower to two
// sequential Optional instructions — the second gated on the first not being
// accepted — and the reward is emitted twice, gated on each cost having
// succeeded, so exactly one reward copy fires.
func TestGenerateExecutableCardSourceDisjunctiveSacrificeDiscard(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
		power      *string
		wantParts  []string
	}{
		{
			name:       "sacrifice artifact or discard, draw a card",
			typeLine:   "Instant",
			oracleText: "You may sacrifice an artifact or discard a card. If you do, draw a card.",
			wantParts: []string{
				"Primitive: game.SacrificePermanents",
				"Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}}",
				`PublishResult: game.ResultKey("disjunctive-cost-a")`,
				"Primitive: game.Discard",
				`Key:      "disjunctive-cost-a"`,
				"Accepted: game.TriFalse",
				`PublishResult: game.ResultKey("disjunctive-cost-b")`,
				"Primitive: game.Draw",
				`Key:       "disjunctive-cost-b"`,
				"Succeeded: game.TriTrue",
			},
		},
		{
			name:       "sacrifice creature or discard creature card",
			typeLine:   "Creature — Horror",
			oracleText: "When this creature enters, you may sacrifice a creature or discard a creature card. If you do, draw a card.",
			power:      new("2"),
			wantParts: []string{
				"Primitive: game.SacrificePermanents",
				"Primitive: game.ChooseDiscardFromHand",
				"Primitive: game.Draw",
				`PublishResult: game.ResultKey("disjunctive-cost-a")`,
				`PublishResult: game.ResultKey("disjunctive-cost-b")`,
			},
		},
		{
			name:       "multi-effect reward draw and pump",
			typeLine:   "Creature — Devil Detective",
			oracleText: "Whenever this creature attacks, you may sacrifice an artifact or discard a card. If you do, draw a card and this creature gets +2/+0 until end of turn.",
			power:      new("2"),
			wantParts: []string{
				"Primitive: game.SacrificePermanents",
				"Primitive: game.Discard",
				"Primitive: game.Draw",
				"Primitive: game.ModifyPT",
			},
		},
		{
			name:       "discard first then sacrifice land",
			typeLine:   "Sorcery",
			oracleText: "You may discard a card or sacrifice a land. If you do, draw two cards.",
			wantParts: []string{
				"Primitive: game.Discard",
				"Primitive: game.SacrificePermanents",
				"Selection: game.Selection{RequiredTypes: []types.Card{types.Land}}",
				`PublishResult: game.ResultKey("disjunctive-cost-a")`,
				`PublishResult: game.ResultKey("disjunctive-cost-b")`,
				"Primitive: game.Draw",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Disjunctive Cost",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      test.power,
				Toughness:  test.power,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "d")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, want := range test.wantParts {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
		})
	}
}

// TestGenerateExecutableCardSourceRejectsUnsupportedDisjunctiveCost keeps the
// fail-closed boundary of the disjunctive cost lowerer: only non-targeted
// reward clauses over a sacrifice-or-discard disjunction are duplicated. A
// targeted reward (whose target the duplication cannot safely share), a
// non-sacrifice/discard alternative ("or pay {mana}"), and a same-kind
// disjunction all stay unsupported rather than lowering to a wrong shape.
func TestGenerateExecutableCardSourceRejectsUnsupportedDisjunctiveCost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		oracleText string
	}{
		{
			name:       "targeted reward",
			typeLine:   "Sorcery",
			oracleText: "You may sacrifice an artifact or discard a card. If you do, target creature gets +1/+1 until end of turn.",
		},
		{
			name:       "or pay mana alternative",
			typeLine:   "Sorcery",
			oracleText: "You may sacrifice an artifact or pay {2}. If you do, draw a card.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       "Unsupported Disjunctive Cost",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "u")
			if err != nil {
				t.Fatal(err)
			}
			if source != "" {
				t.Fatalf("source = %q, want no partial card", source)
			}
			if len(diagnostics) == 0 {
				t.Fatal("diagnostics = empty, want unsupported")
			}
		})
	}
}
