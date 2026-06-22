package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func graveyardRedirectPermanent(
	g *game.Game,
	controller game.PlayerID,
	ownerFilter game.TriggerControllerFilter,
	fromBattlefieldOnly bool,
	cardTypes ...types.Card,
) *game.Permanent {
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Voidkeeper",
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectReplacement("redirect", ownerFilter, game.TriggerControllerAny, fromBattlefieldOnly, cardTypes...),
			},
		},
	}
	permanent := addCombatPermanent(g, controller, def)
	registerPermanentReplacementEffects(g, permanent)
	return permanent
}

func TestGraveyardRedirectExilesAnyPlayersCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false)
	cardID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Doomed Card"}})

	if !moveCardBetweenZones(g, game.Player2, cardID, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if g.Players[game.Player2].Graveyard.Contains(cardID) {
		t.Fatal("redirect did not move card away from graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(cardID) {
		t.Fatal("redirect did not exile the card")
	}
}

func TestGraveyardRedirectOpponentScopeSparesController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerOpponent, false)
	ownCard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "My Card"}})

	if !moveCardBetweenZones(g, game.Player1, ownCard, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(ownCard) {
		t.Fatal("opponent-scoped redirect wrongly exiled the controller's own card")
	}

	oppCard := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Their Card"}})
	if !moveCardBetweenZones(g, game.Player2, oppCard, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if !g.Players[game.Player2].Exile.Contains(oppCard) {
		t.Fatal("opponent-scoped redirect did not exile the opponent's card")
	}
}

func TestGraveyardRedirectTypeFilterSkipsNonmatchingCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, false, types.Instant, types.Sorcery)

	creatureID := addCardToHand(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Beast", Types: []types.Card{types.Creature}},
	})
	if !moveCardBetweenZones(g, game.Player1, creatureID, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(creatureID) {
		t.Fatal("type-filtered redirect wrongly exiled a creature card")
	}

	instantID := addCardToHand(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Bolt", Types: []types.Card{types.Instant}},
	})
	if !moveCardBetweenZones(g, game.Player1, instantID, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(instantID) {
		t.Fatal("type-filtered redirect did not exile a matching instant card")
	}
}

func TestGraveyardRedirectBattlefieldOnlySparesHandDiscard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardRedirectPermanent(g, game.Player1, game.TriggerControllerAny, true)

	handCard := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discarded"}})
	if !moveCardBetweenZones(g, game.Player1, handCard, zone.Hand, zone.Graveyard) {
		t.Fatal("moveCardBetweenZones() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(handCard) {
		t.Fatal("battlefield-only redirect wrongly exiled a card leaving the hand")
	}

	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Dies", Types: []types.Card{types.Creature}},
	})
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(permanent.CardInstanceID) {
		t.Fatal("battlefield-only redirect did not exile a permanent leaving the battlefield")
	}
}

func graveyardDeathRedirectPermanent(
	g *game.Game,
	controller game.PlayerID,
	controlFilter game.TriggerControllerFilter,
	cardTypes ...types.Card,
) *game.Permanent {
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Voidwatcher",
			ReplacementAbilities: []game.ReplacementAbility{
				game.GraveyardRedirectReplacement("redirect", game.TriggerControllerAny, controlFilter, true, cardTypes...),
			},
		},
	}
	permanent := addCombatPermanent(g, controller, def)
	registerPermanentReplacementEffects(g, permanent)
	return permanent
}

func TestGraveyardRedirectControlScopeExilesOpponentsDyingCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardDeathRedirectPermanent(g, game.Player1, game.TriggerControllerOpponent, types.Creature)

	ownCreature := addCombatPermanent(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Mine", Types: []types.Card{types.Creature}},
	})
	if !movePermanentToZone(g, ownCreature, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(ownCreature.CardInstanceID) {
		t.Fatal("opponent-control redirect wrongly exiled the controller's own dying creature")
	}

	oppCreature := addCombatPermanent(g, game.Player2, &game.CardDef{
		CardFace: game.CardFace{Name: "Theirs", Types: []types.Card{types.Creature}},
	})
	if !movePermanentToZone(g, oppCreature, zone.Graveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	if g.Players[game.Player2].Graveyard.Contains(oppCreature.CardInstanceID) {
		t.Fatal("opponent-control redirect did not move the dying creature away from the graveyard")
	}
	if !g.Players[game.Player2].Exile.Contains(oppCreature.CardInstanceID) {
		t.Fatal("opponent-control redirect did not exile the opponent's dying creature")
	}
}
