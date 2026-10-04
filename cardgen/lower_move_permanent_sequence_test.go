package cardgen

import (
	"fmt"
	"testing"
)

func TestMovePermanentLibrarySequenceCardDef(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		oracle string
		slot   int
		target int
		bottom bool
		gated  bool
	}{
		{"bottom then life", "Put target creature on the bottom of its owner's library. You lose 2 life.", 0, 0, true, false},
		{"top then draw", "Put target creature on top of its owner's library. Draw a card.", 0, 0, false, false},
		{"draw then bottom", "Draw a card. Put target creature on the bottom of its owner's library.", 1, 0, true, false},
		{"earlier player", "Target player gains 1 life. Put target creature on top of its owner's library.", 1, 1, false, false},
		{"earlier permanent", "Tap target artifact. Put target creature on the bottom of its owner's library.", 1, 1, true, false},
		{"conditional move", "Target player gains 1 life. If you control three or more artifacts, put target creature on top of its owner's library.", 1, 1, false, true},
		{"conditional rider", "Put target creature on the bottom of its owner's library. If you control three or more artifacts, you lose 2 life.", 0, 0, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: "Test Library Sequence", Layout: "normal", TypeLine: "Sorcery", OracleText: tc.oracle}
			movePath := fmt.Sprintf("Sequence[%d].Primitive.(game.MovePermanent)", tc.slot)
			assertCardPaths(t, card,
				movePath+".Destination = zone.Library",
				movePath+".Object.kind = game.ObjectReferenceTargetPermanent",
			)
			if tc.target > 0 {
				assertCardPaths(t, card, fmt.Sprintf("%s.Object.targetIndex = %d", movePath, tc.target))
			} else {
				assertCardPathsAbsent(t, card, movePath+".Object.targetIndex")
			}
			if tc.bottom {
				assertCardPaths(t, card, movePath+".LibraryBottom = true")
			} else {
				assertCardPathsAbsent(t, card, movePath+".LibraryBottom")
			}
			if tc.gated {
				assertCardPaths(t, card, fmt.Sprintf("Sequence[%d].Condition.Exists = true", tc.slot))
			}
			assertCardPathsAbsent(t, card, "Sequence[2]", "Modes[1]", movePath+".ControlledChoice", movePath+".Group.domain")
		})
	}
}

func TestMovePermanentLibrarySequenceRefusals(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Put a creature you control on top of its owner's library. Draw a card.",
		"Put all creatures on the bottom of their owners' libraries. You lose 2 life.",
		"Put target creature into its owner's library third from the top. Draw a card.",
		"Put target creature on the bottom of its controller's library. Draw a card.",
		"Tap target creature. Put that creature on the bottom of its owner's library.",
		"Put target creature on the bottom of its owner's library. You lose 2 life unless you control a Villain.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Test Library Near Miss", Layout: "normal", TypeLine: "Sorcery", OracleText: oracle})
		})
	}
}
