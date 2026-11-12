package rules

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLegalActionsIncludeSuspendFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, suspendSorcery(3, cost.Mana{cost.G}))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)

	if !actionsContain(legal, action.SuspendCard(cardID)) {
		t.Fatalf("legal actions = %+v, want suspend action", legal)
	}
}

func TestSuspendActionPaysCostAndExilesWithTimeCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, suspendSorcery(3, cost.Mana{cost.G}))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.SuspendCard(cardID)) {
		t.Fatal("suspend action failed")
	}

	if !forest.Tapped {
		t.Fatal("suspend cost did not tap mana source")
	}
	if g.Players[game.Player1].Hand.Contains(cardID) || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("suspended card did not move from hand to exile")
	}
	if got := g.ExileCounterCount(cardID, counter.Time); got != 3 || !cardIsSuspended(g, cardID) {
		t.Fatalf("time counters = %d suspended = %t, want suspended with 3", got, cardIsSuspended(g, cardID))
	}
	added := requireExileTimeCounterEvent(t, g, game.EventCountersAdded, cardID, 0, 3)
	if card, _ := g.GetCardInstance(cardID); added.Controller != game.Player1 || added.CardZoneVersion != card.ZoneVersion {
		t.Fatalf("counters added event = %+v, want Player1 putting them on the exiled card", added)
	}
}

func TestSuspendUpkeepRemovesTimeCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(2, cost.Mana{cost.O(1)}), 2)

	engine.runSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})

	if got := g.ExileCounterCount(cardID, counter.Time); got != 1 || !cardIsSuspended(g, cardID) {
		t.Fatalf("time counters = %d, want still suspended with 1", got)
	}
	requireExileTimeCounterEvent(t, g, game.EventCountersRemoved, cardID, 2, 1)
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no cast yet", g.Stack.Size())
	}
}

func TestSuspendCastsSpellWhenLastCounterRemoved(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)

	engine.runSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("suspended card remained in exile after last counter")
	}
	if _, ok := g.CardCounters[cardID]; ok || cardIsSuspended(g, cardID) {
		t.Fatal("time counters remained after cast")
	}
	requireExileTimeCounterEvent(t, g, game.EventCountersRemoved, cardID, 1, 1)
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || !obj.Suspend {
		t.Fatalf("stack top = %+v, want suspended spell", obj)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == cardID &&
			event.FromZone == zone.Exile &&
			event.ToZone == zone.Stack &&
			len(event.Colors) == 1 &&
			event.Colors[0] == color.Green
	})
}

// TestSuspendUpkeepTriggersOrderedByOwner proves two suspended cards trigger
// two upkeep abilities that their owner puts on the stack in any order (CR
// 603.3b); the last one placed resolves first, removing that card's last time
// counter and triggering its cast ability before the other card's.
func TestSuspendUpkeepTriggersOrderedByOwner(t *testing.T) {
	for _, resolveSecondFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		firstID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
		secondID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
		want := []id.ID{firstID, secondID}
		if resolveSecondFirst {
			want = []id.ID{secondID, firstID}
		}

		agent := &suspendTriggerOrderAgent{resolveFirst: want[0]}
		engine.beginSuspendUpkeep(g, game.Player1, suspendAgents(agent))
		if agent.requests != 1 || g.Stack.Size() != 2 {
			t.Fatalf("order requests = %d stack = %d, want one ordering choice for two upkeep triggers", agent.requests, g.Stack.Size())
		}
		requireSuspendUpkeepTrigger(t, g, want[0])
		for !g.Stack.IsEmpty() {
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		}

		var castOrder []id.ID
		for _, event := range g.Events {
			if event.Kind == game.EventSpellCast {
				castOrder = append(castOrder, event.CardID)
			}
		}
		if !slices.Equal(castOrder, want) {
			t.Fatalf("cast order = %v, want the owner's chosen order %v", castOrder, want)
		}
	}
}

// suspendTriggerOrderAgent orders triggered abilities so the one from
// resolveFirst is placed last and so resolves first.
type suspendTriggerOrderAgent struct {
	resolveFirst id.ID
	requests     int
}

func (*suspendTriggerOrderAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a *suspendTriggerOrderAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceOrder {
		return request.DefaultSelection
	}
	a.requests++
	prefix := fmt.Sprintf("source=%v ", a.resolveFirst)
	order := make([]int, 0, len(request.Options))
	last := -1
	for i, option := range request.Options {
		if strings.HasPrefix(option.Label, prefix) {
			last = i
			continue
		}
		order = append(order, i)
	}
	return append(order, last)
}

func TestSuspendStateClearsWhenCardLeavesExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(2, cost.Mana{cost.O(1)}), 2)

	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) {
		t.Fatal("moving suspended card from exile to graveyard failed")
	}
	if _, ok := g.CardCounters[cardID]; ok || cardIsSuspended(g, cardID) {
		t.Fatal("time counters remained after card left exile")
	}
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("returning the card to exile failed")
	}
	if g.HasExileCounter(cardID, counter.Time) || cardIsSuspended(g, cardID) {
		t.Fatal("card returned to exile by another route is suspended")
	}
	engine := NewEngine(nil)
	engine.runSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})
	if g.Stack.Size() != 0 || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("unsuspended exiled card was cast at upkeep")
	}
}

// TestSuspendedCardRequiresSuspend proves CR 702.62b's "has suspend": an
// exiled card without suspend that has a time counter isn't suspended, and its
// counter isn't removed at its owner's upkeep.
func TestSuspendedCardRequiresSuspend(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	plain := suspendSorcery(1, cost.Mana{cost.O(1)})
	plain.StaticAbilities = nil
	cardID := addSuspendedCard(g, game.Player1, plain, 1)

	NewEngine(nil).runSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})

	if cardIsSuspended(g, cardID) || g.ExileCounterCount(cardID, counter.Time) != 1 || g.Stack.Size() != 0 {
		t.Fatal("exiled card without suspend was treated as suspended")
	}
}

// TestSuspendedCardCountersCloneIndependently proves a cloned game's suspended
// card keeps its own time counters.
func TestSuspendedCardCountersCloneIndependently(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(3, cost.Mana{cost.O(1)}), 3)
	clone := g.Clone()

	NewEngine(nil).runSuspendUpkeep(clone, game.Player1, [game.NumPlayers]PlayerAgent{})

	if got := clone.ExileCounterCount(cardID, counter.Time); got != 2 || !cardIsSuspended(clone, cardID) {
		t.Fatalf("clone time counters = %d, want 2", got)
	}
	if got := g.ExileCounterCount(cardID, counter.Time); got != 3 {
		t.Fatalf("original time counters = %d, want 3 after the clone's upkeep", got)
	}
}

// TestExileCardCounterEventsIgnoredByPermanentCounterTriggers proves abilities
// that observe counters put on or removed from permanents (CR 122.6) don't
// trigger on time counters put on or removed from a suspended card in exile.
func TestExileCardCounterEventsIgnoredByPermanentCounterTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	watcher := vanillaCreature("Counter Watcher", 1, 1)
	for _, event := range []game.EventKind{game.EventCountersAdded, game.EventCountersRemoved} {
		watcher.TriggeredAbilities = append(watcher.TriggeredAbilities, game.TriggeredAbility{
			Text: "Whenever a time counter is put on or removed from a creature, you gain 1 life.",
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:            event,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
				MatchCounterKind: true,
				CounterKind:      counter.Time,
			}},
			Content: game.AbilityContent{Modes: []game.Mode{{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
			}}}}},
		})
	}
	watcherPermanent := addCombatPermanent(g, game.Player1, watcher)
	cardID := addCardToHand(g, game.Player1, suspendCreature(2, cost.Mana{cost.G}))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.SuspendCard(cardID)) {
		t.Fatal("suspend action failed")
	}
	engine.runSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})
	requireExileTimeCounterEvent(t, g, game.EventCountersAdded, cardID, 0, 2)
	requireExileTimeCounterEvent(t, g, game.EventCountersRemoved, cardID, 2, 1)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no permanent-counter trigger for the exiled card", g.Stack.Size())
	}

	// Control: a time counter put on a creature permanent still triggers.
	if !addCountersToPermanent(g, watcherPermanent, counter.Time, 1) {
		t.Fatal("time counter was not put on the watcher")
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want the permanent-counter trigger for the creature permanent", g.Stack.Size())
	}
}

func TestSuspendedCreatureEntersWithSuspendHaste(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendCreature(1, cost.Mana{cost.O(1)}), 1)

	engine.runSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent := permanentByCardID(g, cardID)
	if permanent == nil {
		t.Fatal("suspended creature did not enter")
	}
	if !canAttackWith(g, permanent, game.Player1) {
		t.Fatal("suspended creature cannot attack despite suspend haste")
	}
}

func suspendSorcery(counters int, suspendCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Suspend Sorcery",
		Types:    []types.Card{types.Sorcery},
		Colors:   []color.Color{color.Green},
		ManaCost: opt.Val(cost.Mana{cost.O(9)}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.SuspendKeyword{Cost: suspendCost, TimeCounters: counters}},
		}}},
	}
}

func suspendCreature(counters int, suspendCost cost.Mana) *game.CardDef {
	pt := game.PT{Value: 2}
	card := suspendSorcery(counters, suspendCost)
	card.Name = "Suspend Creature"
	card.Types = []types.Card{types.Creature}
	card.Power = opt.Val(pt)
	card.Toughness = opt.Val(pt)
	return card
}

func addSuspendedCard(g *game.Game, playerID game.PlayerID, def *game.CardDef, counters int) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: playerID}
	g.Players[playerID].Exile.Add(cardID)
	g.AddExileCounter(cardID, counter.Time, counters)
	return cardID
}

// beginSuspendUpkeep begins an upkeep step of playerID's turn and puts the
// resulting beginning-of-upkeep triggers, including suspend's, on the stack.
func (e *Engine) beginSuspendUpkeep(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent) {
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = playerID, playerID
	g.Turn.Phase, g.Turn.Step = game.PhaseBeginning, game.StepUpkeep
	emitBeginningOfStepEvent(g, game.StepUpkeep)
	e.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
}

// runSuspendUpkeep begins playerID's upkeep and resolves suspend's upkeep and
// cast triggers from the top of the stack, putting each newly triggered
// ability on the stack before the next resolution. It stops at anything else,
// leaving any spell cast this way on the stack.
func (e *Engine) runSuspendUpkeep(g *game.Game, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent) {
	e.beginSuspendUpkeep(g, playerID, agents)
	e.resolveSuspendTriggers(g, agents)
}

func (e *Engine) resolveSuspendTriggers(g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	for {
		obj, ok := g.Stack.Peek()
		if !ok || (!obj.SuspendCast && !obj.SuspendUpkeep) {
			return
		}
		e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		e.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	}
}

// requireSuspendUpkeepTrigger asserts the top of the stack is suspend's
// upkeep ability for the exiled card, controlled by its owner.
func requireSuspendUpkeepTrigger(t *testing.T, g *game.Game, cardID id.ID) *game.StackObject {
	t.Helper()
	card, _ := g.GetCardInstance(cardID)
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility || !obj.SuspendUpkeep || obj.SuspendCast || obj.SourceCardID != cardID ||
		obj.Controller != card.Owner || obj.SourceZone != zone.Exile || obj.SourceZoneVersion != card.ZoneVersion {
		t.Fatalf("stack top = %+v, want the suspend upkeep trigger for %v", obj, cardID)
	}
	return obj
}

// requireSuspendCastTrigger asserts the top of the stack is the cast ability
// of suspend for the exiled card, controlled by its owner, before anything
// has been cast.
func requireSuspendCastTrigger(t *testing.T, g *game.Game, cardID id.ID) *game.StackObject {
	t.Helper()
	card, _ := g.GetCardInstance(cardID)
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility || !obj.SuspendCast || obj.SourceCardID != cardID ||
		obj.Controller != card.Owner || obj.SourceZone != zone.Exile || obj.SourceZoneVersion != card.ZoneVersion {
		t.Fatalf("stack top = %+v, want the suspend cast trigger for %v", obj, cardID)
	}
	if !g.Players[card.Owner].Exile.Contains(cardID) || len(suspendCastEvents(g, cardID)) != 0 {
		t.Fatal("suspended card was cast before its cast trigger resolved")
	}
	return obj
}

// suspendCastTriggerReady resolves suspend's upkeep trigger, removing the last
// time counter, and puts suspend's cast trigger on the stack.
func suspendCastTriggerReady(t *testing.T) (*Engine, *game.Game, id.ID, *game.StackObject) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
	engine.beginSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})
	requireSuspendUpkeepTrigger(t, g, cardID)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want only the suspend cast trigger", g.Stack.Size())
	}
	return engine, g, cardID, requireSuspendCastTrigger(t, g, cardID)
}

// Source review: CR 2026-11-06 (official TXT
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829),
// read 702.62a-d, 603.3b, 603.4, 108.4a and 400.7; Ancestral Vision rulings
// (https://api.scryfall.com/cards/9ec075ba-db56-4dcf-b1dc-fe6270b7ab36/rulings,
// 2024-02-02): the cast ability triggers however the last counter is removed,
// and if it's countered the card stays exiled with no time counters and is no
// longer suspended.

// TestSuspendCastTriggerUsesTheStack proves removing the last time counter
// triggers suspend's cast ability rather than casting during the upkeep turn
// action: players get priority with it on the stack, and it casts the card
// when it resolves.
func TestSuspendCastTriggerUsesTheStack(t *testing.T) {
	engine, g, cardID, _ := suspendCastTriggerReady(t)
	if g.ExileCounterCount(cardID, counter.Time) != 0 || cardIsSuspended(g, cardID) {
		t.Fatal("last time counter was not removed before the trigger")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	requireSuspendCast(t, g, cardID)
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want only the cast spell", g.Stack.Size())
	}
}

// TestStifledSuspendCastTriggerLeavesCardExiled counters the cast ability with
// Stifle: the card isn't cast, stays exiled with no time counters, isn't
// suspended and isn't offered again at a later upkeep.
func TestStifledSuspendCastTriggerLeavesCardExiled(t *testing.T) {
	engine, g, cardID, obj := suspendCastTriggerReady(t)
	stifleID := addCardToHand(g, game.Player2, registryCard(t, "Stifle"))
	g.Players[game.Player2].ManaPool.Add(mana.U, 1)
	g.Turn.PriorityPlayer = game.Player2
	if !engine.applyAction(g, game.Player2, action.CastSpell(stifleID, []game.Target{game.StackObjectTarget(obj.ID)}, 0, nil)) {
		t.Fatal("Stifle could not target the suspend cast trigger")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	requireSuspendNotCast(t, g, cardID)

	engine.runSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{})
	requireSuspendNotCast(t, g, cardID)
}

// TestSuspendCastTriggerAfterCardLeavesExile moves the card out of exile and
// back in response: the returned card is a new object (CR 400.7), so the
// trigger's "if it's exiled" check fails and nothing is cast (CR 603.4).
func TestSuspendCastTriggerAfterCardLeavesExile(t *testing.T) {
	engine, g, cardID, _ := suspendCastTriggerReady(t)
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) ||
		!moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
		t.Fatal("moving the card out of exile and back failed")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	requireSuspendNotCast(t, g, cardID)
}

// TestSuspendCastTriggerFromNonUpkeepRemoval proves the cast ability triggers
// however the last time counter is removed: an opponent's effect removing it
// in a main phase triggers the owner's ability.
func TestSuspendCastTriggerFromNonUpkeepRemoval(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(2, cost.Mana{cost.O(1)}), 2)
	setMainPhasePriority(g, game.Player2)

	change := newExileCardCounterChange(g, game.Player2)
	change.remove(cardID, counter.Time, 1)
	change.publish()
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Stack.Size() != 0 {
		t.Fatal("removing a time counter that wasn't the last triggered the cast ability")
	}
	change = newExileCardCounterChange(g, game.Player2)
	change.remove(cardID, counter.Time, 1)
	change.publish()
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	requireSuspendCastTrigger(t, g, cardID)
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	requireSuspendCast(t, g, cardID)
}

// TestSuspendCastTriggerClonesIndependently resolves the pending cast trigger
// in a clone: the original is unchanged until it resolves its own trigger,
// with the same result.
func TestSuspendCastTriggerClonesIndependently(t *testing.T) {
	engine, g, cardID, _ := suspendCastTriggerReady(t)
	clone := g.Clone()
	engine.resolveTopOfStackWithChoices(clone, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	requireSuspendCast(t, clone, cardID)
	requireSuspendCastTrigger(t, g, cardID)

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, replay := range []*game.Game{g, clone} {
		obj, _ := requireSuspendCast(t, replay, cardID)
		if replay.Stack.Size() != 1 || obj.ManaSpentToCast != 0 {
			t.Fatal("suspend cast trigger outcome differs between original and clone")
		}
	}
}

// TestSuspendCastTriggerThroughUpkeepStep runs the real upkeep step with
// passing players: the trigger goes on the stack in the upkeep priority window
// and the spell it casts resolves before the step ends.
func TestSuspendCastTriggerThroughUpkeepStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseBeginning
	g.TriggerEventCursor = len(g.Events)

	engine.runUpkeepStep(g, simPassPolicies(), &TurnLog{})

	if !g.Stack.IsEmpty() || len(suspendCastEvents(g, cardID)) != 1 || !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatalf("upkeep step did not resolve the suspend cast trigger and its spell: stack=%+v", g.Stack.Objects())
	}
}

// TestCopiedSuspendCastTriggerCastsOnce copies the cast ability (CR 707.10):
// the copy's controller is "you" and may cast the card from its owner's exile;
// the original then finds the card gone from exile and does nothing.
func TestCopiedSuspendCastTriggerCastsOnce(t *testing.T) {
	engine, g, cardID, obj := suspendCastTriggerReady(t)
	addEffectSpellToStack(g, game.Player2, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(obj.ID)})
	agents := [game.NumPlayers]PlayerAgent{}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || !copied.SuspendCast || copied.ID == obj.ID || copied.Controller != game.Player2 {
		t.Fatalf("stack top = %+v, want Player2's copy of the cast ability", copied)
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	spell, _ := requireSuspendCast(t, g, cardID)
	if spell.Controller != game.Player2 {
		t.Fatalf("suspended card cast by %v, want the copy's controller", spell.Controller)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !g.Stack.IsEmpty() || len(suspendCastEvents(g, cardID)) != 1 || !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatalf("stack = %+v; the original cast ability acted on a card that already left exile", g.Stack.Objects())
	}
}
