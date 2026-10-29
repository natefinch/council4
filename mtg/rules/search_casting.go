package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func synchronousSearchOrPayment(g *game.Game) bool {
	return len(g.SearchControlScopes) != 0 || len(g.ManaPaymentScopes) != 0
}

func searchCastCardAllowed(g *game.Game, player game.PlayerID, card *game.CardInstance) bool {
	if card == nil || card.Owner != player || !g.SearchCastingAllowed(player) || !g.Players[player].Library.Contains(card.ID) {
		return false
	}
	def := cardFaceOrDefault(card, game.FaceFront)
	for _, ability := range def.StaticAbilities {
		if ability.CastWhileSearchingOwnLibrary && ability.ZoneOfFunction == zone.Library &&
			conditionSatisfied(g, conditionContext{controller: player}, ability.Condition) {
			castFace := castSelectionFace(g, player, def, false)
			return cardCastRestrictionsSatisfied(g, player, def) && !spellCastProhibited(g, player, castFace) &&
				!spellCastLimitReached(g, player, castFace) && !castFromZoneProhibited(g, player, zone.Library) && isSupportedSpell(def)
		}
	}
	return false
}

func (e *Engine) offerSearchCasts(g *game.Game, player game.PlayerID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	scope := len(g.SearchControlScopes) - 1
	if scope < 0 || g.SearchControlScopes[scope].Player != player {
		panic("search casting requires an active own-library search")
	}
	g.SearchControlScopes[scope].BeforeFind = true
	defer func() { g.SearchControlScopes[scope].BeforeFind = false }()
	for {
		options := []game.ChoiceOption{{Index: 0, Label: "Continue finding cards"}}
		for _, cardID := range g.Players[player].Library.All() {
			card, _ := g.GetCardInstance(cardID)
			if searchCastCardAllowed(g, player, card) {
				options = append(options, game.ChoiceOption{Index: len(options), Label: "Cast " + card.Def.Name, Card: cardChoiceInfo(g, cardID)})
			}
		}
		if len(options) == 1 {
			return
		}
		selected := e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind: game.ChoiceMay, Player: player, Prompt: "Cast a card before finding cards?",
			Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0}, Private: true,
		}, log)
		if selected[0] == 0 {
			return
		}
		cardID := options[selected[0]].Card.Val.CardID
		card, _ := g.GetCardInstance(cardID)
		if !searchCastCardAllowed(g, player, card) ||
			!e.castPaidSpellFromSource(g, g.Players[player], player, cardID, zone.Library, agents, log) {
			return
		}
	}
}
