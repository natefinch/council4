package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestResolvingSaddledConditionUsesExactSourceObject(t *testing.T) {
	for _, outcome := range []string{"saddled", "unsaddled", "target decoy", "unavailable source", "returned source"} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, vanillaCreature("Source", 2, 3))
			decoy := addCombatPermanent(g, game.Player3, vanillaCreature("Decoy", 4, 5))
			source.Saddled = outcome == "saddled"
			decoy.Saddled = true
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				Targets: []game.Target{game.PermanentTarget(decoy.ObjectID)},
			}
			if outcome == "unavailable source" {
				obj.SourceID = 0
			}
			if outcome == "returned source" {
				if !movePermanentToZone(g, source, zone.Graveyard) {
					t.Fatal("source did not leave")
				}
				card, exists := g.GetCardInstance(source.CardInstanceID)
				if !exists {
					t.Fatal("original source card missing")
				}
				returned, entered := createCardPermanent(g, card, game.Player1, zone.Graveyard)
				if !entered || returned.ObjectID == source.ObjectID {
					t.Fatal("source did not enter as a new object")
				}
				returned.Saddled = true
			}
			got := effectConditionSatisfied(g, obj, opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{SourceSaddled: true}),
			}))
			if got != (outcome == "saddled") {
				t.Fatalf("saddled=%v: condition borrowed a target/card/new incarnation or lost its exact live source", got)
			}
		})
	}
}
