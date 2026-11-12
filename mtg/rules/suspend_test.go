package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
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

	engine.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

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

	engine.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

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

func TestSuspendCastsMultipleReadyCardsInIDOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	firstID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
	secondID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)

	engine.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	var castOrder []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast {
			castOrder = append(castOrder, event.CardID)
		}
	}
	if len(castOrder) != 2 || castOrder[0] != firstID || castOrder[1] != secondID {
		t.Fatalf("cast order = %+v, want [%v %v]", castOrder, firstID, secondID)
	}
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
	engine.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
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

	NewEngine(nil).processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

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

	NewEngine(nil).processSuspendUpkeep(clone, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

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
	engine.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
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

	engine.processSuspendUpkeep(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
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
