package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// graveyardRedirectCounterPermanent registers a Dauthi Voidwalker-style
// continuous redirect that exiles a card an opponent owns with a named counter
// on it ("If a card would be put into an opponent's graveyard from anywhere,
// instead exile it with a void counter on it.").
func graveyardRedirectCounterPermanent(g *game.Game, controller game.PlayerID, kind counter.Kind) *game.Permanent {
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Voidwalker",
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectExileWithCounterReplacement(
					"redirect",
					game.TriggerControllerOpponent,
					game.TriggerControllerAny,
					false,
					kind,
				),
			},
		},
	}
	permanent := addCombatPermanent(g, controller, def)
	registerPermanentReplacementEffects(g, permanent)
	return permanent
}

// TestGraveyardRedirectCounterExilesOpponentCardWithCounter verifies that an
// opponent-owned card headed to its owner's graveyard from hand is exiled with
// the named counter placed on it, while the redirect's controller's own card is
// left untouched (opponent scope) and never receives a counter.
func TestGraveyardRedirectCounterExilesOpponentCardWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	oppCard := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Their Card"}})
	if !moveCardBetweenZones(g, game.Player2, oppCard, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if g.Players[game.Player2].Graveyard.Contains(oppCard) {
		t.Fatal("redirect did not move the opponent's card away from the graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(oppCard) {
		t.Fatal("redirect did not exile the opponent's card")
	}
	if !g.HasExileCounter(oppCard, counter.Void) {
		t.Fatal("redirect exiled the opponent's card without a void counter")
	}

	ownCard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "My Card"}})
	if !moveCardBetweenZones(g, game.Player1, ownCard, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(ownCard) {
		t.Fatal("opponent-scoped redirect wrongly exiled the controller's own card")
	}
	if g.HasExileCounter(ownCard, counter.Void) {
		t.Fatal("opponent-scoped redirect placed a counter on the controller's own card")
	}
}

// TestGraveyardRedirectCounterExilesDyingPermanentWithCounter verifies that the
// counter rider is placed when the redirected card leaves the battlefield (a
// dying creature an opponent controls), covering the permanent-death move path.
func TestGraveyardRedirectCounterExilesDyingPermanentWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	permanent := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Dies", Types: []types.Card{types.Creature}},
	})
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if !g.Players[game.Player2].Exile.Contains(permanent.CardInstanceID) {
		t.Fatal("redirect did not exile the dying permanent")
	}
	if !g.HasExileCounter(permanent.CardInstanceID, counter.Void) {
		t.Fatal("redirect exiled the dying permanent without a void counter")
	}
}
