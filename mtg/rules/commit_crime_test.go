package rules

import (
	"testing"

	cardd "github.com/natefinch/council4/mtg/cards/d"
	cardg "github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestCommitCrimeIncludesSplicedTargetsAndExcludesCopies(t *testing.T) {
	t.Run("actual splice", func(t *testing.T) {
		g, engine := newSpliceGame()
		host := addCardToHand(g, game.Player1, cardd.DesperateRitual())
		addCardToHand(g, game.Player1, cardg.GlacialRay())
		addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
			Event: game.EventCrimeCommitted, Player: game.TriggerPlayerYou,
		}, []game.Instruction{{Primitive: game.GainLife{
			Player: game.ControllerReference(), Amount: game.Fixed(1),
		}}}, nil)
		g.Players[game.Player1].ManaPool.Add(mana.C, 2)
		g.Players[game.Player1].ManaPool.Add(mana.R, 2)
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &spliceScriptAgent{
			splices: []string{"Glacial Ray"}, targets: []game.Target{game.PlayerTarget(game.Player2)},
		}}
		if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(host, nil, 0, nil), agents, &TurnLog{}) {
			t.Fatal("actual host with an opponent-targeted splice did not cast")
		}
		obj, _ := g.Stack.Peek()
		if len(obj.Targets) != 0 || len(obj.SplicedTargets) != 1 || !committedCrime(g, obj) {
			t.Fatal("crime detection did not use the actual spliced announcement")
		}
		if countEventsOfKind(g.Events, game.EventCrimeCommitted) != 1 {
			t.Fatal("one announcement did not commit exactly one crime")
		}
		copy := *obj
		copy.ID, copy.Copy = g.IDGen.Next(), true
		emitTargetEvents(g, &copy)
		if countEventsOfKind(g.Events, game.EventCrimeCommitted) != 1 {
			t.Fatal("a spell copy committed a new crime")
		}
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("spliced qualifying target did not trigger the crime ability")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player1].Life != 41 {
			t.Fatal("crime ability did not resolve once")
		}
	})
}

func TestCommitCrimeCountsCompleteAnnouncementOnlyOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackSpell, Controller: game.Player1,
		Targets: []game.Target{game.PlayerTarget(game.Player3)},
		SplicedTargets: [][]game.Target{
			{game.PlayerTarget(game.Player2)}, {game.PlayerTarget(game.Player3)},
		},
	}
	emitTargetEvents(g, obj)
	if countEventsOfKind(g.Events, game.EventCrimeCommitted) != 1 {
		t.Fatal("host and multiple qualifying splice targets did not constitute one crime")
	}
}

// TestCommitCrimeEmittedForOpponentPlayerTarget covers the core "commit a crime"
// event (CR 700.13): a spell that targets an opponent fires its controller's
// "Whenever you commit a crime" trigger.
func TestCommitCrimeEmittedForOpponentPlayerTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCrimeCommitted,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}
	g.Stack.Push(spell)

	emitTargetEvents(g, spell)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("commit-a-crime trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want commit-a-crime trigger to draw one card", got)
	}
}

// TestCommitCrimeNotEmittedForOwnPlayerTarget confirms a spell that targets only
// its own controller is not a crime, so the trigger does not fire.
func TestCommitCrimeNotEmittedForOwnPlayerTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCrimeCommitted,
		Player: game.TriggerPlayerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player1)},
	}
	g.Stack.Push(spell)

	emitTargetEvents(g, spell)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("commit-a-crime trigger fired for a non-crime self-target")
	}
}

// TestCommitCrimeEmittedForOpponentControlledPermanent confirms that targeting a
// permanent an opponent controls commits a crime, and that an "Whenever an
// opponent commits a crime" trigger fires for the targeted player's controller.
func TestCommitCrimeEmittedForOpponentControlledPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:  game.EventCrimeCommitted,
		Player: game.TriggerPlayerOpponent,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	victim := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Victim"}})
	spell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(victim.ObjectID)},
	}
	g.Stack.Push(spell)

	emitTargetEvents(g, spell)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent commit-a-crime trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want opponent commit-a-crime trigger to draw one card", got)
	}
}
