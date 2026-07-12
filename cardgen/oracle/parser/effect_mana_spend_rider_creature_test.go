package parser

import "testing"

// TestParseCreatureCastOrActivateManaSpendRider verifies that each accepted
// "cast creature spells or activate abilities of creatures" phrasing collapses
// into a single typed EffectManaSpendRider with the ManaSpendCastOrActivateCreature
// condition and the Restricted flag set (Castle Garenbrig).
func TestParseCreatureCastOrActivateManaSpendRider(t *testing.T) {
	t.Parallel()
	accepted := []string{
		"Spend this mana only to cast creature spells or activate abilities of creatures.",
		"Spend this mana only to cast a creature spell or activate an ability of a creature.",
		"Spend this mana only to cast creature spells or activate abilities of creature sources.",
		"Spend this mana only to cast a creature spell or activate an ability of a creature source.",
	}
	for _, rider := range accepted {
		effect := riderEffect(t, spellTypeManaSpendRiderAbility(rider))
		if effect == nil || effect.ManaSpendRider == nil {
			t.Fatalf("rider sentence did not collapse: %q", rider)
		}
		if effect.ManaSpendRider.Condition != ManaSpendCastOrActivateCreature {
			t.Fatalf("Condition = %q, want %q for %q", effect.ManaSpendRider.Condition, ManaSpendCastOrActivateCreature, rider)
		}
		if !effect.ManaSpendRider.Restricted {
			t.Fatalf("Restricted = false, want true for %q", rider)
		}
		if effect.ManaSpendRider.Effect != ManaSpendRiderEffectUnknown {
			t.Fatalf("Effect = %q, want unknown for %q", effect.ManaSpendRider.Effect, rider)
		}
	}
}

// TestParseCreatureCastOrActivateManaSpendRiderFailsClosed verifies that near
// misses — a bare creature-spell restriction, an artifact variant, the
// chosen-type variant, or an unmodeled activation clause — are not recognized as
// the creature cast-or-activate rider, so unmodeled wording fails closed.
func TestParseCreatureCastOrActivateManaSpendRiderFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Spend this mana only to cast creature spells.",
		"Spend this mana only to cast creature spells or activate abilities.",
		"Spend this mana only to cast creature spells or activate abilities of artifacts.",
		"Spend this mana only to cast artifact spells or activate abilities of creatures.",
		"Spend this mana only to cast a creature spell of the chosen type or activate an ability of a creature source of the chosen type.",
		"Spend this mana only to cast creature spells or activate an ability of a creature and draw a card.",
	}
	for _, rider := range rejected {
		effect := riderEffect(t, spellTypeManaSpendRiderAbility(rider))
		if effect != nil && effect.ManaSpendRider != nil &&
			effect.ManaSpendRider.Condition == ManaSpendCastOrActivateCreature {
			t.Fatalf("unexpectedly recognized creature cast-or-activate rider: %q", rider)
		}
	}
}
