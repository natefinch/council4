package rules

import (
	"strings"
	"testing"

	cardn "github.com/natefinch/council4/mtg/cards/n"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

type neyithChoiceAgent struct {
	target id.ID
	pay    bool
}

func (*neyithChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *neyithChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		for _, option := range request.Options {
			if len(option.Targets) == 1 && option.Targets[0].PermanentID == a.target {
				return []int{option.Index}
			}
		}
		return []int{0}
	}
	if strings.HasPrefix(request.Prompt, "Pay ") && a.pay {
		return []int{1}
	}
	return []int{0}
}

func putNeyithCombatTrigger(
	t *testing.T,
	engine *Engine,
	g *game.Game,
	agent PlayerAgent,
	log *TurnLog,
) *game.StackObject {
	t.Helper()
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	emitEvent(g, game.Event{
		Kind:       game.EventBeginningOfStep,
		Controller: game.Player1,
		Player:     game.Player1,
		Step:       game.StepBeginningOfCombat,
	})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agentsAll(agent), log) {
		t.Fatal("Neyith beginning-of-combat ability did not trigger")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("Neyith trigger was not put on the stack")
	}
	return obj
}

func TestNeyithDrawTriggerBatchesFightAndBlockedCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cardn.NeyithOfTheDireHunt())
	for range 2 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	}

	first := addCombatPermanent(g, game.Player1, vanillaCreature("First Fighter", 0, 4))
	second := addCombatPermanent(g, game.Player1, vanillaCreature("Second Fighter", 0, 4))
	resolveFightPermanents(g, first, second)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controlled creatures fighting did not trigger Neyith")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("fight trigger count = %d, want one for the simultaneous fight", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size after fight = %d, want 1", got)
	}

	firstBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	secondBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	declare := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: firstBlocker.ObjectID, Blocking: first.ObjectID},
		{Blocker: secondBlocker.ObjectID, Blocking: second.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, declare) {
		t.Fatal("declaring blockers failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controlled creatures becoming blocked did not trigger Neyith")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("blocked trigger count = %d, want one for the declaration batch", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size after blocked batch = %d, want 2", got)
	}
}

func TestNeyithFightTriggerSurvivesLethalFight(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	neyith := addCombatPermanent(g, game.Player1, cardn.NeyithOfTheDireHunt())
	opponent := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	resolveFightPermanents(g, neyith, opponent)
	if changed, _ := engine.checkPermanentStateBasedActions(g, newPassBatchID(g)); !changed {
		t.Fatal("lethal fight did not cause state-based actions")
	}
	if _, ok := permanentByObjectID(g, neyith.ObjectID); ok {
		t.Fatal("Neyith survived lethal fight")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Neyith's captured fight trigger was lost when she died")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one captured fight trigger", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want one card drawn", got)
	}
}

func TestNeyithHybridPaymentUsesLivePowerAndCombatDuration(t *testing.T) {
	for _, hybrid := range []mana.Color{mana.R, mana.G} {
		t.Run(string(hybrid), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, cardn.NeyithOfTheDireHunt())
			target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			g.Players[game.Player1].ManaPool.Add(hybrid, 1)
			agent := &neyithChoiceAgent{target: target.ObjectID, pay: true}
			log := &TurnLog{}

			obj := putNeyithCombatTrigger(t, engine, g, agent, log)
			if len(obj.Targets) != 1 || obj.Targets[0].PermanentID != target.ObjectID {
				t.Fatalf("trigger targets = %+v, want target creature %v", obj.Targets, target.ObjectID)
			}
			target.Counters.Add(counter.PlusOnePlusOne, 1)
			engine.resolveTopOfStackWithChoices(g, agentsAll(agent), log)

			if got := effectivePower(g, target); got != 6 {
				t.Fatalf("effective power = %d, want live power 3 doubled to 6", got)
			}
			target.Counters.Add(counter.PlusOnePlusOne, 1)
			if got := effectivePower(g, target); got != 7 {
				t.Fatalf("effective power after later counter = %d, want fixed +3 doubling modifier", got)
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
				t.Fatalf("mana remaining = %d, want hybrid payment to spend all mana", got)
			}
			if len(g.RuleEffects) != 1 ||
				g.RuleEffects[0].Kind != game.RuleEffectMustBeBlocked ||
				g.RuleEffects[0].AffectedObjectID != target.ObjectID ||
				g.RuleEffects[0].Duration != game.DurationUntilEndOfCombat {
				t.Fatalf("rule effects = %+v, want target must-block this combat", g.RuleEffects)
			}

			g.Turn.Step = game.StepDeclareBlockers
			g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
				Attacker: target.ObjectID,
				Target:   game.AttackTarget{Player: game.Player2},
			}}}
			noBlocks := mustDeclareBlockersPayload(t, action.DeclareBlockers(nil))
			if engine.applyDeclareBlockers(g, game.Player2, noBlocks) {
				t.Fatal("no-block declaration accepted while Neyith's requirement was satisfiable")
			}
			required := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{
				Blocker:  blocker.ObjectID,
				Blocking: target.ObjectID,
			}}))
			if !engine.applyDeclareBlockers(g, game.Player2, required) {
				t.Fatal("required block was rejected")
			}

			expireEndOfCombatRuleEffects(g)
			g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
				Attacker: target.ObjectID,
				Target:   game.AttackTarget{Player: game.Player2},
			}}}
			if !engine.applyDeclareBlockers(g, game.Player2, noBlocks) {
				t.Fatal("must-block requirement survived into an extra combat")
			}
			if got := effectivePower(g, target); got != 7 {
				t.Fatalf("power after combat = %d, want fixed doubling modifier to last until end of turn", got)
			}
			expireCleanupDurations(g)
			if got := effectivePower(g, target); got != 4 {
				t.Fatalf("power after cleanup = %d, want counter-modified power 4", got)
			}
		})
	}
}

func TestNeyithDeclinedOrFailedPaymentHasNoConsequences(t *testing.T) {
	tests := []struct {
		name      string
		pay       bool
		colorMana int
	}{
		{name: "declined", pay: false, colorMana: 1},
		{name: "insufficient mana", pay: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, cardn.NeyithOfTheDireHunt())
			target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			g.Players[game.Player1].ManaPool.Add(mana.G, tc.colorMana)
			beforeMana := g.Players[game.Player1].ManaPool.Total()
			agent := &neyithChoiceAgent{target: target.ObjectID, pay: tc.pay}

			putNeyithCombatTrigger(t, engine, g, agent, &TurnLog{})
			engine.resolveTopOfStackWithChoices(g, agentsAll(agent), &TurnLog{})

			if got := effectivePower(g, target); got != 2 {
				t.Fatalf("effective power = %d, want 2 after unsuccessful payment", got)
			}
			if len(g.RuleEffects) != 0 {
				t.Fatalf("rule effects = %+v, want none after unsuccessful payment", g.RuleEffects)
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != beforeMana {
				t.Fatalf("mana remaining = %d, want unchanged %d", got, beforeMana)
			}
		})
	}
}

func TestNeyithTargetsOnTriggerAndRespectsObjectAndSourceControlChanges(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	neyith := addCombatPermanent(g, game.Player1, cardn.NeyithOfTheDireHunt())
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	noncreature := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	agent := &neyithChoiceAgent{target: target.ObjectID, pay: true}
	log := &TurnLog{}

	obj := putNeyithCombatTrigger(t, engine, g, agent, log)
	if len(obj.Targets) != 1 || obj.Targets[0].PermanentID != target.ObjectID {
		t.Fatalf("stack targets = %+v, want chosen creature", obj.Targets)
	}
	for _, choice := range log.Choices {
		if choice.Request.Kind != game.ChoiceTarget {
			continue
		}
		for _, option := range choice.Request.Options {
			if len(option.Targets) == 1 && option.Targets[0].PermanentID == noncreature.ObjectID {
				t.Fatal("noncreature permanent was offered as a Neyith target")
			}
		}
	}

	neyith.Controller = game.Player2
	target.Controller = game.Player3
	engine.resolveTopOfStackWithChoices(g, agentsAll(agent), log)
	if got := effectivePower(g, target); got != 4 {
		t.Fatalf("power after control changes = %d, want 4", got)
	}
	if len(g.RuleEffects) != 1 || g.RuleEffects[0].AffectedObjectID != target.ObjectID {
		t.Fatalf("rule effects after control changes = %+v, want chosen object", g.RuleEffects)
	}

	expireEndOfCombatRuleEffects(g)
	emitEvent(g, game.Event{
		Kind:       game.EventBeginningOfStep,
		Controller: game.Player1,
		Player:     game.Player1,
		Step:       game.StepBeginningOfCombat,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Neyith controlled by Player2 triggered during Player1's extra combat")
	}
}

func TestNeyithTriggerFizzlesWhenTargetLeavesBeforeResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cardn.NeyithOfTheDireHunt())
	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	agent := &neyithChoiceAgent{target: target.ObjectID, pay: true}

	putNeyithCombatTrigger(t, engine, g, agent, &TurnLog{})
	if !movePermanentToZone(g, target, zone.Graveyard) {
		t.Fatal("moving target to graveyard failed")
	}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agentsAll(agent), log)

	if got := g.Players[game.Player1].ManaPool.Total(); got != 3 {
		t.Fatalf("mana remaining = %d, want no payment after target became illegal", got)
	}
	if len(g.RuleEffects) != 0 || len(g.ContinuousEffects) != 0 {
		t.Fatalf("effects after fizzle: rules=%+v continuous=%+v", g.RuleEffects, g.ContinuousEffects)
	}
	if len(log.Choices) != 0 {
		t.Fatalf("resolution choices = %+v, want no payment prompt after fizzle", log.Choices)
	}
}
