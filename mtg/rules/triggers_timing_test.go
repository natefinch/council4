package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestBeginningOfMainPhaseModalTriggerLocksDistinctModesInPrintedOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventBeginningOfStep,
		Controller: game.TriggerControllerYou,
		Step:       game.StepPrecombatMain,
	}, nil, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card instance not found")
	}
	treasure := testTreasureToken()
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	card.Def.TriggeredAbilities[0].Content = game.AbilityContent{
		MinModes: 1,
		MaxModes: 3,
		Modes: []game.Mode{
			{Text: "Sell Contraband", Sequence: []game.Instruction{
				{Primitive: game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(treasure)}},
				{Primitive: game.LoseLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			}},
			{Text: "Buy Information", Sequence: []game.Instruction{
				{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
				{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.ControllerReference()}},
			}},
			{Text: "Hire a Mercenary", Sequence: []game.Instruction{{Primitive: game.LoseLife{Amount: game.Fixed(3), Player: game.ControllerReference()}}}},
		},
	}
	agent := &choiceOnlyAgent{choices: [][]int{{1, 0}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	log := &TurnLog{}

	g.AppendEvent(game.Event{Kind: game.EventBeginningOfStep, Step: game.StepPrecombatMain, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("modal trigger was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !slices.Equal(obj.ChosenModes, []int{0, 1}) {
		t.Fatalf("chosen modes = %v, want selected modes locked in printed order [0 1]", obj.ChosenModes)
	}
	if len(log.Choices) != 1 {
		t.Fatalf("choice log = %#v, want one modal choice", log.Choices)
	}
	request := log.Choices[0].Request
	if request.Kind != game.ChoiceModal || request.Player != game.Player1 ||
		request.MinChoices != 1 || request.MaxChoices != 3 ||
		len(request.Options) != 3 ||
		request.Options[0].Label != "Sell Contraband" ||
		request.Options[1].Label != "Buy Information" ||
		request.Options[2].Label != "Hire a Mercenary" {
		t.Fatalf("modal request = %#v, want printed labels and one-to-three range", request)
	}

	startingLife := g.Players[game.Player1].Life
	engine.resolveTopOfStack(g, log)
	if got := g.Players[game.Player1].Life; got != startingLife-3 {
		t.Fatalf("controller life = %d, want %d after selected modes resolve", got, startingLife-3)
	}
	if got := countTokenDef(g, treasure); got != 1 {
		t.Fatalf("treasure tokens = %d, want one", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("controller hand size = %d, want one drawn card", got)
	}
	var losses []int
	for _, event := range g.Events {
		if event.Kind == game.EventLifeLost && event.Player == game.Player1 {
			losses = append(losses, event.Amount)
		}
	}
	if !slices.Equal(losses, []int{1, 2}) {
		t.Fatalf("life-loss events = %v, want printed mode order [1 2]", losses)
	}
}

func TestTriggeredModalChoiceRejectsDuplicateModes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event: game.EventBeginningOfStep,
		Step:  game.StepPrecombatMain,
	}, nil, nil)
	card, ok := g.GetCardInstance(permanent.CardInstanceID)
	if !ok {
		t.Fatal("trigger source card instance not found")
	}
	card.Def.TriggeredAbilities[0].Content = game.AbilityContent{
		MinModes: 1,
		MaxModes: 2,
		Modes: []game.Mode{
			{Text: "First"},
			{Text: "Second"},
		},
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1, 1}}},
	}
	log := &TurnLog{}

	g.AppendEvent(game.Event{Kind: game.EventBeginningOfStep, Step: game.StepPrecombatMain, Player: game.Player1})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("modal trigger was not put on stack")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !slices.Equal(obj.ChosenModes, []int{0}) {
		t.Fatalf("chosen modes = %v, want distinct fallback [0]", obj.ChosenModes)
	}
	if len(log.Choices) != 1 || !log.Choices[0].UsedFallback {
		t.Fatalf("choice log = %#v, want duplicate selection rejected with fallback", log.Choices)
	}
}

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

func TestPhasingStateTriggerSourcePreservesLatch(t *testing.T) {
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
	source.PhasedOut = true
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("phased-out state trigger source triggered")
	}
	source.PhasedOut = false
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("state trigger re-fired after phasing while original remained on stack")
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
		Selection:  opt.Val(game.Selection{Controller: game.ControllerOpponent}),
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
