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

// transmuteCard builds a hand-castable card whose sole activated ability is
// Transmute with the given activation mana cost, searching the library for a
// card whose mana value equals searchManaValue.
func transmuteCard(transmuteCost cost.Mana, searchManaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Transmute Test Card",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
		ActivatedAbilities: []game.ActivatedAbility{
			game.TransmuteActivatedAbility(transmuteCost, searchManaValue),
		},
	}}
}

func TestLegalActionsIncludesTransmuteFromHandAtSorcerySpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	transmuteID := addCardToHand(g, game.Player1, transmuteCard(cost.Mana{cost.O(1)}, 3))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	legal := engine.legalActions(g, game.Player1)

	if !actionsContain(legal, action.ActivateAbility(transmuteID, 0, nil, 0)) {
		t.Fatalf("legal actions = %+v, want transmute activation", legal)
	}
}

func TestTransmuteNotLegalAtInstantSpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	transmuteID := addCardToHand(g, game.Player1, transmuteCard(cost.Mana{cost.O(1)}, 3))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	g.Turn.PriorityPlayer = game.Player1

	legal := engine.legalActions(g, game.Player1)

	if actionsContain(legal, action.ActivateAbility(transmuteID, 0, nil, 0)) {
		t.Fatalf("legal actions = %+v, want no transmute activation at instant speed", legal)
	}
}

func TestTransmuteNotLegalOutsideHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	transmuteID := addCardToLibrary(g, game.Player1, transmuteCard(cost.Mana{cost.O(1)}, 3))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	legal := engine.legalActions(g, game.Player1)

	if actionsContain(legal, action.ActivateAbility(transmuteID, 0, nil, 0)) {
		t.Fatalf("legal actions = %+v, want no transmute activation from library", legal)
	}
}

func TestTransmuteDiscardsAndSearchesExactManaValueToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	transmuteID := addCardToHand(g, game.Player1, transmuteCard(cost.Mana{cost.O(1)}, 3))
	matchID := addCardToLibrary(g, game.Player1, manaValueCard("Exact Match", 3))
	wrongID := addCardToLibrary(g, game.Player1, manaValueCard("Wrong Value", 2))
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(transmuteID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for transmute")
	}
	if !forest.Tapped {
		t.Fatal("transmute mana cost did not tap available land")
	}
	if g.Players[game.Player1].Hand.Contains(transmuteID) {
		t.Fatal("transmuted card remained in hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(transmuteID) {
		t.Fatal("transmuted card was not discarded to graveyard")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 transmute ability", g.Stack.Size())
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackActivatedAbility || obj.SourceID != transmuteID || len(obj.AdditionalCostsPaid) != 1 {
		t.Fatalf("transmute stack object = %+v, want activated ability sourced from discarded card", obj)
	}
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.Event) bool {
		return event.Player == game.Player1 &&
			event.CardID == transmuteID &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Graveyard
	})

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Exact Match"}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(matchID) || g.Players[game.Player1].Library.Contains(matchID) {
		t.Fatal("transmute did not move the exact mana value card to hand")
	}
	if !g.Players[game.Player1].Library.Contains(wrongID) {
		t.Fatal("transmute moved a non-matching mana value card out of the library")
	}
}

// TestTransmuteWrongManaValueUnavailable proves the search cannot select a card
// whose mana value differs from the baked value even when the player asks for
// it; the only offered option is the matching card.
func TestTransmuteWrongManaValueUnavailable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	transmuteID := addCardToHand(g, game.Player1, transmuteCard(cost.Mana{cost.O(1)}, 3))
	wrongID := addCardToLibrary(g, game.Player1, manaValueCard("Wrong Value", 5))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(transmuteID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for transmute")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Wrong Value"}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(wrongID) {
		t.Fatal("transmute moved a non-matching mana value card to hand")
	}
	if !g.Players[game.Player1].Library.Contains(wrongID) {
		t.Fatal("transmute removed the non-matching card from the library")
	}
	if !g.Players[game.Player1].Graveyard.Contains(transmuteID) {
		t.Fatal("transmuted card should remain discarded after a failed search")
	}
}

// TestTransmuteSearchMayFailToFind proves the qualified search is allowed to
// fail: with no matching card the source stays discarded and nothing enters the
// hand.
func TestTransmuteSearchMayFailToFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	transmuteID := addCardToHand(g, game.Player1, transmuteCard(cost.Mana{cost.O(1)}, 3))
	onlyID := addCardToLibrary(g, game.Player1, manaValueCard("Only Card", 1))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(transmuteID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for transmute")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: ""}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Hand.Contains(onlyID) {
		t.Fatal("transmute pulled a non-matching card despite failing to find")
	}
	if !g.Players[game.Player1].Library.Contains(onlyID) {
		t.Fatal("transmute disturbed the library on a failed search")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size after resolution = %d, want 0", g.Stack.Size())
	}
	if !g.Players[game.Player1].Graveyard.Contains(transmuteID) {
		t.Fatal("transmuted card should remain discarded after a failed search")
	}
}

// TestTransmuteManaValueZeroBoundary proves the mana-value-0 boundary (Tolaria
// West): a Transmute that searches for mana value 0 pulls a zero-cost card such
// as a land to hand.
func TestTransmuteManaValueZeroBoundary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	transmuteID := addCardToHand(g, game.Player1, transmuteCard(cost.Mana{cost.O(1)}, 0))
	landID := addCardToLibrary(g, game.Player1, basicLandDef(types.Island))
	nonzeroID := addCardToLibrary(g, game.Player1, manaValueCard("Nonzero", 1))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(transmuteID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for transmute")
	}

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Island"}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(landID) || g.Players[game.Player1].Library.Contains(landID) {
		t.Fatal("transmute for mana value 0 did not move the land to hand")
	}
	if !g.Players[game.Player1].Library.Contains(nonzeroID) {
		t.Fatal("transmute for mana value 0 moved a nonzero card out of the library")
	}
}
