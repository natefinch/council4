package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func optionalModalTriggeredCreature() *game.CardDef {
	def := triggeredCreature(&game.TriggerPattern{
		Event:  game.EventAttackerDeclared,
		Source: game.TriggerSourceSelf,
	}, nil, nil)
	def.TriggeredAbilities[0].Content = game.AbilityContent{
		MinModes: 0,
		MaxModes: 1,
		Modes: []game.Mode{
			{
				Text:    "Target player gains 3 life.",
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
				Sequence: []game.Instruction{{Primitive: game.GainLife{
					Amount: game.Fixed(3),
					Player: game.TargetPlayerReference(0),
				}}},
			},
			{
				Text: "You gain 7 life.",
				Sequence: []game.Instruction{{Primitive: game.GainLife{
					Amount: game.Fixed(7),
					Player: game.ControllerReference(),
				}}},
			},
		},
	}
	return def
}

func TestTriggeredModesEmptySelectionDeclinesPlacement(t *testing.T) {
	for _, tc := range []struct {
		name         string
		selection    []int
		singleMode   bool
		noAgent      bool
		wantFallback bool
	}{
		{name: "legal nil selection"},
		{name: "legal empty selection", selection: []int{}},
		{name: "single optional mode", singleMode: true},
		{name: "invalid answer falls back to empty", selection: []int{99}, wantFallback: true},
		{name: "missing agent falls back to empty", noAgent: true, wantFallback: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := optionalModalTriggeredCreature()
			if tc.singleMode {
				def.TriggeredAbilities[0].Content.Modes = def.TriggeredAbilities[0].Content.Modes[:1]
			}
			source := addCombatPermanent(g, game.Player2, def)
			g.Events = append(g.Events, game.Event{
				Kind:        game.EventAttackerDeclared,
				Controller:  game.Player2,
				PermanentID: source.ObjectID,
			})
			var agents [game.NumPlayers]PlayerAgent
			if !tc.noAgent {
				agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{tc.selection}}
			}
			log := &TurnLog{}

			if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
				t.Error("legal zero-mode selection reported successful trigger placement")
			}
			if obj, ok := g.Stack.Peek(); ok {
				t.Errorf("legal zero-mode selection queued stack object %v with modes %v", obj.ID, obj.ChosenModes)
			}
			if len(log.Triggers) != 0 {
				t.Errorf("declined trigger logged %d placements, want none", len(log.Triggers))
			}
			if len(log.Choices) != 1 {
				t.Fatalf("choice requests = %d, want only the mode choice and no target choice", len(log.Choices))
			}
			choice := log.Choices[0]
			if choice.Request.Kind != game.ChoiceModal || choice.Request.Player != game.Player2 ||
				choice.Request.MinChoices != 0 || choice.Request.MaxChoices != 1 ||
				len(choice.Selected) != 0 || choice.UsedFallback != tc.wantFallback {
				t.Fatalf("mode choice = %+v, want empty choice with fallback=%v", choice, tc.wantFallback)
			}
		})
	}
}

func TestTriggeredModesPlacementAndResolution(t *testing.T) {
	for _, tc := range []struct {
		name           string
		choices        [][]int
		requiredMode   bool
		optionalTarget bool
		nonmodal       bool
		wantModes      []int
		wantTargets    []game.Target
		wantCounts     []int
		wantLifePlayer game.PlayerID
		wantLifeGain   int
		wantChoices    int
		wantFallback   bool
	}{
		{
			name: "selected targeted mode", choices: [][]int{{0}, {2}},
			wantModes: []int{0}, wantTargets: []game.Target{game.PlayerTarget(game.Player3)}, wantCounts: []int{1},
			wantLifePlayer: game.Player3, wantLifeGain: 3, wantChoices: 2,
		},
		{
			name: "selected untargeted mode", choices: [][]int{{1}},
			wantModes: []int{1}, wantLifePlayer: game.Player2, wantLifeGain: 7, wantChoices: 1,
		},
		{
			name: "selected mode with zero optional targets", choices: [][]int{{0}, {4}},
			optionalTarget: true, wantModes: []int{0}, wantCounts: []int{0}, wantChoices: 2,
		},
		{
			name: "invalid empty required mode uses fallback", choices: [][]int{nil, {2}},
			requiredMode: true, wantModes: []int{0}, wantTargets: []game.Target{game.PlayerTarget(game.Player3)}, wantCounts: []int{1},
			wantLifePlayer: game.Player3, wantLifeGain: 3, wantChoices: 2, wantFallback: true,
		},
		{
			name: "nonmodal trigger does not choose modes", choices: [][]int{{2}},
			nonmodal: true, wantTargets: []game.Target{game.PlayerTarget(game.Player3)}, wantCounts: []int{1},
			wantLifePlayer: game.Player3, wantLifeGain: 3, wantChoices: 1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := optionalModalTriggeredCreature()
			content := &def.TriggeredAbilities[0].Content
			if tc.requiredMode {
				content.MinModes = 1
			}
			if tc.optionalTarget {
				content.Modes[0].Targets[0].MinTargets = 0
			}
			if tc.nonmodal {
				*content = content.Modes[0].Ability()
			}
			source := addCombatPermanent(g, game.Player2, def)
			g.Events = append(g.Events, game.Event{
				Kind: game.EventAttackerDeclared, Controller: game.Player2, PermanentID: source.ObjectID,
			})
			agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: tc.choices}}
			log := &TurnLog{}

			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
				t.Fatal("trigger was not placed")
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.Kind != game.StackTriggeredAbility || obj.Controller != game.Player2 ||
				obj.SourceID != source.ObjectID || obj.SourceCardID != source.CardInstanceID || obj.AbilityIndex != 0 {
				t.Fatalf("stack object = %+v, want triggered ability with original source/controller identity", obj)
			}
			if !slices.Equal(obj.ChosenModes, tc.wantModes) || !slices.Equal(obj.Targets, tc.wantTargets) {
				t.Fatalf("modes/targets = %v/%v, want %v/%v", obj.ChosenModes, obj.Targets, tc.wantModes, tc.wantTargets)
			}
			if !slices.Equal(obj.TargetCounts, tc.wantCounts) {
				t.Fatalf("target counts = %v, want %v", obj.TargetCounts, tc.wantCounts)
			}
			if len(log.Triggers) != 1 || log.Triggers[0].StackObjectID != obj.ID {
				t.Fatalf("trigger log = %+v, want one placement of stack object %v", log.Triggers, obj.ID)
			}
			if len(log.Choices) != tc.wantChoices {
				t.Fatalf("choice requests = %d, want %d", len(log.Choices), tc.wantChoices)
			}
			for i, choice := range log.Choices {
				wantKind := game.ChoiceTarget
				if !tc.nonmodal && i == 0 {
					wantKind = game.ChoiceModal
				}
				if choice.Request.Kind != wantKind || choice.Request.Player != game.Player2 ||
					choice.UsedFallback != (tc.wantFallback && i == 0) {
					t.Fatalf("choice %d = %+v, want controller choice of kind %v", i, choice, wantKind)
				}
			}

			engine.resolveTopOfStack(g, log)
			if _, ok := g.Stack.Peek(); ok {
				t.Fatal("resolved trigger remained on stack")
			}
			for playerID, player := range g.Players {
				wantLife := 40
				if game.PlayerID(playerID) == tc.wantLifePlayer {
					wantLife += tc.wantLifeGain
				}
				if player.Life != wantLife {
					t.Fatalf("player %d life = %d, want %d", playerID, player.Life, wantLife)
				}
			}
		})
	}
}

func TestTriggeredModesEmptySelectionPreservesUniqueModes(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := optionalModalTriggeredCreature()
	def.TriggeredAbilities[0].Content.ModesUniquePerTurn = true
	source := addCombatPermanent(g, game.Player2, def)
	use := game.TriggeredAbilityUse{SourceID: source.ObjectID, AbilityIndex: 0}
	agent := &choiceOnlyAgent{choices: [][]int{nil, {0}, {2}, nil, {1}, nil}}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}

	for i, tc := range []struct {
		wantPlaced  bool
		wantHistory uint64
		wantOptions []int
	}{
		{wantOptions: []int{0, 1}},
		{wantPlaced: true, wantHistory: 0b01, wantOptions: []int{0, 1}},
		{wantHistory: 0b01, wantOptions: []int{1}},
		{wantPlaced: true, wantHistory: 0b11, wantOptions: []int{1}},
		{wantHistory: 0b11},
	} {
		g.Events = append(g.Events, game.Event{
			Kind: game.EventAttackerDeclared, Controller: game.Player2, PermanentID: source.ObjectID,
		})
		log := &TurnLog{}
		if placed := engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log); placed != tc.wantPlaced {
			t.Fatalf("trigger %d placement = %v, want %v", i, placed, tc.wantPlaced)
		}
		if history, exists := g.ChosenModesThisTurn[use]; history != tc.wantHistory || exists != (tc.wantHistory != 0) {
			t.Fatalf("trigger %d mode history = %b/%v, want %b/%v", i, history, exists, tc.wantHistory, tc.wantHistory != 0)
		}
		if len(log.Choices) == 0 || log.Choices[0].Request.Kind != game.ChoiceModal || log.Choices[0].UsedFallback {
			t.Fatalf("trigger %d choices = %+v, want valid mode choice without fallback", i, log.Choices)
		}
		var options []int
		for _, option := range log.Choices[0].Request.Options {
			options = append(options, option.Index)
		}
		if !slices.Equal(options, tc.wantOptions) {
			t.Fatalf("trigger %d mode options = %v, want %v", i, options, tc.wantOptions)
		}
		if tc.wantPlaced {
			engine.resolveTopOfStack(g, log)
		} else if len(log.Choices) != 1 || len(log.Triggers) != 0 {
			t.Fatalf("declined trigger %d solicited targets or logged placement: %+v", i, log)
		}
		if _, ok := g.Stack.Peek(); ok {
			t.Fatalf("trigger %d left an unexpected stack object", i)
		}
	}
	if g.Players[game.Player3].Life != 43 || g.Players[game.Player2].Life != 47 {
		t.Fatalf("life totals = %d/%d, want each selected unique mode to resolve once",
			g.Players[game.Player3].Life, g.Players[game.Player2].Life)
	}
}

func TestTriggeredModesEmptySelectionReleasesStateTriggerLatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := sacrificeWhenControlNoIsland(g, game.Player2)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("state trigger source card not found")
	}
	card.Def.TriggeredAbilities[0].Content = optionalModalTriggeredCreature().TriggeredAbilities[0].Content
	agent := &choiceOnlyAgent{choices: [][]int{nil, {1}, {1}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	log := &TurnLog{}

	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("declined state trigger was placed on stack")
	}
	if _, ok := g.Stack.Peek(); ok || len(log.Choices) != 1 || log.Choices[0].UsedFallback || len(log.Triggers) != 0 {
		t.Fatalf("declined state trigger left a stack object or unexpected choices/placements: %+v", log)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("declining a mode kept the state trigger latched")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != source.ObjectID || obj.SourceCardID != source.CardInstanceID ||
		obj.Controller != game.Player2 || !slices.Equal(obj.ChosenModes, []int{1}) {
		t.Fatalf("state trigger stack object = %+v, want selected mode with original identity", obj)
	}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("state trigger fired again while its selected mode was still on stack")
	}
	if len(log.Choices) != 2 || len(log.Triggers) != 1 {
		t.Fatalf("latched state trigger made extra choices/placements: %+v", log)
	}
	engine.resolveTopOfStack(g, log)
	if g.Players[game.Player2].Life != 47 {
		t.Fatalf("controller life = %d, want state trigger to resolve selected mode", g.Players[game.Player2].Life)
	}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("resolved state trigger remained latched")
	}
	engine.resolveTopOfStack(g, log)
	if g.Players[game.Player2].Life != 54 || len(log.Choices) != 3 || len(log.Triggers) != 2 {
		t.Fatalf("state trigger did not fire and resolve again: life=%d, choices=%d, placements=%d",
			g.Players[game.Player2].Life, len(log.Choices), len(log.Triggers))
	}
}
