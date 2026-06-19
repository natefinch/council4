package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTapDownSpellSequence verifies that a single-target tap-down spell
// ("Tap target creature. It doesn't untap during its controller's next untap
// step.") lowers to a two-instruction sequence — tap the target, then skip its
// next untap — both bound to the single permanent target.
func TestLowerTapDownSpellSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tap Down",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Tap target creature. It doesn't untap during its controller's next untap step.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	tap, ok := mode.Sequence[0].Primitive.(game.Tap)
	if !ok || tap.Object != game.TargetPermanentReference(0) {
		t.Fatalf("first primitive = %+v, want tap of target 0", mode.Sequence[0].Primitive)
	}
	stun, ok := mode.Sequence[1].Primitive.(game.SkipNextUntap)
	if !ok || stun.Object != game.TargetPermanentReference(0) {
		t.Fatalf("second primitive = %+v, want SkipNextUntap of target 0", mode.Sequence[1].Primitive)
	}
}

// TestLowerTapDownTriggeredSequence verifies the tap-down sequence also lowers
// inside an enters-the-battlefield trigger (the Frost Lynx family), the most
// common printing of the pattern.
func TestLowerTapDownTriggeredSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Frost Beast",
		Layout:     "normal",
		TypeLine:   "Creature — Elemental",
		OracleText: "When Test Frost Beast enters, tap target creature an opponent controls. That creature doesn't untap during its controller's next untap step.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %+v, want one target and two instructions", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Tap); !ok {
		t.Fatalf("first primitive = %T, want game.Tap", mode.Sequence[0].Primitive)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.SkipNextUntap); !ok {
		t.Fatalf("second primitive = %T, want game.SkipNextUntap", mode.Sequence[1].Primitive)
	}
}

// TestLowerTapDownFailsClosed verifies tap-down shapes the SkipNextUntap
// primitive cannot model stay unsupported: a multi-step "next two untap steps"
// window and the plural "those creatures" form whose references do not bind to
// a single target.
func TestLowerTapDownFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Tap target creature. It doesn't untap during its controller's next two untap steps.",
		"Tap up to two target creatures. Those creatures don't untap during their controller's next untap step.",
	}
	for _, text := range rejected {
		faces, _ := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Reject",
			Layout:     "normal",
			TypeLine:   "Instant",
			OracleText: text,
		})
		for _, face := range faces {
			if face.SpellAbility.Exists {
				t.Errorf("OracleText %q lowered a spell ability, want fail closed", text)
			}
		}
	}
}
