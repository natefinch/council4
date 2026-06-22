package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func heroicPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:              game.EventSpellCast,
		Controller:         game.TriggerControllerYou,
		SpellTargetsSource: true,
	}
}

func castSpellTargeting(g *game.Game, controller game.PlayerID, targets ...game.Target) game.Event {
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: controller,
		Targets:    targets,
	}
	g.Stack.Push(obj)
	event := game.Event{
		Kind:          game.EventSpellCast,
		Controller:    controller,
		StackObjectID: obj.ID,
	}
	emitEvent(g, event)
	return event
}

func TestHeroicTriggerFiresOnlyWhenControllerSpellTargetsSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, heroicPattern(), []game.Instruction{{
		Primitive: game.AddCounter{
			Amount:      game.Fixed(1),
			Object:      game.SourcePermanentReference(),
			CounterKind: counter.PlusOnePlusOne,
		},
	}}, nil)

	// A spell the controller casts that does not target the source must not fire.
	other := addCreaturePermanent(g, game.Player1)
	castSpellTargeting(g, game.Player1, game.PermanentTarget(other.ObjectID))
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("heroic fired on a controller spell that did not target the source")
	}

	// An opponent's spell that targets the source must not fire (heroic is "you cast").
	castSpellTargeting(g, game.Player2, game.PermanentTarget(source.ObjectID))
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("heroic fired on an opponent's spell targeting the source")
	}

	// The controller's spell that targets the source fires.
	if !castSpellTargetingFires(g, engine, source) {
		t.Fatal("heroic did not fire on a controller spell targeting the source")
	}
}

func castSpellTargetingFires(g *game.Game, engine *Engine, source *game.Permanent) bool {
	castSpellTargeting(g, game.Player1, game.PermanentTarget(source.ObjectID))
	return engine.putTriggeredAbilitiesOnStack(g)
}
