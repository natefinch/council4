package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestEventPlayerTaxedOptionalDrawUsesActualSpellCaster(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerOpponent,
	}, eventPlayerTaxedOptionalDrawInstructions(cost.Mana{cost.O(1)}), nil)
	spellID := addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, types.Forest)
	addBasicLandPermanent(g, game.Player2, types.Forest)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player2, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("opponent spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent spell-cast trigger was not put on the stack")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want no draw after caster paid", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Player != game.Player2 {
		t.Fatalf("choices = %+v, want payment choice for actual spell caster", log.Choices)
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want cast spell still waiting below trigger", got)
	}
}

func TestEventPlayerTaxedOptionalDrawChoicesAndPayment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		payerCanPay       bool
		payerAccepts      bool
		controllerAccepts bool
		wantDraw          bool
		wantChoicePlayers []game.PlayerID
		wantLandTapped    bool
	}{
		{
			name:              "pays",
			payerCanPay:       true,
			payerAccepts:      true,
			controllerAccepts: true,
			wantChoicePlayers: []game.PlayerID{game.Player2},
			wantLandTapped:    true,
		},
		{
			name:              "declines and controller draws",
			payerCanPay:       true,
			controllerAccepts: true,
			wantDraw:          true,
			wantChoicePlayers: []game.PlayerID{game.Player2, game.Player1},
		},
		{
			name:              "declines and controller declines",
			payerCanPay:       true,
			wantChoicePlayers: []game.PlayerID{game.Player2, game.Player1},
		},
		{
			name:              "cannot pay and controller draws",
			controllerAccepts: true,
			wantDraw:          true,
			wantChoicePlayers: []game.PlayerID{game.Player1},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
				Event:      game.EventSpellCast,
				Controller: game.TriggerControllerOpponent,
			}, eventPlayerTaxedOptionalDrawInstructions(cost.Mana{cost.O(1)}), nil)

			var payerLand *game.Permanent
			if tc.payerCanPay {
				payerLand = addBasicLandPermanent(g, game.Player2, types.Forest)
			}
			emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player2})
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("opponent spell-cast trigger was not put on the stack")
			}

			payerChoice := []int{0}
			if tc.payerAccepts {
				payerChoice = []int{1}
			}
			controllerChoice := []int{0}
			if tc.controllerAccepts {
				controllerChoice = []int{1}
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{controllerChoice}},
				game.Player2: &choiceOnlyAgent{choices: [][]int{payerChoice}},
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, &log)

			wantHand := 0
			if tc.wantDraw {
				wantHand = 1
			}
			if got := g.Players[game.Player1].Hand.Size(); got != wantHand {
				t.Fatalf("hand size = %d, want %d", got, wantHand)
			}
			if len(log.Choices) != len(tc.wantChoicePlayers) {
				t.Fatalf("choices = %+v, want players %v", log.Choices, tc.wantChoicePlayers)
			}
			for i, player := range tc.wantChoicePlayers {
				if log.Choices[i].Request.Kind != game.ChoiceMay ||
					log.Choices[i].Request.Player != player {
					t.Fatalf("choice %d = %+v, want may choice for player %v", i, log.Choices[i], player)
				}
			}
			if payerLand != nil && payerLand.Tapped != tc.wantLandTapped {
				t.Fatalf("payer land tapped = %v, want %v", payerLand.Tapped, tc.wantLandTapped)
			}
		})
	}
}

func TestEventPlayerTaxedOptionalDrawOpponentFilterAndMultipleEvents(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Draw"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Draw"}})
	addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerOpponent,
	}, eventPlayerTaxedOptionalDrawInstructions(cost.Mana{cost.O(1)}), nil)

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player1})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controller's spell incorrectly triggered opponent-cast ability")
	}
	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player2})
	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player3})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent spell-cast triggers were not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want two independent triggers", got)
	}

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want one draw per opponent event", got)
	}
}

func TestEventPlayerTaxedOptionalDrawPersistsAfterSourceLeaves(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	source := addTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerOpponent,
	}, eventPlayerTaxedOptionalDrawInstructions(cost.Mana{cost.O(1)}), nil)

	emitEvent(g, game.Event{Kind: game.EventSpellCast, Controller: game.Player2})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent spell-cast trigger was not put on the stack")
	}
	movePermanentToZone(g, source, zone.Graveyard)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want trigger to resolve from source LKI", got)
	}
}

func eventPlayerTaxedOptionalDrawInstructions(manaCost cost.Mana) []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt:   "Pay " + manaCost.String() + "?",
				Payer:    opt.Val(game.EventPlayerReference()),
				ManaCost: opt.Val(manaCost),
			}},
			PublishResult: "unless-paid",
		},
		{
			Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			Optional:  true,
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       "unless-paid",
				Succeeded: game.TriFalse,
			}),
		},
	}
}
