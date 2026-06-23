package parser

import "testing"

// TestParseSpellCostModifierEffect proves the resolving, duration-bounded spell
// cost modifier is recognized across its caster scopes ("you cast", "your
// opponents cast"), its increase/reduction wordings, an optional single
// card-type filter ("Artifact spells", Armor Wars chapter II) or excluded filter
// ("Noncreature spells", Elspeth Conquers Death chapter II), and leading,
// medial, or trailing duration phrases, while permanent statics and malformed
// wordings fail closed and flow through the generic effect parser.
func TestParseSpellCostModifierEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source       string
		recognized   bool
		caster       SpellCostCasterKind
		amount       int
		increase     bool
		duration     EffectDurationKind
		requiredType []CardType
		excludedType []CardType
	}{
		{
			source: "Artifact spells you cast this turn cost {1} less to cast.", recognized: true,
			caster: SpellCostCasterController, amount: 1, increase: false,
			duration: EffectDurationThisTurn, requiredType: []CardType{CardTypeArtifact},
		},
		{
			source: "Until your next turn, spells your opponents cast cost {1} more to cast.", recognized: true,
			caster: SpellCostCasterOpponents, amount: 1, increase: true,
			duration: EffectDurationUntilYourNextTurn,
		},
		{
			source: "Noncreature spells your opponents cast cost {2} more to cast until your next turn.", recognized: true,
			caster: SpellCostCasterOpponents, amount: 2, increase: true,
			duration: EffectDurationUntilYourNextTurn, excludedType: []CardType{CardTypeCreature},
		},
		// Permanent statics carry no duration and fail closed.
		{source: "Artifact spells you cast cost {1} less to cast.", recognized: false},
		// Malformed wordings fail closed.
		{source: "Spells your opponents cast cost {0} more to cast this turn.", recognized: false},
		{source: "Spells your opponents cast cost {X} more to cast this turn.", recognized: false},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			got := len(effects) == 1 && effects[0].Kind == EffectSpellCostModifier
			if got != test.recognized {
				t.Fatalf("recognized = %v, want %v (effects=%#v)", got, test.recognized, effects)
			}
			if !got {
				return
			}
			effect := effects[0]
			if effect.SpellCostModifierCaster != test.caster {
				t.Fatalf("caster = %v, want %v", effect.SpellCostModifierCaster, test.caster)
			}
			if effect.SpellCostModifierAmount != test.amount {
				t.Fatalf("amount = %d, want %d", effect.SpellCostModifierAmount, test.amount)
			}
			if effect.SpellCostModifierIncrease != test.increase {
				t.Fatalf("increase = %v, want %v", effect.SpellCostModifierIncrease, test.increase)
			}
			if effect.Duration != test.duration {
				t.Fatalf("duration = %v, want %v", effect.Duration, test.duration)
			}
			if !cardTypeSlicesEqual(effect.SpellCostModifierRequiredTypes, test.requiredType) {
				t.Fatalf("requiredTypes = %v, want %v", effect.SpellCostModifierRequiredTypes, test.requiredType)
			}
			if !cardTypeSlicesEqual(effect.SpellCostModifierExcludedTypes, test.excludedType) {
				t.Fatalf("excludedTypes = %v, want %v", effect.SpellCostModifierExcludedTypes, test.excludedType)
			}
		})
	}
}
