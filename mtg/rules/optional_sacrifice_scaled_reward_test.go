package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestOptionalSacrificeScaledRewardReadsLastKnownPower proves Disciple of
// Freyalise's reward: after the controller sacrifices another creature, the
// "gain X life and draw X cards, where X is that creature's power" rewards read
// the sacrificed creature's last-known power through the linked object the
// sacrifice published, even though that creature has left the battlefield.
func TestOptionalSacrificeScaledRewardReadsLastKnownPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sacrificed := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	for range 6 {
		addLibraryCard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Library Card",
			Types: []types.Card{types.Land},
		}})
	}
	startLife := g.Players[game.Player1].Life
	startHand := g.Players[game.Player1].Hand.Size()

	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		Controller:   game.Player1,
		SourceID:     g.IDGen.Next(),
		SourceCardID: g.IDGen.Next(),
	}
	log := &TurnLog{}

	resolveInstruction(engine, g, obj, game.SacrificePermanents{
		Player:        game.ControllerReference(),
		Amount:        game.Fixed(1),
		Selection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
		PublishLinked: sacrificedCreatureLinkKey,
	}, log)

	if _, ok := permanentByObjectID(g, sacrificed.ObjectID); ok {
		t.Fatal("creature was not sacrificed")
	}

	rewardAmount := game.Dynamic(game.DynamicAmount{
		Kind:       game.DynamicAmountObjectPower,
		Multiplier: 1,
		Object:     game.LinkedObjectReference(string(sacrificedCreatureLinkKey)),
	})
	resolveInstruction(engine, g, obj, game.GainLife{
		Player: game.ControllerReference(),
		Amount: rewardAmount,
	}, log)
	resolveInstruction(engine, g, obj, game.Draw{
		Player: game.ControllerReference(),
		Amount: rewardAmount,
	}, log)

	if got := g.Players[game.Player1].Life - startLife; got != 4 {
		t.Fatalf("life gained = %d, want 4 (sacrificed creature's power)", got)
	}
	if got := g.Players[game.Player1].Hand.Size() - startHand; got != 4 {
		t.Fatalf("cards drawn = %d, want 4 (sacrificed creature's power)", got)
	}
}

// sacrificedCreatureLinkKey mirrors the cardgen lowerer's linked-object key so
// this runtime test exercises the exact key the generated card publishes under.
const sacrificedCreatureLinkKey = game.LinkedKey("sacrificed-creature")
