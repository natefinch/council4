package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// sacrificeSourceUnlessPayContent mirrors the ability content produced by
// lowerSacrificeSourceUnlessPaySpell for "sacrifice this creature unless you pay
// {cost}.": offer the controller a fixed payment, then sacrifice the source if
// the payment was not made.
func sacrificeSourceUnlessPayContent(manaCost cost.Mana) game.AbilityContent {
	const resultKey = game.ResultKey("sacrifice-unless-paid")
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.Pay{Payment: game.ResolutionPayment{
					Prompt:   "Pay " + manaCost.String() + "?",
					ManaCost: opt.Val(manaCost),
				}},
				PublishResult: resultKey,
			},
			{
				Primitive: game.Sacrifice{Object: game.SourcePermanentReference()},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       resultKey,
					Succeeded: game.TriFalse,
				}),
			},
		},
	}.Ability()
}

func pushSacrificeUnlessPayTrigger(g *game.Game, source *game.Permanent, manaCost cost.Mana) {
	trigger := game.TriggeredAbility{Content: sacrificeSourceUnlessPayContent(manaCost)}
	g.Stack.Push(&game.StackObject{
		ID:            g.IDGen.Next(),
		Kind:          game.StackTriggeredAbility,
		SourceID:      source.ObjectID,
		SourceCardID:  source.CardInstanceID,
		Controller:    source.Controller,
		InlineTrigger: &trigger,
	})
}

// TestSacrificeUnlessPaySacrificesSourceWhenControllerCannotPay proves that the
// "sacrifice this creature unless you pay {U}" ability sacrifices its own source
// when the controller has no mana to pay the upkeep cost.
func TestSacrificeUnlessPaySacrificesSourceWhenControllerCannotPay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Phantasmal Forces",
		Types: []types.Card{types.Creature},
	}})
	pushSacrificeUnlessPayTrigger(g, source, cost.Mana{cost.U})

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) {
		t.Fatal("source was not sacrificed when the controller could not pay")
	}
}

// TestSacrificeUnlessPayKeepsSourceWhenControllerPays proves that paying the
// upkeep cost spares the source permanent from sacrifice.
func TestSacrificeUnlessPayKeepsSourceWhenControllerPays(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Phantasmal Forces",
		Types: []types.Card{types.Creature},
	}})
	addBasicLandPermanent(g, game.Player1, types.Island)
	pushSacrificeUnlessPayTrigger(g, source, cost.Mana{cost.U})

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) {
		t.Fatal("source was sacrificed even though the controller paid the cost")
	}
}
