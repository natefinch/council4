package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/u"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestUnderworldBreachRealCardGrantsComputedEscape exercises the real,
// compiler-generated Underworld Breach end to end: with the enchantment on the
// battlefield, each nonland card in the controller's graveyard gains escape whose
// cost is the card's own mana cost plus exiling three other graveyard cards. This
// proves the generated RuleEffectGrantGraveyardCardKeyword + GraveyardCastCost
// flows through the runtime synthesis and payment planning exactly like the
// hand-crafted reference definition, keeping the registered card honest.
func TestUnderworldBreachRealCardGrantsComputedEscape(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, cards.UnderworldBreach())

	spell := &game.CardDef{CardFace: game.CardFace{
		Name:     "Grave Bolt",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.R}),
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
	}}
	cardID := addCardToGraveyard(g, game.Player1, spell)
	fuel := []id.ID{
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel One"}}),
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Two"}}),
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Three"}}),
	}
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("real Underworld Breach did not grant a payable escape cast from the graveyard")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("escaping card was not removed from the graveyard when cast")
	}
	for _, f := range fuel {
		if !g.Players[game.Player1].Exile.Contains(f) {
			t.Fatalf("fuel card %d was not exiled to pay the computed escape cost", f)
		}
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want granted escape graveyard cast", obj)
	}
	if obj.Flashback {
		t.Fatal("granted escape must not be marked flashback (escape does not exile on resolution)")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("granted escape card was exiled on resolution; escape returns it to the graveyard")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("granted escape card did not return to the graveyard after resolving")
	}
}

// TestUnderworldBreachRealCardExcludesLandCards proves the registered card's
// nonland selection is text-blind: a land card in the graveyard never gains
// escape, so it offers no graveyard cast option even with ample fuel present.
func TestUnderworldBreachRealCardExcludesLandCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addPermanentForSBA(g, game.Player1, cards.UnderworldBreach())

	landCard := &game.CardDef{CardFace: game.CardFace{
		Name:  "Escape Land",
		Types: []types.Card{types.Land},
	}}
	cardID := addCardToGraveyard(g, game.Player1, landCard)
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel One"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Two"}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fuel Three"}})
	setMainPhasePriority(g, game.Player1)

	if engine.applyAction(g, game.Player1, action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)) {
		t.Fatal("land card gained escape from Underworld Breach; nonland selection must exclude it")
	}
}
