package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func ravenousCreatureDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Test Ravenous",
		ManaCost:  opt.Val(cost.Mana{cost.X, cost.G}),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.RavenousEntersWithCountersReplacement(),
		},
		TriggeredAbilities: []game.TriggeredAbility{
			game.RavenousDrawTriggeredAbility(),
		},
	}}
}

func TestRavenousCastXThresholdAndTriggerTiming(t *testing.T) {
	for _, tc := range []struct {
		name     string
		x        int
		wantDraw bool
	}{
		{name: "zero", x: 0},
		{name: "below five", x: 4},
		{name: "at five", x: 5, wantDraw: true},
		{name: "above five", x: 6, wantDraw: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
			spellID := addCardToHand(g, game.Player1, ravenousCreatureDef())
			for range tc.x + 1 {
				addBasicLandPermanent(g, game.Player1, types.Forest)
			}
			g.Turn.Phase = game.PhasePrecombatMain
			g.Turn.Step = game.StepNone

			if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, tc.x, nil)) {
				t.Fatalf("cast Ravenous with X=%d failed", tc.x)
			}
			engine.resolveTopOfStack(g, &TurnLog{})

			permanent := permanentForCard(g, spellID)
			if permanent == nil {
				t.Fatal("Ravenous spell did not become a permanent")
			}
			if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != tc.x {
				t.Fatalf("+1/+1 counters = %d, want %d", got, tc.x)
			}
			if !permanent.EnteredFromCast || permanent.CastXValue != tc.x {
				t.Fatalf("permanent cast state = cast:%v X:%d, want cast with X=%d", permanent.EnteredFromCast, permanent.CastXValue, tc.x)
			}
			if got := g.Players[game.Player1].Hand.Size(); got != 0 {
				t.Fatalf("hand size before Ravenous trigger resolves = %d, want 0", got)
			}
			triggered := engine.putTriggeredAbilitiesOnStack(g)
			if triggered != tc.wantDraw {
				t.Fatalf("putTriggeredAbilitiesOnStack = %v, want %v", triggered, tc.wantDraw)
			}
			if !tc.wantDraw {
				return
			}
			top, _ := g.Stack.Peek()
			if top.XValue != tc.x {
				t.Fatalf("Ravenous trigger X = %d, want preserved cast X=%d", top.XValue, tc.x)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if got := g.Players[game.Player1].Hand.Size(); got != 1 {
				t.Fatalf("hand size after Ravenous trigger = %d, want 1", got)
			}
		})
	}
}

func TestRavenousNonCastEntryTreatsXAsZero(t *testing.T) {
	for _, tc := range []struct {
		name    string
		options permanentCreationOptions
	}{
		{
			name:    "copied spell state ignores copied X",
			options: permanentCreationOptions{XValue: 9},
		},
		{name: "put directly onto battlefield"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, ravenousCreatureDef())
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				t.Fatal("Ravenous card missing")
			}
			g.Players[game.Player1].Hand.Remove(cardID)

			permanent, ok := createCardPermanentFaceWithOptions(
				engine,
				g,
				card,
				game.Player1,
				zone.Hand,
				game.FaceFront,
				nil,
				tc.options,
				[game.NumPlayers]PlayerAgent{},
				nil,
			)
			if !ok {
				t.Fatal("non-cast Ravenous entry failed")
			}
			if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
				t.Fatalf("+1/+1 counters = %d, want 0 for non-cast entry", got)
			}
			if permanent.EnteredFromCast || permanent.CastXValue != 0 {
				t.Fatalf("permanent cast state = cast:%v X:%d, want non-cast X=0", permanent.EnteredFromCast, permanent.CastXValue)
			}
			event := g.Events[len(g.Events)-1]
			if event.Kind != game.EventPermanentEnteredBattlefield || event.EnterWasCast || event.EnterXValue != 0 {
				t.Fatalf("entry event = %#v, want non-cast X=0", event)
			}
			if engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("non-cast Ravenous entry incorrectly created a draw trigger")
			}
		})
	}
}

func TestRavenousCounterReplacementDoesNotChangeDrawThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, counterDoublingReplacementCardDef())
	cardID := addCardToHand(g, game.Player1, ravenousCreatureDef())
	card, _ := g.GetCardInstance(cardID)

	permanent, ok := createCardPermanentFaceWithOptions(
		engine,
		g,
		card,
		game.Player1,
		zone.Stack,
		game.FaceFront,
		nil,
		permanentCreationOptions{WasCast: true, XValue: 4},
		[game.NumPlayers]PlayerAgent{},
		nil,
	)
	if !ok {
		t.Fatal("Ravenous entry failed")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 8 {
		t.Fatalf("replacement-adjusted +1/+1 counters = %d, want 8", got)
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Ravenous draw trigger used replaced counter count instead of chosen X=4")
	}
}

func TestRavenousDrawUsesTriggerControllerAfterControlChange(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	cardID := addCardToHand(g, game.Player1, ravenousCreatureDef())
	card, _ := g.GetCardInstance(cardID)
	g.Players[game.Player1].Hand.Remove(cardID)
	permanent, ok := createCardPermanentFaceWithOptions(
		engine,
		g,
		card,
		game.Player1,
		zone.Stack,
		game.FaceFront,
		nil,
		permanentCreationOptions{
			WasCast:           true,
			CastController:    game.Player1,
			HasCastController: true,
			XValue:            5,
		},
		[game.NumPlayers]PlayerAgent{},
		nil,
	)
	if !ok || !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Ravenous X=5 trigger was not created")
	}

	permanent.Controller = game.Player2
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("original trigger controller hand = %d, want 1", got)
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 0 {
		t.Fatalf("new permanent controller hand = %d, want 0", got)
	}
}

func TestEnterTriggerXIsCapturedOnlyForEnteringSource(t *testing.T) {
	event := game.Event{
		Kind:         game.EventPermanentEnteredBattlefield,
		PermanentID:  41,
		EnterWasCast: true,
		EnterXValue:  6,
	}
	if got := triggeredAbilityXValue(&pendingTriggeredAbility{
		sourceID: 41,
		event:    event,
		hasEvent: true,
	}); got != 6 {
		t.Fatalf("self enter trigger X = %d, want 6", got)
	}
	if got := triggeredAbilityXValue(&pendingTriggeredAbility{
		sourceID: 42,
		event:    event,
		hasEvent: true,
	}); got != 0 {
		t.Fatalf("watching permanent trigger X = %d, want 0", got)
	}
	event.EnterWasCast = false
	if got := triggeredAbilityXValue(&pendingTriggeredAbility{
		sourceID: 41,
		event:    event,
		hasEvent: true,
	}); got != 0 {
		t.Fatalf("non-cast self enter trigger X = %d, want 0", got)
	}
}
