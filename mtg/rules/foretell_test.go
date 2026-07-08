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

// foretellInstant is a minimal castable instant carrying the Foretell keyword
// with the given foretell cost. Its printed mana cost is high so a cast paying
// only the foretell cost is distinguishable from a normal cast in tests.
func foretellInstant(foretellCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         "Foretell Instant",
		Types:        []types.Card{types.Instant},
		ManaCost:     opt.Val(cost.Mana{cost.O(9)}),
		SpellAbility: opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.ForetellKeyword{Cost: foretellCost}},
		}},
	}}
}

// TestLegalActionsIncludeForetellFromHand verifies a hand card with the Foretell
// keyword offers a foretell special action when its {2} cost can be paid.
func TestLegalActionsIncludeForetellFromHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, foretellInstant(cost.Mana{cost.U}))
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Island)
	setSorceryTiming(g, game.Player1)

	legal := engine.legalActions(g, game.Player1)

	if !actionsContain(legal, action.ForetellCard(cardID)) {
		t.Fatalf("legal actions = %+v, want foretell action", legal)
	}
}

// TestForetellOfferedOutsideSorcerySpeed verifies foretelling is available any
// time the player has priority during their turn, not only at sorcery speed
// (CR 702.144a): it is offered during the player's combat phase.
func TestForetellOfferedOutsideSorcerySpeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, foretellInstant(cost.Mana{cost.U}))
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Island)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers

	if isSorcerySpeed(g, game.Player1) {
		t.Fatal("precondition: timing should not be sorcery speed")
	}
	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.ForetellCard(cardID)) {
		t.Fatalf("legal actions = %+v, want foretell action at instant speed", legal)
	}
}

// TestForetellActionPaysTwoAndExiles verifies foretelling pays the fixed {2},
// moves the card from hand to exile, and records the foretell turn.
func TestForetellActionPaysTwoAndExiles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, foretellInstant(cost.Mana{cost.U}))
	island1 := addBasicLandPermanent(g, game.Player1, types.Island)
	island2 := addBasicLandPermanent(g, game.Player1, types.Island)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3

	if !engine.applyAction(g, game.Player1, action.ForetellCard(cardID)) {
		t.Fatal("foretell action failed")
	}
	if !island1.Tapped || !island2.Tapped {
		t.Fatal("foretell cost did not tap two mana sources")
	}
	if g.Players[game.Player1].Hand.Contains(cardID) || !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("foretold card did not move from hand to exile")
	}
	if foretoldTurn, ok := g.ForetoldCards[cardID]; !ok || foretoldTurn != 3 {
		t.Fatalf("ForetoldCards[card] = (%d,%v), want (3,true)", foretoldTurn, ok)
	}
}

// TestForetoldCardNotCastableSameTurn verifies a card foretold this turn cannot
// be cast from exile until a later turn (CR 702.144b).
func TestForetoldCardNotCastableSameTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, foretellInstant(cost.Mana{cost.U}))
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Island)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3

	if !engine.applyAction(g, game.Player1, action.ForetellCard(cardID)) {
		t.Fatal("foretell action failed")
	}
	if cardIsForetoldInExile(g, cardID) {
		t.Fatal("card is castable from exile the same turn it was foretold")
	}
	legal := engine.legalActions(g, game.Player1)
	if actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("cast-from-exile offered the same turn the card was foretold")
	}
}

// TestForetoldCardCastableForForetellCostLaterTurn verifies a foretold card is
// castable from exile on a later turn for its foretell cost (not free and not its
// full mana cost), clearing the ForetoldCards entry once cast.
func TestForetoldCardCastableForForetellCostLaterTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, foretellInstant(cost.Mana{cost.U}))
	addBasicLandPermanent(g, game.Player1, types.Island)
	addBasicLandPermanent(g, game.Player1, types.Island)
	castIsland := addBasicLandPermanent(g, game.Player1, types.Island)
	setSorceryTiming(g, game.Player1)
	g.Turn.TurnNumber = 3

	if !engine.applyAction(g, game.Player1, action.ForetellCard(cardID)) {
		t.Fatal("foretell action failed")
	}

	// Advance to a later turn. One Island remains untapped, enough to pay the
	// {U} foretell cost but nowhere near the printed {9} mana cost.
	g.Turn.TurnNumber = 4
	setSorceryTiming(g, game.Player1)
	if castIsland.Tapped {
		t.Fatal("precondition: an Island should remain untapped for the foretell cast")
	}

	if !cardIsForetoldInExile(g, cardID) {
		t.Fatal("card is not castable from exile on a later turn")
	}
	legal := engine.legalActions(g, game.Player1)
	if !actionsContain(legal, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatalf("legal actions = %+v, want foretell cast-from-exile of the foretold card", legal)
	}

	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(cardID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("casting the foretold card from exile failed")
	}
	if _, onStack := g.Stack.Peek(); !onStack {
		t.Fatal("foretold spell was not put on the stack")
	}
	if !castIsland.Tapped {
		t.Fatal("foretell cast did not pay the {U} foretell cost")
	}
	if _, ok := g.ForetoldCards[cardID]; ok {
		t.Fatal("ForetoldCards entry not cleared after casting the foretold card")
	}
}
