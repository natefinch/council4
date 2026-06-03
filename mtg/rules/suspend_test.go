package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
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
	suspended := g.SuspendedCards[cardID]
	if suspended.Controller != game.Player1 || suspended.TimeCounters != 3 {
		t.Fatalf("suspended state = %+v, want controller P1 with 3 counters", suspended)
	}
}

func TestSuspendUpkeepRemovesTimeCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(2, cost.Mana{cost.O(1)}), 2)

	engine.processSuspendUpkeep(g, game.Player1)

	if got := g.SuspendedCards[cardID].TimeCounters; got != 1 {
		t.Fatalf("time counters = %d, want 1", got)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no cast yet", g.Stack.Size())
	}
}

func TestSuspendCastsSpellWhenLastCounterRemoved(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)

	engine.processSuspendUpkeep(g, game.Player1)

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("suspended card remained in exile after last counter")
	}
	if _, ok := g.SuspendedCards[cardID]; ok {
		t.Fatal("suspended state remained after cast")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || !obj.Suspend {
		t.Fatalf("stack top = %+v, want suspended spell", obj)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.GameEvent) bool {
		return event.CardID == cardID && event.FromZone == game.ZoneExile && event.ToZone == game.ZoneStack
	})
}

func TestSuspendCastsMultipleReadyCardsInIDOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	firstID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)
	secondID := addSuspendedCard(g, game.Player1, suspendSorcery(1, cost.Mana{cost.O(1)}), 1)

	engine.processSuspendUpkeep(g, game.Player1)

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

func TestSuspendedCreatureEntersWithSuspendHaste(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addSuspendedCard(g, game.Player1, suspendCreature(1, cost.Mana{cost.O(1)}), 1)

	engine.processSuspendUpkeep(g, game.Player1)
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
		ManaCost: opt.Val(cost.Mana{cost.O(9)}),
		Abilities: []game.AbilityDef{{
			Kind:                game.StaticAbility,
			Keywords:            []game.Keyword{game.Suspend},
			SuspendCost:         opt.Val(suspendCost),
			SuspendTimeCounters: counters,
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
	g.SuspendedCards[cardID] = game.SuspendedCard{Owner: playerID, Controller: playerID, TimeCounters: counters}
	return cardID
}
