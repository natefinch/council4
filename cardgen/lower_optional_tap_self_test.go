package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerOptionalUntapSelfTrigger proves a "you may tap/untap <this creature/it>"
// trigger body lowers to the mandatory referenced Tap/Untap primitive while the
// ability itself carries the optionality. The residual "you may untap it" clause
// is non-exact, so the tap/untap reference path tolerates the demotion and lowers
// the self/back-reference identically to its mandatory sibling. Nettle Sentinel
// untaps the source; the canonical "untap it" form untaps the triggering
// permanent.
func TestLowerOptionalUntapSelfTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Nettle Sentinel",
		Layout:   "normal",
		TypeLine: "Creature — Elf Warrior",
		OracleText: "This creature doesn't untap during your untap step.\n" +
			"Whenever you cast a green spell, you may untap this creature.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0]
	if !trigger.Optional {
		t.Fatal("trigger optional = false, want true")
	}
	mode := trigger.Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one untap", mode.Sequence)
	}
	untap, ok := mode.Sequence[0].Primitive.(game.Untap)
	if !ok || untap.Object != game.SourcePermanentReference() {
		t.Fatalf("sequence[0] = %#v, want Untap on source", mode.Sequence[0].Primitive)
	}
}

// TestLowerOptionalTapSelfGatedReflexive proves the gated "you may tap it. If you
// do, <reward>." trigger lowers: the self-reference Tap is marked Optional,
// publishes its result, and the reward is gated on the tap succeeding (Gitaxian
// Anatomist taps itself, then proliferates).
func TestLowerOptionalTapSelfGatedReflexive(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gitaxian Anatomist",
		Layout:     "normal",
		TypeLine:   "Creature — Phyrexian Wizard",
		OracleText: "When this creature enters, you may tap it. If you do, proliferate.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want tap then proliferate", mode.Sequence)
	}
	tap, ok := mode.Sequence[0].Primitive.(game.Tap)
	if !ok || tap.Object != game.EventPermanentReference() || !mode.Sequence[0].Optional {
		t.Fatalf("sequence[0] = %#v, want optional Tap on triggering permanent", mode.Sequence[0])
	}
	if _, ok := mode.Sequence[1].Primitive.(game.Proliferate); !ok {
		t.Fatalf("sequence[1] = %#v, want Proliferate", mode.Sequence[1].Primitive)
	}
}
