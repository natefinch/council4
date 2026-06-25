package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func spellTargetsControllerCreaturePattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:            game.EventSpellCast,
		Controller:       game.TriggerControllerYou,
		SpellTargetAllow: game.TargetAllowPermanent,
		SpellTargetPattern: opt.Val(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}
}

func TestSpellTargetPatternFiresOnControllerCreatureRelation(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, spellTargetsControllerCreaturePattern(), []game.Instruction{{
		Primitive: game.AddCounter{
			Amount:      game.Fixed(1),
			Object:      game.SourcePermanentReference(),
			CounterKind: counter.PlusOnePlusOne,
		},
	}}, nil)

	ownCreature := addCreaturePermanent(g, game.Player1)
	opponentCreature := addCreaturePermanent(g, game.Player2)

	// A controller spell targeting an opponent's creature must not fire.
	castSpellTargeting(g, game.Player1, game.PermanentTarget(opponentCreature.ObjectID))
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("fired on a spell targeting an opponent's creature")
	}

	// An opponent's spell targeting the controller's creature must not fire ("you cast").
	castSpellTargeting(g, game.Player2, game.PermanentTarget(ownCreature.ObjectID))
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("fired on an opponent's spell")
	}

	// The controller's spell targeting their own creature fires.
	castSpellTargeting(g, game.Player1, game.PermanentTarget(ownCreature.ObjectID))
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("did not fire on a controller spell targeting their own creature")
	}
}
