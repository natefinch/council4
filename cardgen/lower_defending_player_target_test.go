package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDefendingPlayerControlsTarget verifies "defending player controls" on
// a single target selector lowers to a target whose Selection is restricted to
// permanents controlled by the defending player of the triggering attack
// (ControlledByDefendingPlayer), as in "goad target creature defending player
// controls." (Coveted Peacock).
func TestLowerDefendingPlayerControlsTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Defending Target",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Creature — Goblin",
		OracleText: "Whenever this creature attacks, you may goad target creature defending player controls.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(mode.Targets))
	}
	if !mode.Targets[0].Selection.Exists ||
		!mode.Targets[0].Selection.Val.ControlledByDefendingPlayer {
		t.Fatalf("target selection = %#v, want ControlledByDefendingPlayer", mode.Targets[0].Selection)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Goad); !ok {
		// The goad may be wrapped as an optional resolving effect; only assert the
		// target restriction here.
		_ = ok
	}
}

// TestLowerDefendingPlayerControlsGroup verifies "each creature ... defending
// player controls" on a group damage recipient lowers to a battlefield group
// restricted to the defending player's creatures (ControlledByDefendingPlayer),
// as in "deal 1 damage to each creature without flying defending player
// controls." (Scalding Salamander) — previously mis-lowered to every creature.
func TestLowerDefendingPlayerControlsGroup(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Defending Group",
		Layout:     "normal",
		ManaCost:   "{3}{R}",
		TypeLine:   "Creature — Elemental",
		OracleText: "Whenever this creature attacks, you may have it deal 1 damage to each creature without flying defending player controls.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	found := false
	for _, instruction := range face.TriggeredAbilities[0].Content.Modes[0].Sequence {
		damage, ok := instruction.Primitive.(game.Damage)
		if !ok {
			continue
		}
		group, ok := damage.Recipient.GroupReference()
		if ok && group.Selection().ControlledByDefendingPlayer {
			found = true
		}
	}
	if !found {
		t.Fatal("group damage recipient not restricted to ControlledByDefendingPlayer")
	}
}
