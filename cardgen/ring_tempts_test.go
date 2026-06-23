package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// abilityHasRingTempts reports whether any mode sequence of the ability content
// contains a RingTempts primitive scoped to the resolving controller.
func abilityHasRingTempts(t *testing.T, content game.AbilityContent) bool {
	t.Helper()
	for _, mode := range content.Modes {
		for _, instr := range mode.Sequence {
			prim, ok := instr.Primitive.(game.RingTempts)
			if !ok {
				continue
			}
			if prim.Player.Kind() != game.PlayerReferenceController {
				t.Fatalf("RingTempts player = %v, want controller", prim.Player.Kind())
			}
			return true
		}
	}
	return false
}

// TestGenerateBirthdayEscapeRingTempts proves the standalone designation effect
// "The Ring tempts you." lowers to a RingTempts primitive after a fixed draw.
func TestGenerateBirthdayEscapeRingTempts(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Birthday Escape",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card. The Ring tempts you.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability was not lowered")
	}
	if !abilityHasRingTempts(t, face.SpellAbility.Val) {
		t.Fatalf("spell ability has no RingTempts primitive: %#v", face.SpellAbility.Val)
	}
}

// TestGenerateClaimThePreciousRingTempts proves a removal spell whose trailing
// sentence is "The Ring tempts you." lowers the ring designation alongside the
// destroy effect.
func TestGenerateClaimThePreciousRingTempts(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Claim the Precious",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target creature. The Ring tempts you.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability was not lowered")
	}
	if !abilityHasRingTempts(t, face.SpellAbility.Val) {
		t.Fatalf("spell ability has no RingTempts primitive: %#v", face.SpellAbility.Val)
	}
}

// TestGenerateWarOfTheLastAllianceRingTempts proves the target Saga generates
// and its chapter III lowers the trailing "The Ring tempts you." sub-effect to a
// RingTempts primitive after the temporary double-strike keyword grant.
func TestGenerateWarOfTheLastAllianceRingTempts(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "War of the Last Alliance",
		Layout:   "saga",
		TypeLine: "Enchantment — Saga",
		OracleText: "(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)\n" +
			"I, II — Search your library for a legendary creature card, reveal it, put it into your hand, then shuffle.\n" +
			"III — Creatures you control gain double strike until end of turn. The Ring tempts you.",
	})
	if len(face.ChapterAbilities) == 0 {
		t.Fatal("no chapter abilities lowered")
	}
	chapterIII := face.ChapterAbilities[len(face.ChapterAbilities)-1]
	if !abilityHasRingTempts(t, chapterIII.Content) {
		t.Fatalf("chapter III has no RingTempts primitive: %#v", chapterIII.Content)
	}
}
