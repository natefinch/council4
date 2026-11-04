package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// A spell that resolves is reported as resolved whatever replaces its final
// move to the graveyard (#3612).
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     608.2n ("As the final part of an instant or sorcery spell's
//     resolution, the spell is put into its owner's graveyard."), 702.88a
//     (rebound: "instead of putting it into your graveyard as it resolves,
//     exile it") and 608.2b (a spell whose targets are all illegal doesn't
//     resolve and is put into its owner's graveyard).
//   - Trumpeting Herd, oracle_id 69501650-ed48-4ebf-9287-b0338c5bc5d5 (print
//     b0f3b68e-f616-4687-bc2d-075165162cd1): "Create a 3/3 green Elephant
//     creature token. Rebound"; Scryfall ruling 2019-06-14: a countered
//     rebound spell doesn't resolve and goes to its owner's graveyard.
//   - Green Sun's Zenith, oracle_id 0d96b60b-a060-48ee-bb83-93f1c4a10669
//     (print 70291c7b-a86f-4466-8502-c28765a89b2a): "Search your library for a
//     green creature card with mana value X or less, put it onto the
//     battlefield, then shuffle. Shuffle Green Sun's Zenith into its owner's
//     library."
//
// Rebound, a shuffle into the library or a return to hand changes only where
// the resolved spell goes; the spell still resolved, so exactly one
// spell-resolved event names its card, stack object and controller. A
// countered or fizzled spell reports none.

func spellResolvedEventsFor(g *game.Game, cardID id.ID) []game.Event {
	var events []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventSpellResolved && event.CardID == cardID {
			events = append(events, event)
		}
	}
	return events
}

func requireOneSpellResolvedEvent(t *testing.T, g *game.Game, cardID, stackObjectID id.ID, controller game.PlayerID) {
	t.Helper()
	events := spellResolvedEventsFor(g, cardID)
	if len(events) != 1 {
		t.Fatalf("spell-resolved events for card %v = %+v, want exactly one", cardID, events)
	}
	event := events[0]
	if event.SourceID != cardID || event.StackObjectID != stackObjectID || event.Controller != controller {
		t.Fatalf("spell-resolved event = %+v, want card %v, stack object %v, controller %v",
			event, cardID, stackObjectID, controller)
	}
}

func TestReboundSpellReportsResolution(t *testing.T) {
	t.Parallel()
	registry, herd, _ := reboundReplacementRegistry(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	cardID, obj := castTrumpetingHerd(t, g, engine, herd)
	stackObjectID := obj.ID
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "rebound exile" {
		t.Fatalf("resolve logs = %+v, want one rebound exile", log.Resolves)
	}
	if controlledElephants(g, game.Player1) != 1 || !reboundTriggerPending(g, cardID) {
		t.Fatal("Trumpeting Herd did not resolve and rebound")
	}
	requireOneSpellResolvedEvent(t, g, cardID, stackObjectID, game.Player1)
	exiled := eventIndex(g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == cardID && event.FromZone == zone.Stack && event.ToZone == zone.Exile
	})
	resolved := eventIndex(g.Events, game.EventSpellResolved, func(event game.Event) bool {
		return event.CardID == cardID
	})
	if exiled < 0 || resolved < exiled {
		t.Fatalf("rebound exile at %d, resolution at %d; want resolution reported after the move", exiled, resolved)
	}

	// The rebound recast is a new spell that resolves into the graveyard.
	assertReboundUpkeep(t, g, engine, cardID, true)
	recast, _ := g.Stack.Peek()
	recastID := recast.ID
	engine.resolveTopOfStack(g, &TurnLog{})
	events := spellResolvedEventsFor(g, cardID)
	if len(events) != 2 || events[1].StackObjectID != recastID || recastID == stackObjectID ||
		!g.Players[game.Player1].Graveyard.Contains(cardID) || controlledElephants(g, game.Player1) != 2 {
		t.Fatalf("spell-resolved events = %+v, want the recast %v reported once after the original", events, recastID)
	}
}

func TestReboundSpellResolutionIndependentOfReplacementChoice(t *testing.T) {
	t.Parallel()
	for _, chooseRebound := range []bool{true, false} {
		name := "Dauthi first"
		if chooseRebound {
			name = "rebound first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			registry, herd, dauthi := reboundReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, registry.Programs())
			resolveDauthiVoidwalker(t, g, engine, game.Player2, dauthi)
			cardID, obj := castTrumpetingHerd(t, g, engine, herd)
			stackObjectID := obj.ID
			var requests []game.ChoiceRequest
			engine.resolveTopOfStackWithChoices(g, reboundOrderAgents(t, game.Player1, chooseRebound, &requests), &TurnLog{})
			requireReboundTwoWayChoice(t, requests, game.Player1, g.ReplacementEffects[0].Description)
			if reboundTriggerPending(g, cardID) != chooseRebound ||
				(g.ExileCounterCount(cardID, counter.Void) == 1) == chooseRebound {
				t.Fatal("replacement choice was not applied")
			}
			requireOneSpellResolvedEvent(t, g, cardID, stackObjectID, game.Player1)
		})
	}
}

func TestReboundBuybackChoiceReportsResolutionOnce(t *testing.T) {
	t.Parallel()
	for _, chooseRebound := range []bool{true, false} {
		name := "buyback first"
		if chooseRebound {
			name = "rebound first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, reboundBuybackFixture())
			obj := castReboundBuybackFixture(t, g, engine, cardID)
			stackObjectID := obj.ID
			var requests []game.ChoiceRequest
			engine.resolveTopOfStackWithChoices(g, reboundOrderAgents(t, game.Player1, chooseRebound, &requests), &TurnLog{})
			requireReboundTwoWayChoice(t, requests, game.Player1, "Buyback")
			if g.Players[game.Player1].Exile.Contains(cardID) != chooseRebound {
				t.Fatal("replacement choice was not applied")
			}
			requireOneSpellResolvedEvent(t, g, cardID, stackObjectID, game.Player1)
		})
	}
}

func TestReboundUnresolvedSpellReportsNoResolution(t *testing.T) {
	t.Parallel()
	t.Run("countered", func(t *testing.T) {
		t.Parallel()
		registry, herd, _ := reboundReplacementRegistry(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngineWithPrograms(nil, registry.Programs())
		cardID, obj := castTrumpetingHerd(t, g, engine, herd)
		counterSpellID := addEffectSpellToStack(g, game.Player2, game.CounterObject{Object: game.TargetStackObjectReference(0)},
			[]game.Target{game.StackObjectTarget(obj.ID)})
		engine.resolveTopOfStack(g, &TurnLog{})
		if len(spellResolvedEventsFor(g, counterSpellID)) != 1 {
			t.Fatal("counterspell's resolution was not reported")
		}
		if events := spellResolvedEventsFor(g, cardID); len(events) != 0 ||
			!g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("countered Trumpeting Herd reported %+v or left the graveyard", events)
		}
	})
	t.Run("fizzled", func(t *testing.T) {
		t.Parallel()
		registry, _, _ := reboundReplacementRegistry(t)
		staggershock := registry.Lookup("Staggershock")
		if staggershock == nil || !staggershock.HasKeyword(game.Rebound) {
			t.Fatal("registered Staggershock with rebound missing")
		}
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngineWithPrograms(nil, registry.Programs())
		creature := addCreaturePermanent(g, game.Player3)
		cardID := addCardToHand(g, game.Player1, staggershock)
		for range 3 {
			addBasicLandPermanent(g, game.Player1, types.Mountain)
		}
		giveMainPhasePriority(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.CastSpell(cardID,
			[]game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
			t.Fatal("Player1 could not cast Staggershock")
		}
		if !movePermanentToZone(g, creature, zone.Exile) {
			t.Fatal("could not remove Staggershock's target")
		}
		log := TurnLog{}
		engine.resolveTopOfStack(g, &log)
		if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
			t.Fatalf("resolve logs = %+v, want countered by rules", log.Resolves)
		}
		if events := spellResolvedEventsFor(g, cardID); len(events) != 0 || reboundTriggerPending(g, cardID) {
			t.Fatalf("fizzled Staggershock reported %+v or rebounded", events)
		}
	})
}

func TestReboundCopyAndCloneReportResolutionSeparately(t *testing.T) {
	t.Parallel()
	registry, herd, _ := reboundReplacementRegistry(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	cardID, obj := castTrumpetingHerd(t, g, engine, herd)
	originalID := obj.ID
	copied := copyStackObject(engine, g, obj)
	if copied == nil || !copied.Copy {
		t.Fatalf("top of stack = %+v, want a copy of Trumpeting Herd", copied)
	}
	copyID := copied.ID
	engine.resolveTopOfStack(g, &TurnLog{})
	requireOneSpellResolvedEvent(t, g, cardID, copyID, game.Player1)
	replay := g.Clone()
	engine.resolveTopOfStack(g, &TurnLog{})
	events := spellResolvedEventsFor(g, cardID)
	if len(events) != 2 || events[1].StackObjectID != originalID || !reboundTriggerPending(g, cardID) {
		t.Fatalf("spell-resolved events = %+v, want the copy then the rebounding original %v", events, originalID)
	}
	if len(spellResolvedEventsFor(replay, cardID)) != 1 {
		t.Fatal("original's resolution leaked into the clone")
	}
	engine.resolveTopOfStack(replay, &TurnLog{})
	if replayed := spellResolvedEventsFor(replay, cardID); len(replayed) != 2 ||
		replayed[1].StackObjectID != originalID || !reboundTriggerPending(replay, cardID) {
		t.Fatalf("clone's spell-resolved events = %+v, want its own rebounding resolution", replayed)
	}
}

func TestShuffledSpellReportsResolution(t *testing.T) {
	t.Parallel()
	zenith := cards.NewDefaultRegistry().Lookup("Green Sun's Zenith")
	if zenith == nil {
		t.Fatal("registered Green Sun's Zenith missing")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addImplementationSpellToStack(g, game.Player1, zenith, nil)
	obj, _ := g.Stack.Peek()
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "library" ||
		!g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatalf("resolve logs = %+v, want Green Sun's Zenith shuffled into its library", log.Resolves)
	}
	requireOneSpellResolvedEvent(t, g, cardID, obj.ID, game.Player1)
}

func TestSpellReturnedToHandReportsResolution(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.MoveResolvingSpell{Destination: zone.Hand}},
	}, nil)
	obj, _ := g.Stack.Peek()
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "hand" ||
		!g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatalf("resolve logs = %+v, want the spell returned to its owner's hand", log.Resolves)
	}
	requireOneSpellResolvedEvent(t, g, cardID, obj.ID, game.Player1)
}
