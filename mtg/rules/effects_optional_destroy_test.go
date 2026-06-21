package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addToughnessCreaturePermanent adds a creature permanent with the given
// toughness so an "equal to its toughness" follow-up has a value to read.
func addToughnessCreaturePermanent(g *game.Game, controller game.PlayerID, toughness int) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:      "Toughness Creature",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: toughness}),
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// optionalDestroyGainLifeInstructions builds Noxious Gearhulk's resolving flow:
// "you may destroy another target creature. If a creature is destroyed this way,
// you gain life equal to its toughness." The destroy is Optional and publishes
// its result; the gain-life is gated on that result having succeeded and reads
// the destroyed creature's toughness.
func optionalDestroyGainLifeInstructions() []game.Instruction {
	return []game.Instruction{
		{
			Optional:      true,
			Primitive:     game.Destroy{Object: game.TargetPermanentReference(0)},
			PublishResult: game.ResultKey("if-you-do"),
		},
		{
			Primitive: game.GainLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:       game.DynamicAmountObjectToughness,
					Multiplier: 1,
					Object:     game.TargetPermanentReference(0),
				}),
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       game.ResultKey("if-you-do"),
				Succeeded: game.TriTrue,
			}),
		},
	}
}

// TestOptionalDestroyThisWayAcceptDestroysAndGainsToughness verifies that
// accepting the optional destroy moves the target creature to the graveyard and
// gains life equal to that creature's toughness.
func TestOptionalDestroyThisWayAcceptDestroysAndGainsToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	victim := addToughnessCreaturePermanent(g, game.Player2, 4)
	addInstructionSpellToStackForController(
		g, game.Player1,
		optionalDestroyGainLifeInstructions(),
		[]game.Target{game.PermanentTarget(victim.ObjectID)},
	)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if _, ok := g.PermanentByID(victim.ObjectID); ok {
		t.Fatal("accepting must destroy the targeted creature")
	}
	if !g.Players[game.Player2].Graveyard.Contains(victim.CardInstanceID) {
		t.Fatal("destroyed creature must be in its owner's graveyard")
	}
	if got := g.Players[game.Player1].Life; got != before+4 {
		t.Fatalf("life = %d, want %d (gain equal to destroyed toughness)", got, before+4)
	}
}

// TestOptionalDestroyThisWayDeclineSkipsDestroyAndGain verifies that declining
// the optional destroy leaves the target creature on the battlefield and skips
// the gated life gain.
func TestOptionalDestroyThisWayDeclineSkipsDestroyAndGain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	before := g.Players[game.Player1].Life
	victim := addToughnessCreaturePermanent(g, game.Player2, 4)
	addInstructionSpellToStackForController(
		g, game.Player1,
		optionalDestroyGainLifeInstructions(),
		[]game.Target{game.PermanentTarget(victim.ObjectID)},
	)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if _, ok := g.PermanentByID(victim.ObjectID); !ok {
		t.Fatal("declining must leave the targeted creature on the battlefield")
	}
	if got := g.Players[game.Player1].Life; got != before {
		t.Fatalf("life = %d, want %d (declining must skip the gated gain)", got, before)
	}
}
