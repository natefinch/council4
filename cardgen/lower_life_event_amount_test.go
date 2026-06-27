package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// triggeredGainLife returns the GainLife primitive of a single-face card's sole
// triggered ability, failing the test if the shape is anything else.
func triggeredGainLife(t *testing.T, face loweredFaceAbilities) game.GainLife {
	t.Helper()
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(sequence))
	}
	gain, ok := sequence[0].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.GainLife", sequence[0].Primitive)
	}
	return gain
}

// TestLowerGainThatMuchLifeFromDamageTrigger proves that "you gain that much
// life" resolves its "that much" anaphor against the enclosing damage trigger,
// reading the triggering damage rather than a life-change quantity. This is the
// "Whenever <X> deals/is dealt damage, you gain that much life" family (Mourning
// Thrull, Wall of Essence, Armadillo Cloak, ...): the parser pins every "that
// much life" phrase to a single life-change kind, so the lowering must resolve
// it on whichever event actually fired.
func TestLowerGainThatMuchLifeFromDamageTrigger(t *testing.T) {
	t.Parallel()
	power, toughness := "1", "1"
	for _, tc := range []struct {
		name       string
		oracleText string
	}{
		{
			name:       "deals damage",
			oracleText: "Whenever this creature deals damage, you gain that much life.",
		},
		{
			name:       "is dealt combat damage",
			oracleText: "Whenever this creature is dealt combat damage, you gain that much life.",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Drinker",
				Layout:     "normal",
				ManaCost:   "{1}{B}",
				TypeLine:   "Creature — Vampire",
				OracleText: tc.oracleText,
				Power:      &power,
				Toughness:  &toughness,
			})
			if face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventDamageDealt {
				t.Fatalf("event = %v, want EventDamageDealt", face.TriggeredAbilities[0].Trigger.Pattern.Event)
			}
			gain := triggeredGainLife(t, face)
			if gain.Player != game.ControllerReference() {
				t.Fatalf("player = %#v, want controller", gain.Player)
			}
			dynamic := gain.Amount.DynamicAmount()
			if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventDamage {
				t.Fatalf("amount = %#v, want dynamic event damage", gain.Amount)
			}
		})
	}
}

// TestLowerGainThatMuchLifeFromLifeGainTriggerUnchanged guards the pre-existing
// life-change-trigger reading: inside a life-gain trigger, "that much life"
// still resolves to the triggering life change, not the damage path.
func TestLowerGainThatMuchLifeFromLifeGainTriggerUnchanged(t *testing.T) {
	t.Parallel()
	power, toughness := "2", "2"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mirror",
		Layout:     "normal",
		ManaCost:   "{2}{W}",
		TypeLine:   "Creature — Spirit",
		OracleText: "Whenever an opponent gains life, you gain that much life.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if face.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventLifeGained {
		t.Fatalf("event = %v, want EventLifeGained", face.TriggeredAbilities[0].Trigger.Pattern.Event)
	}
	gain := triggeredGainLife(t, face)
	dynamic := gain.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountEventLifeChange {
		t.Fatalf("amount = %#v, want dynamic event life change", gain.Amount)
	}
}

// TestLowerGainThatMuchLifeOutsideTriggerFailsClosed proves the "that much life"
// anaphor stays rejected with no enclosing event quantity to read, so a spell
// that never defines "that much" is never silently mislowered.
func TestLowerGainThatMuchLifeOutsideTriggerFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Charm",
		Layout:     "normal",
		ManaCost:   "{W}",
		TypeLine:   "Sorcery",
		OracleText: "You gain that much life.",
	})
}
