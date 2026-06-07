package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDelveMakesGenericSpellPayableAndExilesGraveyardCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, delveSpell(cost.Mana{cost.O(2)}))
	first := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Graveyard Card"}})
	second := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Graveyard Card"}})
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("delve spell cast failed")
	}
	if g.Players[game.Player1].Graveyard.Contains(first) || g.Players[game.Player1].Graveyard.Contains(second) {
		t.Fatal("delve cards remained in graveyard")
	}
	if !g.Players[game.Player1].Exile.Contains(first) || !g.Players[game.Player1].Exile.Contains(second) {
		t.Fatal("delve cards did not move to exile")
	}
}

func TestDelveDoesNotExileWhenManaCanPay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, delveSpell(cost.Mana{cost.O(1)}))
	graveID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Graveyard Card"}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("delve spell cast failed")
	}
	if !g.Players[game.Player1].Graveyard.Contains(graveID) {
		t.Fatal("delve exiled graveyard card even though mana could pay")
	}
}

func TestDelveExilesOnlyCardsNeededAfterAvailableMana(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, delveSpell(cost.Mana{cost.O(2)}))
	first := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Graveyard Card"}})
	second := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Graveyard Card"}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("delve spell cast failed")
	}
	if !g.Players[game.Player1].Exile.Contains(second) {
		t.Fatal("top graveyard card was not exiled for delve")
	}
	if !g.Players[game.Player1].Graveyard.Contains(first) {
		t.Fatal("delve exiled more graveyard cards than needed")
	}
}

func TestDelveIgnoresCardsOutsideGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, delveSpell(cost.Mana{cost.O(1)}))
	exileID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Exiled Card"}})
	g.Players[game.Player1].Graveyard.Remove(exileID)
	g.Players[game.Player1].Exile.Add(exileID)
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("delve spell cast using non-graveyard card, want failure")
	}
}

func TestDelveCanPayXGenericCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, delveSpell(cost.Mana{cost.X}))
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Graveyard Card"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Graveyard Card"}})
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 2, nil)) {
		t.Fatal("delve spell with X cost failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.XValue != 2 {
		t.Fatalf("stack object = %+v, want XValue 2", obj)
	}
}

func TestDelvePaymentExcludesSourceCardFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := delveSpell(cost.Mana{cost.O(1)})
	sourceID := addCardToGraveyard(g, game.Player1, card)

	if canPayTestSpellCosts(g, testSpellPaymentRequest{playerID: game.Player1, cardID: sourceID, sourceZone: zone.Graveyard, card: card}) {
		t.Fatal("canPaySpellCosts() = true using source card for delve, want false")
	}
}

func delveSpell(manaCost cost.Mana) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Delve Spell",
		Types:           []types.Card{types.Sorcery},
		ManaCost:        opt.Val(manaCost),
		SpellAbility:    opt.Val(game.AbilityContent{}),
		StaticAbilities: []game.StaticAbility{game.DelveStaticBody}},
	}
}

func addCardToGraveyard(g *game.Game, playerID game.PlayerID, def *game.CardDef) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: playerID}
	g.Players[playerID].Graveyard.Add(cardID)
	return cardID
}
