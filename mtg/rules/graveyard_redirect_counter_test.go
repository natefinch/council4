package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
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

// TestGraveyardRedirectCounterExilesDiscardedOpponentCardWithCounter drives a
// real discard through discardCardFromHandInBatch (an opponent discarding while
// the redirect is in play, e.g. Dauthi Voidwalker vs. a Mind Rot). "From
// anywhere" covers hand-to-graveyard discards, so the card must be exiled with a
// void counter and therefore be selectable by Dauthi's {T}, Sacrifice ability.
func TestGraveyardRedirectCounterExilesDiscardedOpponentCardWithCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectCounterPermanent(g, game.Player1, counter.Void)

	discarded := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	if !discardCardFromHand(g, game.Player2, discarded) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	if g.Players[game.Player2].Graveyard.Contains(discarded) {
		t.Fatal("redirect did not keep the discarded card out of the graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(discarded) {
		t.Fatal("redirect did not exile the discarded card")
	}
	if !g.HasExileCounter(discarded, counter.Void) {
		t.Fatal("redirect exiled the discarded card without a void counter")
	}

	// The discarded card now bears a void counter, so Dauthi's activated ability
	// must offer it and grant its controller a free-play permission bound to it.
	log := resolvePlayChosenExiledCard(t, g, game.PlayChosenExiledCard{
		Player:                game.ControllerReference(),
		Zone:                  zone.Exile,
		OwnerScope:            game.PlayerOpponent,
		Counter:               opt.Val(counter.Void),
		Duration:              game.DurationThisTurn,
		WithoutPayingManaCost: true,
	}, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	})
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want exactly one resolution choice", log.Choices)
	}
	if got := len(log.Choices[0].Request.Options); got != 1 {
		t.Fatalf("choice options = %d, want 1 (the discarded void-countered card)", got)
	}
	effect, ok := playFromZoneRuleEffect(g, discarded)
	if !ok {
		t.Fatal("Dauthi's ability granted no play permission for the discarded card")
	}
	if !effect.WithoutPayingManaCost {
		t.Fatal("granted permission is not flagged without paying its mana cost")
	}
	if !castFromZoneWithoutPayingManaCost(g, game.Player1, discarded, zone.Exile, game.FaceFront) {
		t.Fatal("controller should be able to play the discarded card without paying its mana cost")
	}
}
