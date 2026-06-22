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

// TestLowerSourceStunManaAbility verifies the self-source stun "This land
// doesn't untap during your next untap step." (the dual lands Mogg Hollows /
// Rootwater Depths) lowers as a mana-ability rider: the mana ability emits the
// AddMana instruction followed by a SkipNextUntap on the source land itself.
func TestLowerSourceStunManaAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mogg Hollows",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{T}: Add {R} or {G}. This land doesn't untap during your next untap step.",
	})
	if len(face.ManaAbilities) != 2 {
		t.Fatalf("got %d mana abilities, want 2", len(face.ManaAbilities))
	}
	mode := face.ManaAbilities[1].Content.Modes[0]
	if len(mode.Sequence) < 2 {
		t.Fatalf("rider mana ability sequence = %+v, want add-mana then stun", mode.Sequence)
	}
	last := mode.Sequence[len(mode.Sequence)-1]
	stun, ok := last.Primitive.(game.SkipNextUntap)
	if !ok || stun.Object != game.SourcePermanentReference() {
		t.Fatalf("primitive = %+v, want SkipNextUntap of source", last.Primitive)
	}
}

// TestLowerSourceStunActivated verifies the self-source stun also lowers as the
// trailing clause of a non-mana activated ability ("{2}{W}, {T}: This creature
// deals 3 damage to target attacking or blocking creature. This creature doesn't
// untap during your next untap step.", Arbalest Elite): the SkipNextUntap
// addresses the source and follows the damage instruction.
func TestLowerSourceStunActivated(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Arbalest Elite",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		OracleText: "{2}{W}, {T}: This creature deals 3 damage to target attacking or blocking creature. This creature doesn't untap during your next untap step.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("got %d activated abilities, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %+v, want two instructions", mode.Sequence)
	}
	stun, ok := mode.Sequence[1].Primitive.(game.SkipNextUntap)
	if !ok || stun.Object != game.SourcePermanentReference() {
		t.Fatalf("primitive = %+v, want SkipNextUntap of source", mode.Sequence[1].Primitive)
	}
}

// TestLowerSourceStunFailsClosed verifies self-source stun wordings the single
// source SkipNextUntap cannot model stay unsupported: a duration-bearing "for as
// long as" variant and a "your next two untap steps" multi-step window both fail
// closed rather than dropping the rider.
func TestLowerSourceStunFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"{T}: Add {C}. This land doesn't untap during your next two untap steps.",
		"{T}: Add {C}. This land doesn't untap during your untap step for as long as you control a Mountain.",
	}
	for _, text := range rejected {
		faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
			Name:       "Test Reject Source Stun",
			Layout:     "normal",
			TypeLine:   "Land",
			OracleText: text,
		})
		if len(diagnostics) == 0 {
			t.Errorf("OracleText %q lowered with no diagnostics, want fail closed", text)
		}
		for _, face := range faces {
			for _, mana := range face.ManaAbilities {
				for _, mode := range mana.Content.Modes {
					if len(mode.Sequence) > 1 {
						t.Errorf("OracleText %q lowered a stun rider, want fail closed", text)
					}
				}
			}
		}
	}
}
