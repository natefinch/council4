package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceDividedDamage asserts that "deals N damage
// divided as you choose among <cardinality> <targets>" with a fixed or variable
// X total lowers to a single multi-target spec and a Divided Damage instruction
// whose maximum target count is capped at a fixed total and left at the wording's
// bound for a variable X.
func TestGenerateExecutableCardSourceDividedDamage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		oracleText  string
		typeLine    string
		wantedSnips []string
	}{
		{
			name:       "any number of targets",
			oracleText: "Test Bolt deals 3 damage divided as you choose among any number of targets.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 3",
				"Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer",
				"Amount:    game.Fixed(3)",
				"Recipient: game.AnyTargetDamageRecipient(0)",
				"Divided:   true",
			},
		},
		{
			name:       "variable X among any number of targets",
			oracleText: "Test Bolt deals X damage divided as you choose among any number of targets.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 99",
				"Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer",
				"Kind: game.DynamicAmountX",
				"Recipient: game.AnyTargetDamageRecipient(0)",
				"Divided:   true",
			},
		},
		{
			name:       "variable X among any number of target creatures",
			oracleText: "Test Bolt deals X damage divided as you choose among any number of target creatures.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 99",
				"RequiredTypesAny: []types.Card{types.Creature}",
				"Kind: game.DynamicAmountX",
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
				"RequiredTypesAny: []types.Card{types.Creature}",
				"Divided:   true",
			},
		},
		{
			name:       "up to three target creatures",
			oracleText: "Test Bolt deals 3 damage divided as you choose among up to three target creatures.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 3",
				"RequiredTypesAny: []types.Card{types.Creature}",
				"Divided:   true",
			},
		},
		{
			name:       "keyword filtered creatures",
			oracleText: "Test Bolt deals 4 damage divided as you choose among any number of target creatures with flying.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 4",
				"RequiredTypesAny: []types.Card{types.Creature}",
				"Keyword: game.Flying",
				"Divided:   true",
			},
		},
		{
			name:       "combat state filtered creatures",
			oracleText: "Test Bolt deals 5 damage divided as you choose among any number of target attacking creatures.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 5",
				"RequiredTypesAny: []types.Card{types.Creature}",
				"CombatState: game.CombatStateAttacking",
				"Divided:   true",
			},
		},
		{
			name:       "combat state without keyword",
			oracleText: "Test Bolt deals X damage divided as you choose among any number of target attacking or blocking creatures without flying.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 99",
				"RequiredTypesAny: []types.Card{types.Creature}",
				"CombatState: game.CombatStateAttackingOrBlocking",
				"ExcludedKeyword: game.Flying",
				"Divided:   true",
			},
		},
		{
			name:       "type union",
			oracleText: "Test Bolt deals 5 damage divided as you choose among any number of target creatures and/or planeswalkers.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 5",
				"RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}",
				"Divided:   true",
			},
		},
		{
			name:       "controller filtered union",
			oracleText: "Test Bolt deals 5 damage divided as you choose among any number of target creatures and/or planeswalkers your opponents control.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 5",
				"RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}",
				"game.ControllerOpponent",
				"Divided:   true",
			},
		},
		{
			name:       "color filtered creatures",
			oracleText: "Test Bolt deals 3 damage divided as you choose among one, two, or three target white and/or blue creatures.",
			wantedSnips: []string{
				"MinTargets: 1",
				"MaxTargets: 3",
				"RequiredTypesAny: []types.Card{types.Creature}",
				"[]color.Color{color.White, color.Blue}",
				"Divided:   true",
			},
		},
		{
			name:       "dynamic count total",
			oracleText: "Test Bolt deals X damage divided as you choose among any number of target creatures and/or planeswalkers your opponents control, where X is the number of lands you control.",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 99",
				"RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}",
				"game.ControllerOpponent",
				"game.DynamicAmountCountSelector",
				"Divided:   true",
			},
		},
		{
			name:       "dynamic source counter total",
			oracleText: "At the beginning of your end step, this enchantment deals X damage divided as you choose among any number of target creatures, where X is the number of age counters on it.",
			typeLine:   "Enchantment",
			wantedSnips: []string{
				"MinTargets: 0",
				"MaxTargets: 99",
				"RequiredTypesAny: []types.Card{types.Creature}",
				"game.DynamicAmountObjectCounters",
				"counter.Age",
				"Divided:      true",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			typeLine := test.typeLine
			if typeLine == "" {
				typeLine = "Instant"
			}
			card := &ScryfallCard{
				Name:       "Test Bolt",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   typeLine,
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
			name:       "variable total with plus rider",
			oracleText: "Test Bolt deals X plus 1 damage divided as you choose among any number of targets.",
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
