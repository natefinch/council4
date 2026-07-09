package parser

import (
	"slices"
	"testing"
)

// firstEffectWithAmount parses source and returns the first effect carrying a
// dynamic amount, used by the "that player controls" damage count-subject tests.
func firstEffectWithAmount(t *testing.T, source string) *EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for ai := range document.Abilities {
		ability := &document.Abilities[ai]
		for si := range ability.Sentences {
			sentence := &ability.Sentences[si]
			for ei := range sentence.Effects {
				if sentence.Effects[ei].Amount.DynamicKind != EffectDynamicAmountNone {
					return &sentence.Effects[ei]
				}
			}
		}
	}
	return nil
}

// TestParseThatPlayerControlsDamageCountSubject covers the "deals damage ...
// equal to the number of <filtered permanents> that player controls" count
// subject shared by Anathemancer and Jovial Evil: the count is scoped to the
// damaged player through SelectionControllerThatPlayer and carries the noun
// phrase's type/supertype/color filters.
func TestParseThatPlayerControlsDamageCountSubject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		source       string
		multiplier   int
		requireTypes []CardType
		excluded     []Supertype
		colorsAny    []Color
	}{
		{
			name:         "anathemancer nonbasic lands",
			source:       "Anathemancer deals damage to target player equal to the number of nonbasic lands that player controls.",
			multiplier:   1,
			requireTypes: []CardType{CardTypeLand},
			excluded:     []Supertype{SupertypeBasic},
		},
		{
			name:         "jovial evil twice white creatures",
			source:       "Jovial Evil deals X damage to target opponent, where X is twice the number of white creatures that player controls.",
			multiplier:   2,
			requireTypes: []CardType{CardTypeCreature},
			colorsAny:    []Color{ColorWhite},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			effect := firstEffectWithAmount(t, test.source)
			if effect == nil {
				t.Fatalf("source %q yielded no dynamic amount", test.source)
			}
			if effect.Kind != EffectDealDamage {
				t.Fatalf("effect.Kind = %v, want EffectDealDamage", effect.Kind)
			}
			amount := effect.Amount
			if amount.DynamicKind != EffectDynamicAmountCount {
				t.Fatalf("DynamicKind = %v, want EffectDynamicAmountCount", amount.DynamicKind)
			}
			if amount.Multiplier != test.multiplier {
				t.Fatalf("Multiplier = %d, want %d", amount.Multiplier, test.multiplier)
			}
			if amount.Selection == nil {
				t.Fatal("Selection is nil")
			}
			sel := amount.Selection
			if sel.Controller != SelectionControllerThatPlayer {
				t.Fatalf("Controller = %v, want SelectionControllerThatPlayer", sel.Controller)
			}
			gotTypes := slices.Clone(sel.RequiredTypesAny)
			slices.Sort(gotTypes)
			wantTypes := slices.Clone(test.requireTypes)
			slices.Sort(wantTypes)
			if !slices.Equal(gotTypes, wantTypes) {
				t.Fatalf("RequiredTypesAny = %v, want %v", sel.RequiredTypesAny, test.requireTypes)
			}
			if !slices.Equal(sel.ExcludedSupertypes, test.excluded) {
				t.Fatalf("ExcludedSupertypes = %v, want %v", sel.ExcludedSupertypes, test.excluded)
			}
			if !slices.Equal(sel.ColorsAny, test.colorsAny) {
				t.Fatalf("ColorsAny = %v, want %v", sel.ColorsAny, test.colorsAny)
			}
		})
	}
}
