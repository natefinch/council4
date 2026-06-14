package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCastTriggerGoesOnStackAboveCastSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	spellID := addCardToHand(g, game.Player1, greenInstant())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("cast instant failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("cast trigger was not put on stack")
	}

	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility {
		t.Fatalf("top of stack = %+v, want cast trigger above cast spell", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want cast trigger to draw one card", got)
	}
}

func TestBeginningOfUpkeepTriggerResolvesBeforeDrawStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Upkeep Draw"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepUpkeep,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want upkeep trigger plus draw step draw", got)
	}
}

func TestBeginningOfEndStepTriggerResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "End Step Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepEnd,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want end-step trigger draw", got)
	}
}

func TestBeginningOfDrawStepTriggerResolvesAfterTurnDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Trigger Draw"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Turn Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepDraw,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want turn draw plus draw-step trigger", got)
	}
}

func TestBeginningOfCombatTriggerResolves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Combat Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepBeginningOfCombat,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runCombatPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want beginning-of-combat trigger draw", got)
	}
}

func TestBeginningOfMainPhaseTriggerResolvesAtCorrectBoundary(t *testing.T) {
	tests := []struct {
		name       string
		phase      game.Phase
		wrongPhase game.Phase
		step       game.Step
	}{
		{"precombat", game.PhasePrecombatMain, game.PhasePostcombatMain, game.StepPrecombatMain},
		{"postcombat", game.PhasePostcombatMain, game.PhasePrecombatMain, game.StepPostcombatMain},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Main Phase Draw"}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Should Stay in Library"}})
			addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
				Event:      game.EventBeginningOfStep,
				Controller: game.TriggerControllerYou,
				Step:       test.step,
			}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

			recorder := &mainPhaseStepRecorder{}
			agents := [game.NumPlayers]PlayerAgent{game.Player1: recorder}
			engine.runMainPhase(g, agents, test.wrongPhase, &TurnLog{})
			if got := g.Players[game.Player1].Hand.Size(); got != 0 {
				t.Fatalf("hand size at wrong boundary = %d, want 0", got)
			}
			engine.runMainPhase(g, agents, test.phase, &TurnLog{})

			if got := g.Players[game.Player1].Hand.Size(); got != 1 {
				t.Fatalf("hand size = %d, want one main-phase trigger draw", got)
			}
			if recorder.sawNonNoneStep || g.Turn.Step != game.StepNone {
				t.Fatalf("main-phase priority observed step %v, want StepNone", g.Turn.Step)
			}
		})
	}
}

func TestBeginningOfMainPhaseTriggerDoesNotFireOnOpponentTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player2
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Should Not Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Controller: game.TriggerControllerYou,
		Step:       game.StepPrecombatMain,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runMainPhase(g, [game.NumPlayers]PlayerAgent{}, game.PhasePrecombatMain, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want no draw on opponent's main phase", got)
	}
}

func TestBeginningOfStepTriggerRequiresExplicitStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Should Not Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want only turn draw without broad step trigger", got)
	}
}

type mainPhaseStepRecorder struct {
	sawNonNoneStep bool
}

func (r *mainPhaseStepRecorder) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if (obs.Turn.Phase == game.PhasePrecombatMain || obs.Turn.Phase == game.PhasePostcombatMain) &&
		obs.Turn.Step != game.StepNone {
		r.sawNonNoneStep = true
	}
	return action.Pass()
}

func TestResolvedStateTriggerReleasesLatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.State = opt.Val(game.StateTriggerCondition{MatchControllerLifeLessOrEqual: true, ControllerLifeLessOrEqual: 10})
	g.Players[game.Player1].Life = 10

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger was not put on stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("resolved state trigger did not re-fire while condition remained true")
	}
}

func TestStateTriggerDoesNotRetriggerBeforeLeavingStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{}, nil, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.State = opt.Val(game.StateTriggerCondition{
		MatchControllerLifeLessOrEqual: true,
		ControllerLifeLessOrEqual:      10,
	})
	g.Players[game.Player1].Life = 10

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger was not put on stack")
	}
	g.Players[game.Player1].Life = 11
	engine.putTriggeredAbilitiesOnStack(g)
	g.Players[game.Player1].Life = 10
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger re-fired before the original left the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want original state trigger only", got)
	}
}

func TestCounteredStateTriggerReleasesLatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{}, []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}, nil)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.State = opt.Val(game.StateTriggerCondition{
		MatchControllerLifeLessOrEqual: true,
		ControllerLifeLessOrEqual:      10,
	})
	g.Players[game.Player1].Life = 10

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger was not put on stack")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Kind != game.StackTriggeredAbility {
		t.Fatal("state trigger stack object missing")
	}
	if !counterStackObject(g, trigger.ID) {
		t.Fatal("state trigger was not countered")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("countered state trigger did not re-fire while condition remained true")
	}
}

func TestStateTriggerWithoutLegalTargetsReleasesLatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{}, nil, []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Predicate:  game.TargetPredicate{Controller: game.ControllerOpponent},
	}})
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.TriggeredAbilities[0].Trigger.State = opt.Val(game.StateTriggerCondition{
		MatchControllerLifeLessOrEqual: true,
		ControllerLifeLessOrEqual:      10,
	})
	g.Players[game.Player1].Life = 10

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger with no legal targets was put on stack")
	}
	addBasicLandPermanent(g, game.Player2, types.Forest)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger did not re-fire after legal target became available")
	}
}
