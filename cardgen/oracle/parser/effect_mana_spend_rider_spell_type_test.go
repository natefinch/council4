package parser

import "testing"

// spellTypeManaSpendRiderAbility wraps the rider sentence in a colorless mana
// ability so the parser processes it in an activated-ability context like the
// real cards (Vodalian Arcanist, Nardole, Pillar of the Paruns).
func spellTypeManaSpendRiderAbility(rider string) string {
	return "{T}: Add {C}. " + rider
}

// TestParseSpellTypeManaSpendRider verifies that each recognized spell-type
// restriction phrasing collapses into a single typed EffectManaSpendRider with
// the expected closed condition and the Restricted flag set.
func TestParseSpellTypeManaSpendRider(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		rider     string
		condition ManaSpendConditionKind
	}{
		{"instant or sorcery singular", "Spend this mana only to cast an instant or sorcery spell.", ManaSpendCastInstantOrSorcerySpell},
		{"instant and/or sorcery plural", "Spend this mana only to cast instant and/or sorcery spells.", ManaSpendCastInstantOrSorcerySpell},
		{"instant and sorcery plural", "Spend this mana only to cast instant and sorcery spells.", ManaSpendCastInstantOrSorcerySpell},
		{"noncreature singular", "Spend this mana only to cast a noncreature spell.", ManaSpendCastNoncreatureSpell},
		{"noncreature plural", "Spend this mana only to cast noncreature spells.", ManaSpendCastNoncreatureSpell},
		{"creature plural", "Spend this mana only to cast creature spells.", ManaSpendCastCreatureSpell},
		{"multicolored singular", "Spend this mana only to cast a multicolored spell.", ManaSpendCastMulticoloredSpell},
		{"multicolored plural", "Spend this mana only to cast multicolored spells.", ManaSpendCastMulticoloredSpell},
		{"planeswalker singular", "Spend this mana only to cast a planeswalker spell.", ManaSpendCastPlaneswalkerSpell},
		{"planeswalker plural", "Spend this mana only to cast planeswalker spells.", ManaSpendCastPlaneswalkerSpell},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			effect := riderEffect(t, spellTypeManaSpendRiderAbility(tc.rider))
			if effect == nil || effect.ManaSpendRider == nil {
				t.Fatalf("rider sentence did not collapse: %q", tc.rider)
			}
			if effect.ManaSpendRider.Condition != tc.condition {
				t.Fatalf("Condition = %q, want %q", effect.ManaSpendRider.Condition, tc.condition)
			}
			if !effect.ManaSpendRider.Restricted {
				t.Fatalf("Restricted = false, want true for %q", tc.rider)
			}
			if effect.ManaSpendRider.Effect != ManaSpendRiderEffectUnknown {
				t.Fatalf("Effect = %q, want unknown", effect.ManaSpendRider.Effect)
			}
		})
	}
}

// TestParseSpellTypeManaSpendRiderFailsClosed verifies that unmodeled spell-type
// selectors, extra qualifiers, or trailing content are not recognized, so they
// keep their lossless generic effects and fail closed downstream.
func TestParseSpellTypeManaSpendRiderFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Spend this mana only to cast a spell.",
		"Spend this mana only to cast spells.",
		"Spend this mana only to cast an artifact spell.",
		"Spend this mana only to cast a colorless spell.",
		"Spend this mana only to cast a creature spell with power 4 or greater.",
		"Spend this mana only to cast a multicolored creature spell.",
		"Spend this mana only to cast an instant spell.",
		"Spend this mana only to cast an instant or creature spell.",
		"Spend this mana only to cast a planeswalker spell you own.",
		"Spend this mana only to cast a noncreature spell or activate an ability.",
	}
	for _, rider := range rejected {
		effect := riderEffect(t, spellTypeManaSpendRiderAbility(rider))
		if effect != nil && spellTypeCondition(effect.ManaSpendRider) {
			t.Fatalf("unexpectedly recognized spell-type rider: %q", rider)
		}
	}
}

func spellTypeCondition(rider *ManaSpendRiderSyntax) bool {
	if rider == nil {
		return false
	}
	switch rider.Condition {
	case ManaSpendCastInstantOrSorcerySpell,
		ManaSpendCastNoncreatureSpell,
		ManaSpendCastMulticoloredSpell,
		ManaSpendCastPlaneswalkerSpell,
		ManaSpendCastCreatureSpell:
		return true
	default:
		return false
	}
}
