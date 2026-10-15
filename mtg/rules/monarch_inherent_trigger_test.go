package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/zone"
)

type monarchEndPriorityProbe struct {
	seen  bool
	hand  int
	stack []*game.StackObject
}

func (p *monarchEndPriorityProbe) ChooseAction(obs PlayerObservation, _ []action.Action) action.Action {
	if obs.Turn.Step == game.StepEnd && !p.seen {
		p.seen = true
		p.hand = obs.g.Players[game.Player1].Hand.Size()
		p.stack = obs.g.Stack.Objects()
	}
	return action.Pass()
}

func TestMonarchInherentDrawUsesOrdinaryEndStepStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	setMonarch(g, game.Player1)
	fillLibrary(g, game.Player1, 4)
	probe := &monarchEndPriorityProbe{}
	policies := simPassPolicies()
	policies[game.Player1] = probe
	NewEngine(nil).runEndingPhaseWithLog(g, policies, &TurnLog{})
	if !probe.seen || probe.hand != 0 || len(probe.stack) != 1 {
		t.Fatal("first end-step priority did not see an undrawn card and the inherent trigger on stack")
	}
	top := probe.stack[0]
	if top.Kind != game.StackTriggeredAbility || top.SourceID != 0 || top.SourceCardID != 0 ||
		top.SourceTokenDef != nil || top.Controller != game.Player1 {
		t.Fatal("inherent monarch trigger invented a source or lost its trigger-time controller")
	}
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("ordinary all-pass resolution did not draw exactly once")
	}
}

func TestMonarchInherentDrawCanBeCounteredAndKeepsCapturedPlayer(t *testing.T) {
	for _, counterAbility := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player1
		setMonarch(g, game.Player1)
		fillLibrary(g, game.Player1, 3)
		engine := NewEngine(nil)
		emitBeginningOfStepEvent(g, game.StepEnd)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("inherent draw was not put onto the ordinary stack")
		}
		top, _ := g.Stack.Peek()
		setMonarch(g, game.Player2)
		if counterAbility {
			if !counterTargetStackObject(g, &game.StackObject{Targets: []game.Target{game.StackObjectTarget(top.ID)}}, 0, false, game.CounteredSpellGraveyard) {
				t.Fatal("ordinary ability counter operation rejected the source-less trigger")
			}
		} else {
			engine.resolveTopOfStack(g, &TurnLog{})
		}
		want := 1
		if counterAbility {
			want = 0
		}
		if g.Players[game.Player1].Hand.Size() != want || g.Players[game.Player2].Hand.Size() != 0 {
			t.Fatal("countering or changing the monarch retargeted the captured draw")
		}
	}
}

func TestMonarchInherentAndPermanentOwnedEndTriggersBothResolve(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	setMonarch(g, game.Player1)
	fillLibrary(g, game.Player1, 4)
	archivist := monarchEndStepDrawPermanent(g, game.Player3)
	engine := NewEngine(nil)
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStack(g)
	if g.Stack.Size() != 2 {
		t.Fatal("permanent-owned and inherent monarch end-step triggers did not coexist")
	}
	sources := make(map[game.PlayerID]bool)
	for _, object := range g.Stack.Objects() {
		if object.Controller == game.Player1 && object.SourceID == 0 ||
			object.Controller == game.Player3 && object.SourceID == archivist.ObjectID {
			sources[object.Controller] = true
		}
	}
	if len(sources) != 2 {
		t.Fatal("inherent and permanent-owned trigger provenance were conflated")
	}
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if g.Players[game.Player1].Hand.Size() != 2 {
		t.Fatal("ordinary resolution omitted one monarch end-step draw")
	}
}

func TestMonarchInherentOwnCreatureDamageStillCreatesOccurrence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMonarch(g, game.Player2)
	source := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	markPlayerCombatDamage(g, source, game.Player2, 1, &TurnLog{})
	engine := NewEngine(nil)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatal("own-creature combat damage suppressed a real monarch trigger")
	}
	setMonarch(g, game.Player3)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].IsMonarch {
		t.Fatal("the captured own-controller beneficiary was not resolved")
	}
}

func TestMonarchInherentTransferUsesDepartedCreatureLKI(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMonarch(g, game.Player2)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	markPlayerCombatDamage(g, source, game.Player2, 2, &TurnLog{})
	source.Controller = game.Player3
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("could not remove the damage creature")
	}
	engine := NewEngine(nil)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("source-less transfer disappeared with the damage creature")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player3].IsMonarch {
		t.Fatal("transfer did not use the creature's last-known controller")
	}
}

func TestMonarchInherentCombatTransferWaitsForResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMonarch(g, game.Player2)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	markPlayerCombatDamage(g, attacker, game.Player2, 3, &TurnLog{})
	if !g.Players[game.Player2].IsMonarch || g.Players[game.Player1].IsMonarch {
		t.Fatal("combat damage transferred monarchy before its trigger resolved")
	}
	engine := NewEngine(nil)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat damage did not queue the source-less transfer")
	}
	top, _ := g.Stack.Peek()
	if top.Controller != game.Player2 || top.SourceID != 0 {
		t.Fatal("transfer was not controlled by the trigger-time monarch without a source")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("transfer did not resolve through the ordinary stack")
	}
}

func TestMonarchInherentSimultaneousDamageKeepsEveryOccurrence(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMonarch(g, game.Player2)
	var events []damageEvent
	for _, controller := range []game.PlayerID{game.Player1, game.Player3, game.Player1} {
		source := addCombatCreaturePermanentWithPower(g, controller, 1)
		event := combatDamageEvent(g, source, 1)
		event.player = game.Player2
		events = append(events, event)
	}
	resolveDamageBatch(g, events, g.IDGen.Next(), true, &TurnLog{})
	if !g.Players[game.Player2].IsMonarch {
		t.Fatal("the simultaneous event transferred the crown before priority")
	}
	engine := NewEngine(nil)
	engine.putTriggeredAbilitiesOnStack(g)
	if g.Stack.Size() != 3 {
		t.Fatalf("simultaneous monarch occurrences = %d, want one per damage source", g.Stack.Size())
	}
	for _, object := range g.Stack.Objects() {
		if object.Controller != game.Player2 || object.SourceID != 0 || object.SourceCardID != 0 {
			t.Fatal("simultaneous occurrence lost its source-less trigger-time controller")
		}
	}
	for !g.Stack.IsEmpty() {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("all captured beneficiaries were not resolved in ordinary stack order")
	}
}

func TestMonarchInherentSourceLessDrawCopyAndClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	setMonarch(g, game.Player1)
	fillLibrary(g, game.Player1, 4)
	engine := NewEngine(nil)
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStack(g)
	original, _ := g.Stack.Peek()
	addEffectSpellToStack(g, game.Player3, game.CopyStackObject{
		Object: game.TargetStackObjectReference(0), Count: 1,
	}, []game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 2 {
		t.Fatal("ordinary copy operation did not copy the source-less draw")
	}
	branch := engine.Simulator().ResolvePriority(g, simPassPolicies())
	if branch.Players[game.Player1].Hand.Size() != 2 || g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("copied draw lost its captured beneficiary or Simulator independence")
	}
	if branch.Players[game.Player3].Hand.Size() != 0 {
		t.Fatal("copy controller replaced the original draw beneficiary")
	}
}

func TestMonarchInherentTransferReadsCreatureControllerAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMonarch(g, game.Player2)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	markPlayerCombatDamage(g, source, game.Player2, 2, &TurnLog{})
	source.Controller = game.Player3
	engine := NewEngine(nil)
	engine.putTriggeredAbilitiesOnStack(g)
	object, _ := g.Stack.Peek()
	addEffectSpellToStack(g, game.Player4, game.ChangeStackObjectController{
		Object: game.TargetStackObjectReference(0), Controller: game.ControllerReference(),
	}, []game.Target{game.StackObjectTarget(object.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player3].IsMonarch || g.Players[game.Player1].IsMonarch || g.Players[game.Player4].IsMonarch {
		t.Fatal("transfer did not read the creature's controller independently of the ability's controller")
	}
}

func TestMonarchInherentDepartureAndCannotBecome(t *testing.T) {
	for _, controllerLeaves := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		setMonarch(g, game.Player2)
		source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		markPlayerCombatDamage(g, source, game.Player2, 2, &TurnLog{})
		engine := NewEngine(nil)
		if controllerLeaves {
			eliminatePlayer(g, game.Player2)
			engine.putTriggeredAbilitiesOnStack(g)
			if !g.Stack.IsEmpty() {
				t.Fatal("departed monarch controlled a newly placed inherent trigger")
			}
		} else {
			engine.putTriggeredAbilitiesOnStack(g)
			g.Players[game.Player1].CantBecomeMonarchThisTurn = true
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].IsMonarch {
				t.Fatal("transfer bypassed cannot-become at actual resolution")
			}
		}
	}
}
