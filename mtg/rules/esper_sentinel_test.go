package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestFirstFilteredSpellCastOrdinalPerOpponentAndTurn(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, esperSentinelPattern(), []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}, nil)

	emitEvent(g, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player2,
		CardTypes:  []types.Card{types.Creature},
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("creature spell triggered first-noncreature ability")
	}
	emitEvent(g, game.Event{
		Kind:       game.EventSpellCopied,
		Controller: game.Player2,
		CardTypes:  []types.Card{types.Instant},
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("spell copy triggered cast ability")
	}
	emitEvent(g, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player2,
		CardTypes:  []types.Card{types.Instant},
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		event := g.Events[len(g.Events)-1]
		pattern := esperSentinelPattern()
		t.Fatalf("first noncreature spell did not trigger: event=%#v ordinal=%d matches=%v events=%#v",
			event, filteredSpellCastOrdinalThisTurn(g, event, &pattern.CardSelection),
			triggerMatchesEvent(g, source, pattern, event), g.Events)
	}
	emitEvent(g, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player2,
		CardTypes:  []types.Card{types.Sorcery},
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("second noncreature spell triggered")
	}
	emitEvent(g, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player3,
		CardTypes:  []types.Card{types.Artifact},
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first noncreature spell by a different opponent did not trigger")
	}

	engine.advanceToNextTurn(g)
	emitEvent(g, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player2,
		CardTypes:  []types.Card{types.Enchantment},
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first noncreature spell did not reset on the next turn")
	}
}

func TestFirstFilteredSpellCastTriggersUseAPNAPOrder(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player1
	addTriggeredPermanent(g, game.Player1, esperSentinelPattern(), nil, nil)
	addTriggeredPermanent(g, game.Player3, esperSentinelPattern(), nil, nil)

	emitEvent(g, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player2,
		CardTypes:  []types.Card{types.Instant},
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("simultaneous triggers were not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2", got)
	}
	top, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack unexpectedly empty")
	}
	if top.Controller != game.Player3 {
		t.Fatalf("top trigger controller = %v, want Player3 after APNAP placement", top.Controller)
	}
}

func TestDynamicSourcePowerResolutionPayment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		plusCounters  int
		minusCounters int
		removeSource  bool
		lands         int
		acceptPayment bool
		wantDraw      bool
		wantTapped    int
		wantPrompt    string
	}{
		{
			name:          "uses current increased power and payer pays",
			plusCounters:  2,
			lands:         3,
			acceptPayment: true,
			wantTapped:    3,
			wantPrompt:    "Pay {3}?",
		},
		{
			name:          "uses LKI after source leaves",
			plusCounters:  1,
			removeSource:  true,
			lands:         2,
			acceptPayment: true,
			wantTapped:    2,
			wantPrompt:    "Pay {2}?",
		},
		{
			name:          "payer declines",
			lands:         1,
			acceptPayment: false,
			wantDraw:      true,
			wantPrompt:    "Pay {1}?",
		},
		{
			name:          "zero power costs zero",
			minusCounters: 1,
			acceptPayment: true,
			wantPrompt:    "Pay {0}?",
		},
		{
			name:          "negative power clamps to zero",
			minusCounters: 2,
			acceptPayment: true,
			wantPrompt:    "Pay {0}?",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			source := addTriggeredPermanent(g, game.Player1, esperSentinelPattern(), esperSentinelInstructions(), nil)

			emitEvent(g, game.Event{
				Kind:       game.EventSpellCast,
				Controller: game.Player2,
				CardTypes:  []types.Card{types.Instant},
			})
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("spell-cast trigger was not put on the stack")
			}
			source.Counters.Add(counter.PlusOnePlusOne, tc.plusCounters)
			source.Counters.Add(counter.MinusOneMinusOne, tc.minusCounters)
			if tc.removeSource {
				movePermanentToZone(g, source, zone.Graveyard)
			}
			lands := make([]*game.Permanent, tc.lands)
			for i := range lands {
				lands[i] = addBasicLandPermanent(g, game.Player2, types.Forest)
			}
			payerChoice := []int{0}
			if tc.acceptPayment {
				payerChoice = []int{1}
			}
			agents := [game.NumPlayers]PlayerAgent{
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
			tapped := 0
			for _, land := range lands {
				if land.Tapped {
					tapped++
				}
			}
			if tapped != tc.wantTapped {
				t.Fatalf("tapped lands = %d, want %d", tapped, tc.wantTapped)
			}
			if len(log.Choices) != 1 || log.Choices[0].Request.Player != game.Player2 ||
				log.Choices[0].Request.Prompt != tc.wantPrompt {
				t.Fatalf("choices = %+v, want payer Player2 prompt %q", log.Choices, tc.wantPrompt)
			}
		})
	}
}

func TestDynamicSourcePowerResolutionPaymentUsesOriginalObjectLKIThroughBlink(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addTriggeredPermanent(g, game.Player1, esperSentinelPattern(), esperSentinelInstructions(), nil)
	source.Counters.Add(counter.PlusOnePlusOne, 2)

	emitEvent(g, game.Event{
		Kind:       game.EventSpellCast,
		Controller: game.Player2,
		CardTypes:  []types.Card{types.Instant},
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("spell-cast trigger was not put on the stack")
	}
	if !movePermanentToZone(g, source, zone.Exile) {
		t.Fatal("failed to exile source")
	}
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	returned, ok := createCardPermanent(g, card, game.Player1, zone.Exile)
	if !ok {
		t.Fatal("failed to return source card to the battlefield")
	}
	if returned.ObjectID == source.ObjectID {
		t.Fatal("returned source reused the original object identity")
	}
	returned.Counters.Add(counter.MinusOneMinusOne, 1)

	lands := make([]*game.Permanent, 3)
	for i := range lands {
		lands[i] = addBasicLandPermanent(g, game.Player2, types.Forest)
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if len(log.Choices) != 1 || log.Choices[0].Request.Prompt != "Pay {3}?" {
		t.Fatalf("choices = %+v, want original object's LKI payment prompt %q", log.Choices, "Pay {3}?")
	}
	for i, land := range lands {
		if !land.Tapped {
			t.Fatalf("land %d was not tapped to pay original object's LKI power", i)
		}
	}
}

func esperSentinelPattern() *game.TriggerPattern {
	return &game.TriggerPattern{
		Event:                      game.EventSpellCast,
		Controller:                 game.TriggerControllerOpponent,
		PlayerEventOrdinalThisTurn: 1,
		CardSelection: game.Selection{
			ExcludedTypes: []types.Card{types.Creature},
		},
	}
}

func esperSentinelInstructions() []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Payer: opt.Val(game.EventPlayerReference()),
				DynamicGenericManaCost: opt.Val(&game.DynamicAmount{
					Kind:       game.DynamicAmountObjectPower,
					Multiplier: 1,
					Object:     game.SourcePermanentReference(),
				}),
			}},
			PublishResult: "unless-paid",
		},
		{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       "unless-paid",
				Succeeded: game.TriFalse,
			}),
		},
	}
}
