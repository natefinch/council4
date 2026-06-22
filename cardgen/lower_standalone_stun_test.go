package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerStandaloneStunSpell verifies that a standalone targeted stun spell
// ("Target creature doesn't untap during its controller's next untap step.")
// lowers to a single SkipNextUntap on the spell's one permanent target, with no
// preceding tap.
func TestLowerStandaloneStunSpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sleeper Dart",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature doesn't untap during its controller's next untap step.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %+v, want one target and one instruction", mode)
	}
	stun, ok := mode.Sequence[0].Primitive.(game.SkipNextUntap)
	if !ok || stun.Object != game.TargetPermanentReference(0) {
		t.Fatalf("primitive = %+v, want SkipNextUntap of target 0", mode.Sequence[0].Primitive)
	}
}

// TestLowerStandaloneStunActivated verifies the standalone stun also lowers as an
// activated ability ("{1}{U}, {T}: Target creature doesn't untap during its
// controller's next untap step.", House Guildmage), the most common printing of
// the bare stun.
func TestLowerStandaloneStunActivated(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Stun Mage",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard",
		OracleText: "{1}{U}, {T}: Target creature doesn't untap during its controller's next untap step.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %+v, want one target and one instruction", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.SkipNextUntap); !ok {
		t.Fatalf("primitive = %T, want game.SkipNextUntap", mode.Sequence[0].Primitive)
	}
}

// TestLowerStandaloneStunFailsClosed verifies stun shapes the single SkipNextUntap
// primitive cannot model stay unsupported: a multi-step "next two untap steps"
// window (which the parser splits into multiple effects) and the duration-bearing
// "until your next turn" variant both fail closed.
func TestLowerStandaloneStunFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"Target creature doesn't untap during its controller's next two untap steps.",
		"Target creature doesn't untap during its controller's untap step for as long as you control this enchantment.",
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
