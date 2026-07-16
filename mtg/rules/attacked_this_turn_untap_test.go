package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func attackedThisTurnUntapGroup() game.GroupReference {
	return game.AttackedThisTurnGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})
}

func eachCombatAttackerUntapDef() *game.DelayedTriggerDef {
	return &game.DelayedTriggerDef{
		EventPattern: opt.Val(game.TriggerPattern{
			Event: game.EventBeginningOfStep,
			Step:  game.StepBeginningOfCombat,
		}),
		Window: game.DelayedWindowThisTurn,
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Untap{Group: attackedThisTurnUntapGroup()},
		}}}.Ability(),
	}
}

func scheduleEachCombatAttackerUntap(t *testing.T, g *game.Game, controller game.PlayerID) {
	t.Helper()
	if !scheduleDelayedTrigger(g, &game.StackObject{Controller: controller}, eachCombatAttackerUntapDef()) {
		t.Fatal("scheduleDelayedTrigger failed")
	}
}

func declareAttackedHistory(g *game.Game, permanents ...*game.Permanent) {
	for _, permanent := range permanents {
		emitEvent(g, game.Event{
			Kind:           game.EventAttackerDeclared,
			SourceObjectID: permanent.ObjectID,
			PermanentID:    permanent.ObjectID,
			Controller:     permanent.Controller,
			Player:         game.Player2,
		})
	}
}

func fireBeginningOfCombatAndResolve(t *testing.T, g *game.Game, engine *Engine) {
	t.Helper()
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("beginning of combat did not fire delayed trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
}

func TestEachCombatThisTurnUntapsAllPriorAttackersAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	first := addCombatCreaturePermanent(g, game.Player1)
	second := addCombatCreaturePermanent(g, game.Player1)
	nonattacker := addCombatCreaturePermanent(g, game.Player1)
	first.Tapped, second.Tapped, nonattacker.Tapped = true, true, true
	declareAttackedHistory(g, first, second, first)
	scheduleEachCombatAttackerUntap(t, g, game.Player1)

	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("beginning of combat did not fire delayed trigger")
	}
	if !first.Tapped || !second.Tapped {
		t.Fatal("attackers untapped before the delayed trigger resolved")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if first.Tapped || second.Tapped {
		t.Fatalf("prior attackers remained tapped: first=%v second=%v", first.Tapped, second.Tapped)
	}
	if !nonattacker.Tapped {
		t.Fatal("nonattacker was untapped")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("repeating delayed trigger count = %d, want 1", len(g.DelayedTriggers))
	}
}

func TestEachCombatThisTurnUsesCumulativeObjectIdentityHistory(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	survivor := addCombatCreaturePermanent(g, game.Player1)
	departed := addCombatCreaturePermanent(g, game.Player1)
	declareAttackedHistory(g, survivor, departed)
	scheduleEachCombatAttackerUntap(t, g, game.Player1)

	g.Battlefield = []*game.Permanent{survivor}
	returned := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: departed.CardInstanceID,
		Owner:          departed.Owner,
		Controller:     departed.Controller,
		Tapped:         true,
	}
	g.Battlefield = append(g.Battlefield, returned)
	survivor.Tapped = true
	fireBeginningOfCombatAndResolve(t, g, engine)

	if survivor.Tapped {
		t.Fatal("surviving prior attacker remained tapped")
	}
	if !returned.Tapped {
		t.Fatal("returned new object inherited the old object's attack history")
	}

	later := addCombatCreaturePermanent(g, game.Player1)
	declareAttackedHistory(g, later)
	survivor.Tapped, later.Tapped = true, true
	fireBeginningOfCombatAndResolve(t, g, engine)
	if survivor.Tapped || later.Tapped {
		t.Fatalf("cumulative attackers remained tapped: survivor=%v later=%v", survivor.Tapped, later.Tapped)
	}
}

func TestEachCombatThisTurnMultipleResolutionsStackAndExpire(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Tapped = true
	declareAttackedHistory(g, attacker)
	scheduleEachCombatAttackerUntap(t, g, game.Player1)
	scheduleEachCombatAttackerUntap(t, g, game.Player1)

	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 2 {
		t.Fatalf("stack size = %d, want two delayed triggers", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if attacker.Tapped {
		t.Fatal("attacker remained tapped")
	}
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("delayed triggers = %d, want two repeating triggers", len(g.DelayedTriggers))
	}

	expireEventDelayedTriggers(g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers survived turn cleanup: %d", len(g.DelayedTriggers))
	}
	g.Turn.ExtraPhases = []game.Phase{game.PhaseCombat}
	engine.advanceToNextTurn(g)
	if len(g.Turn.ExtraPhases) != 0 {
		t.Fatalf("extra phase schedule survived turn transition: %#v", g.Turn.ExtraPhases)
	}
	attacker.Tapped = true
	addEffectSpellToStack(g, game.Player1, game.Untap{Group: attackedThisTurnUntapGroup()}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !attacker.Tapped {
		t.Fatal("previous turn's attack history survived the turn transition")
	}
}

func TestEachCombatThisTurnSurvivesSourceLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	source := addCombatCreaturePermanent(g, game.Player1)
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Tapped = true
	declareAttackedHistory(g, attacker)

	if !scheduleDelayedTrigger(g, &game.StackObject{
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
	}, eachCombatAttackerUntapDef()) {
		t.Fatal("scheduleDelayedTrigger failed")
	}
	g.Battlefield = []*game.Permanent{attacker}

	fireBeginningOfCombatAndResolve(t, g, engine)
	if attacker.Tapped {
		t.Fatal("delayed trigger stopped working after its source left")
	}
}
