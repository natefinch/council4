package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerCommandBeaconMovesCommanderToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Command Beacon",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{T}, Sacrifice this land: Put your commander into your hand from the command zone.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1 (the sacrifice ability)", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("ability content modes = %#v, want a single-instruction mode", ability.Content.Modes)
	}
	move, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.MoveCommander)
	if !ok {
		t.Fatalf("primitive = %T, want game.MoveCommander", ability.Content.Modes[0].Sequence[0].Primitive)
	}
	if move.Player.Kind() != game.PlayerReferenceController {
		t.Fatalf("Player = %#v, want controller reference", move.Player)
	}
	if move.Destination != zone.Hand {
		t.Fatalf("Destination = %v, want Hand", move.Destination)
	}
}

func TestLowerMoveCommanderFailsClosedForBattlefieldDestination(t *testing.T) {
	t.Parallel()
	// "onto the battlefield" is a distinct reanimation-style shape with extra
	// control nuances; the command-zone-to-hand lowerer must not claim it.
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Beacon",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}, Sacrifice this land: Put your commander onto the battlefield from the command zone.",
	})
}
