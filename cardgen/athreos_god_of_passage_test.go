package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerTargetOpponentPayLifeUnlessEventCardReturn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Athreos, God of Passage",
		Layout:   "normal",
		TypeLine: "Legendary Enchantment Creature — God",
		OracleText: "Indestructible\n" +
			"As long as your devotion to white and black is less than seven, Athreos isn't a creature.\n" +
			"Whenever another creature you own dies, return it to your hand unless target opponent pays 3 life.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 ||
		mode.Targets[0].Allow != game.TargetAllowPlayer ||
		!mode.Targets[0].Selection.Exists ||
		mode.Targets[0].Selection.Val.Player != game.PlayerOpponent {
		t.Fatalf("targets = %#v", mode.Targets)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	pay, ok := mode.Sequence[0].Primitive.(game.Pay)
	if !ok ||
		!pay.Payment.Payer.Exists ||
		pay.Payment.Payer.Val != game.TargetPlayerReference(0) ||
		len(pay.Payment.AdditionalCosts) != 1 ||
		pay.Payment.AdditionalCosts[0].Kind != cost.AdditionalPayLife ||
		pay.Payment.AdditionalCosts[0].Amount != 3 ||
		mode.Sequence[0].PublishResult != targetPlayerPaidResultKey {
		t.Fatalf("payment instruction = %#v", mode.Sequence[0])
	}
	move, ok := mode.Sequence[1].Primitive.(game.MoveCard)
	if !ok ||
		move.Card.Kind != game.CardReferenceEvent ||
		move.FromZone != zone.Graveyard ||
		move.Destination != zone.Hand ||
		!mode.Sequence[1].ResultGate.Exists ||
		mode.Sequence[1].ResultGate.Val.Key != targetPlayerPaidResultKey ||
		mode.Sequence[1].ResultGate.Val.Succeeded != game.TriFalse {
		t.Fatalf("return instruction = %#v", mode.Sequence[1])
	}
}
