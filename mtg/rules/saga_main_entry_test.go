package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type sagaEntryProbe struct {
	saga      id.ID
	drawLore  []int
	mainLore  []int
	mainStack []int
}

func (p *sagaEntryProbe) ChooseAction(obs PlayerObservation, _ []action.Action) action.Action {
	if permanent, present := permanentByObjectID(obs.g, p.saga); present {
		lore := permanent.Counters.Get(counter.Lore)
		if obs.Turn.Step == game.StepDraw {
			p.drawLore = append(p.drawLore, lore)
		}
		if obs.Turn.Phase == game.PhasePrecombatMain && obs.Turn.Step == game.StepNone {
			p.mainLore = append(p.mainLore, lore)
			p.mainStack = append(p.mainStack, obs.g.Stack.Size())
		}
	}
	return action.Pass()
}

func TestSagaMainEntryLoreTimingWithAndWithoutDraw(t *testing.T) {
	for _, skipDraw := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player1
		saga := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{sagaDrawChapter(2), sagaDrawChapter(3)}))
		saga.Counters.Add(counter.Lore, 1)
		stockLibrary(g, game.Player1, 6)
		if skipDraw {
			g.SkippedSteps[game.Player1] = map[game.Step]int{game.StepDraw: 1}
		}
		probe := &sagaEntryProbe{saga: saga.ObjectID}
		policies := simPassPolicies()
		policies[game.Player1] = probe
		engine := NewEngine(nil)
		engine.runBeginningPhase(g, policies, &TurnLog{})
		for _, lore := range probe.drawLore {
			if lore != 1 {
				t.Fatal("Saga advanced inside the draw-step priority window")
			}
		}
		engine.runMainPhase(g, policies, game.PhasePrecombatMain, &TurnLog{})
		if len(probe.mainLore) == 0 || probe.mainLore[0] != 2 || probe.mainStack[0] == 0 {
			t.Fatalf("precombat main did not advance lore/queue its chapter (skipDraw=%v): %v/%v",
				skipDraw, probe.mainLore, probe.mainStack)
		}
	}
}

func TestSagaMainEntryAdditionalMainDoesNotAdvanceAgain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	saga := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{sagaDrawChapter(5)}))
	saga.Counters.Add(counter.Lore, 1)
	engine := NewEngine(nil)
	engine.runMainPhase(g, simPassPolicies(), game.PhasePrecombatMain, &TurnLog{})
	engine.runMainPhase(g, simPassPolicies(), game.PhasePrecombatMain, &TurnLog{})
	if saga.Counters.Get(counter.Lore) != 2 || g.Turn.Phase != game.PhasePostcombatMain {
		t.Fatal("additional main phases were classified from enum names instead of turn history")
	}
}

func TestSagaMainEntryFiltersEffectiveChaptersAndUsesCounterReplacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	var sagas []*game.Permanent
	for range 4 {
		permanent := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{sagaDrawChapter(5)}))
		permanent.Counters.Add(counter.Lore, 1)
		sagas = append(sagas, permanent)
	}
	sagas[1].PhasedOut = true
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{Layer: game.LayerAbility, AffectedObjectID: sagas[2].ObjectID, RemoveAllAbilities: true},
		game.ContinuousEffect{Layer: game.LayerType, AffectedObjectID: sagas[3].ObjectID, SetTypes: []types.Card{types.Land}},
	)
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventCountersAdded, MatchCounterKind: true,
		CounterKindFilter: counter.Lore, CounterMultiplier: 2,
	})
	NewEngine(nil).runMainPhase(g, simPassPolicies(), game.PhasePrecombatMain, &TurnLog{})
	for index, saga := range sagas {
		want := 1
		if index == 0 {
			want = 3
		}
		if got := saga.Counters.Get(counter.Lore); got != want {
			t.Fatalf("Saga %d lore = %d, want %d", index, got, want)
		}
	}
}

type sagaDrawControlPolicy struct {
	sagaEntryProbe

	changed bool
}

func (p *sagaDrawControlPolicy) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Step == game.StepDraw && !p.changed {
		for _, candidate := range legal {
			if activation, ok := candidate.ActivateAbilityPayload(); ok && activation.SourceID == p.saga {
				p.changed = true
				return candidate
			}
		}
	}
	return p.sagaEntryProbe.ChooseAction(obs, legal)
}

func TestSagaMainEntryUsesControllerAfterDrawPriority(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	def := sagaCard([]game.ChapterAbility{sagaDrawChapter(2), sagaDrawChapter(3)})
	def.ActivatedAbilities = []game.ActivatedAbility{{Content: game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyContinuous{
			Object: opt.Val(game.SourcePermanentReference()),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerControl, NewController: opt.Val(game.Player2),
			}},
		},
	}}}.Ability()}}
	saga := addCombatPermanent(g, game.Player1, def)
	saga.Counters.Add(counter.Lore, 1)
	stockLibrary(g, game.Player1, 6)
	probe := &sagaDrawControlPolicy{sagaEntryProbe: sagaEntryProbe{saga: saga.ObjectID}}
	policies := simPassPolicies()
	policies[game.Player1] = probe
	engine := NewEngine(nil)
	engine.runBeginningPhase(g, policies, &TurnLog{})
	engine.runMainPhase(g, policies, game.PhasePrecombatMain, &TurnLog{})
	if !probe.changed || effectiveController(g, saga) != game.Player2 || saga.Counters.Get(counter.Lore) != 1 {
		t.Fatal("precombat-main lore ignored a real control change resolved during draw priority")
	}
}

func TestSagaMainEntrySimulatorDoesNotRepeatAlreadyEnteredMain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	saga := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{sagaDrawChapter(5)}))
	saga.Counters.Add(counter.Lore, 1)
	engine := NewEngine(nil)
	engine.runMainPhase(g, simPassPolicies(), game.PhasePrecombatMain, &TurnLog{})
	g.Turn.ExtraPhases = []game.Phase{game.PhasePrecombatMain, game.PhaseBeginning}
	stockLibrary(g, game.Player1, 6)
	forwarded := engine.Simulator().PlayForward(g, simPassPolicies(), 0)
	continued, present := permanentByObjectID(forwarded, saga.ObjectID)
	if !present || continued.Counters.Get(counter.Lore) != 2 || saga.Counters.Get(counter.Lore) != 2 {
		t.Fatal("Simulator continuation or extra beginning/main phases repeated Saga advancement")
	}
	if g.Turn.MainPhasesThisTurn != 1 {
		t.Fatal("Simulator changed the original phase-entry history")
	}
}

func TestSagaMainEntryInsertedMainPreservesChronologicalTiming(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	engine.runMainPhaseWithContinuation(g, simPassPolicies(), game.PhasePostcombatMain, game.PhasePrecombatMain, &TurnLog{})
	if g.Turn.Phase != game.PhasePrecombatMain {
		t.Fatal("the first inserted main was not the actual precombat main")
	}
	engine.runMainPhase(g, simPassPolicies(), game.PhasePrecombatMain, &TurnLog{})
	if g.Turn.Phase != game.PhasePostcombatMain || g.Turn.MainPhaseContinuation.Val != game.PhaseCombat {
		t.Fatal("regular main classification lost its upcoming regular combat")
	}
	if !canCastAtCurrentTiming(g, game.Player1, beforeCombatDamageInstant()) ||
		!turnBeforeAttackersDeclared(g.Turn) {
		t.Fatal("postcombat classification closed timing windows before the actual upcoming combat")
	}
}

func TestSagaMainEntryInsertedMainSimulatorKeepsRegularSchedule(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	engine.runMainPhaseWithContinuation(g, simPassPolicies(), game.PhasePostcombatMain, game.PhasePrecombatMain, &TurnLog{})
	forwarded := engine.Simulator().PlayForward(g, simPassPolicies(), 0)
	postcombat := 0
	for _, event := range forwarded.Events {
		if event.Kind == game.EventBeginningOfStep && event.Step == game.StepPostcombatMain {
			postcombat++
		}
	}
	if postcombat != 2 {
		t.Fatalf("inserted-main continuation entered %d later mains, want regular first and second main slots", postcombat)
	}
}

type sagaMainOrderProbe struct {
	sagaEntryProbe

	orderedTogether bool
	topController   game.PlayerID
}

func (p *sagaMainOrderProbe) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Prompt == "Order triggered abilities." && len(request.Options) == 2 {
		p.orderedTogether = true
		return []int{request.Options[1].Index, request.Options[0].Index}
	}
	return nil
}

func (p *sagaMainOrderProbe) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if len(p.mainLore) == 0 && obs.Turn.Phase == game.PhasePrecombatMain && obs.Turn.Step == game.StepNone {
		if top, present := obs.g.Stack.Peek(); present {
			p.topController = top.Controller
		}
	}
	return p.sagaEntryProbe.ChooseAction(obs, legal)
}

func TestSagaMainEntryChapterAndBeginningTriggersShareAPNAPPlacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	saga := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{sagaDrawChapter(2)}))
	saga.Counters.Add(counter.Lore, 1)
	stockLibrary(g, game.Player1, 4)
	for _, player := range []game.PlayerID{game.Player1, game.Player2} {
		addTriggeredPermanent(g, player, &game.TriggerPattern{
			Event: game.EventBeginningOfStep, Step: game.StepPrecombatMain,
		}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	}
	probe := &sagaMainOrderProbe{sagaEntryProbe: sagaEntryProbe{saga: saga.ObjectID}}
	policies := simPassPolicies()
	policies[game.Player1] = probe
	NewEngine(nil).runMainPhase(g, policies, game.PhasePrecombatMain, &TurnLog{})
	if !probe.orderedTogether || len(probe.mainStack) == 0 || probe.mainStack[0] != 3 ||
		probe.topController != game.Player2 {
		t.Fatal("main and final chapter triggers were not ordered together using APNAP")
	}
	if _, present := permanentByObjectID(g, saga.ObjectID); present {
		t.Fatal("final-chapter sacrifice did not wait for its queued chapter to finish")
	}
}
