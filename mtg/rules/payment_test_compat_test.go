package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/rules/payment"
)

type testSpellPaymentRequest struct {
	playerID   game.PlayerID
	cardID     id.ID
	sourceZone game.ZoneType
	card       *game.CardDef
	xValue     int
	kickerPaid bool
	prefs      *payment.Preferences
}

func canPayTestSpellCosts(g *game.Game, req testSpellPaymentRequest) bool {
	return paymentOrch.canPaySpellCosts(g, payment.SpellRequest{
		PlayerID:   req.playerID,
		CardID:     req.cardID,
		SourceZone: req.sourceZone,
		Card:       req.card,
		XValue:     req.xValue,
		KickerPaid: req.kickerPaid,
		Prefs:      req.prefs,
	})
}

func payTestGenericCost(g *game.Game, playerID game.PlayerID, cost *mana.Cost) bool {
	return paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: cost})
}
