package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestResultObjectCardIdentityFreezesReachedZoneVersion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
	obj := &game.StackObject{Controller: game.Player1}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)}, PublishResult: "discard",
	}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	card, _ := g.GetCardInstance(cardID)
	reachedVersion := card.ZoneVersion
	snapshots := obj.ResolutionResultObjects["discard"]
	if len(snapshots) != 1 || snapshots[0].ObjectID != 0 || snapshots[0].CardID != cardID ||
		len(snapshots[0].ZoneCards) != 1 || snapshots[0].ZoneCards[0].ZoneVersion != reachedVersion {
		t.Fatalf("post-move card identity = %#v", snapshots)
	}
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Hand) {
		t.Fatal("subsequent move failed")
	}
	if card.ZoneVersion == reachedVersion || snapshots[0].ZoneCards[0].ZoneVersion != reachedVersion {
		t.Fatal("result identity followed the card into its later zone incarnation")
	}
	if resultObjectsMatchFilter(g, obj, []game.ObjectSnapshot{{ObjectID: 99, CardID: cardID, Types: []types.Card{types.Land}}},
		game.Selection{}, true) {
		t.Fatal("card-only filter accepted a pre-departure permanent identity")
	}
}
