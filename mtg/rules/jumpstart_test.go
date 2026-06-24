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

func jumpStartSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Jump-start Spell",
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(cost.Mana{cost.G}),
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability()),
		StaticAbilities: []game.StaticAbility{game.JumpStartStaticBody},
	}}
}

func TestJumpStartCastsFromGraveyardPayingDiscardAndExilesOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, jumpStartSpell())
	discardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discard Fuel"}})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	setMainPhasePriority(g, game.Player1)

	act := action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("jump-start cast from graveyard failed")
	}
	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("jump-start spell was not removed from the graveyard when cast")
	}
	if g.Players[game.Player1].Hand.Contains(discardID) {
		t.Fatal("jump-start did not discard a card as its additional cost")
	}
	if !g.Players[game.Player1].Graveyard.Contains(discardID) {
		t.Fatal("the discarded card did not reach the graveyard")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Flashback || obj.SourceZone != zone.Graveyard {
		t.Fatalf("stack object = %+v, want jump-start graveyard cast marked for exile", obj)
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("jump-start spell returned to the graveyard; it must be exiled on resolution")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("jump-start spell was not exiled after resolving")
	}
}

func TestJumpStartCannotBeCastFromGraveyardWithoutACardToDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToGraveyard(g, game.Player1, jumpStartSpell())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	setMainPhasePriority(g, game.Player1)

	act := action.CastSpellFromZone(cardID, zone.Graveyard, nil, 0, nil)
	if engine.applyAction(g, game.Player1, act) {
		t.Fatal("jump-start cast succeeded with an empty hand; the discard cost is unpayable")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("jump-start spell left the graveyard despite an unpayable discard cost")
	}
}
