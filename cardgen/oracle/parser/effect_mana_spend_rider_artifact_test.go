package parser

import "testing"

// artifactManaSpendRiderAbility wraps the rider sentence in a colorless mana
// ability so the parser processes it in an activated-ability context like the
// real cards.
func artifactManaSpendRiderAbility(rider string) string {
	return "{T}: Add {C}. " + rider
}

// TestParseArtifactManaSpendRider verifies that each recognized artifact
// restriction phrasing collapses into a single typed EffectManaSpendRider with
// the expected closed condition and the Restricted flag set.
func TestParseArtifactManaSpendRider(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		rider     string
		condition ManaSpendConditionKind
	}{
		{"cast singular", "Spend this mana only to cast an artifact spell.", ManaSpendCastArtifactSpell},
		{"cast plural", "Spend this mana only to cast artifact spells.", ManaSpendCastArtifactSpell},
		{"cast or activate artifacts plural", "Spend this mana only to cast artifact spells or activate abilities of artifacts.", ManaSpendCastOrActivateArtifact},
		{"cast or activate artifact source", "Spend this mana only to cast an artifact spell or activate an ability of an artifact source.", ManaSpendCastOrActivateArtifact},
		{"cast or activate an artifact", "Spend this mana only to cast an artifact spell or activate an ability of an artifact.", ManaSpendCastOrActivateArtifact},
		{"activate only plural", "Spend this mana only to activate abilities of artifacts.", ManaSpendActivateArtifactAbility},
		{"activate only sources", "Spend this mana only to activate abilities of artifact sources.", ManaSpendActivateArtifactAbility},
		{"cast or activate any", "Spend this mana only to cast an artifact spell or activate an ability.", ManaSpendCastArtifactOrActivateAbility},
		{"activate any or cast", "Spend this mana only to activate an ability or cast an artifact spell.", ManaSpendCastArtifactOrActivateAbility},
		{"cast or to activate any", "Spend this mana only to cast an artifact spell or to activate an ability.", ManaSpendCastArtifactOrActivateAbility},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			effect := riderEffect(t, artifactManaSpendRiderAbility(tc.rider))
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

// TestParseArtifactManaSpendRiderFailsClosed verifies that non-artifact or
// malformed restrictions are not recognized as artifact riders, so they keep
// their lossless generic effects and fail closed downstream.
func TestParseArtifactManaSpendRiderFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Spend this mana only to cast a creature spell.",
		"Spend this mana only to cast an instant or sorcery spell.",
		"Spend this mana only to activate abilities.",
		"Spend this mana only to cast artifact or creature spells.",
		"Spend this mana only to cast an artifact spell or activate an ability of a creature.",
		"Spend this mana only to cast an artifact spell or cast an instant.",
		"Spend this mana only to activate abilities of artifacts or activate abilities of creatures.",
	}
	for _, rider := range rejected {
		effect := riderEffect(t, artifactManaSpendRiderAbility(rider))
		if effect != nil && artifactCondition(effect.ManaSpendRider) {
			t.Fatalf("unexpectedly recognized artifact rider: %q", rider)
		}
	}
}

func artifactCondition(rider *ManaSpendRiderSyntax) bool {
	if rider == nil {
		return false
	}
	switch rider.Condition {
	case ManaSpendCastArtifactSpell,
		ManaSpendCastOrActivateArtifact,
		ManaSpendActivateArtifactAbility,
		ManaSpendCastArtifactOrActivateAbility:
		return true
	default:
		return false
	}
}
