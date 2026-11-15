package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func madnessCostForCard(card *game.CardDef) (cost.Mana, bool) {
	if card == nil {
		return nil, false
	}
	return card.MadnessCost()
}

func (e *Engine) resolveMadnessTriggeredAbilityWithChoices(g *game.Game, obj *game.StackObject, ability *game.TriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	cardID := obj.SourceID
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return "missing source"
	}
	player, ok := playerByID(g, obj.Controller)
	if !ok || !player.Exile.Contains(cardID) {
		return "resolved"
	}
	manaCost, ok := game.BodyMadnessCost(ability)
	if !ok || !e.madnessCastAvailable(g, obj.Controller, card, manaCost) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
		return "resolved"
	}
	if !e.chooseMay(g, agents, obj.Controller, "Cast card for its madness cost?", log) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
		return "declined"
	}
	if !e.castMadnessSpellWithChoices(g, obj.Controller, card, manaCost, agents, log) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
	}
	return "resolved"
}

func madnessSpellRequest(player game.PlayerID, card *game.CardInstance, manaCost cost.Mana) payment.SpellRequest {
	return payment.SpellRequest{
		PlayerID: player, CardID: card.ID, SourceZone: zone.Exile, Card: cardFaceOrDefault(card, game.FaceFront),
		CastPermissions: []payment.SpellCastPermission{payment.SpellCastPermissionDefault},
		Alternative:     opt.Val(cost.Alternative{Label: "Madness", ManaCost: opt.Val(manaCost)}),
	}
}

func (e *Engine) madnessCastAvailable(g *game.Game, player game.PlayerID, card *game.CardInstance, manaCost cost.Mana) bool {
	request := madnessSpellRequest(player, card, manaCost)
	if !g.Players[player].Exile.Contains(card.ID) || castFromZoneProhibited(g, player, zone.Exile) ||
		!freeCastFaceCastable(g, player, card.ID, request.Card) {
		return false
	}
	object := &game.StackObject{Kind: game.StackSpell, SourceID: card.ID, Controller: player, SourceZone: zone.Exile, Casting: true}
	view := priorityCastView(g, object)
	cast := action.CastSpellAction{CardID: card.ID, SourceZone: zone.Exile}
	return len(e.resolutionSpellChoices(view, cast, request, true)) > 0
}

// Madness supplies its alternative cost to the same CR 601.2 procedure.
func (e *Engine) castMadnessSpellWithChoices(g *game.Game, playerID game.PlayerID, card *game.CardInstance, manaCost cost.Mana, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !e.madnessCastAvailable(g, playerID, card, manaCost) {
		return false
	}
	cast := action.CastSpellAction{CardID: card.ID, SourceZone: zone.Exile, Face: game.FaceFront}
	return e.resolutionSpell(g, g.Players[playerID], card, cast, madnessSpellRequest(playerID, card, manaCost), freeCastSpell{}, true, agents, log) == actionApplied
}

func firstLegalSpellCastChoice(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef) ([]int, []game.Target, bool) {
	if !isSupportedSpell(spellDef) {
		return nil, nil, false
	}
	for _, modes := range modeChoicesForSpellAt(g, playerID, spellDef) {
		targetResult := targetChoicesForSpell(g, playerID, spellDef, modes, game.CastBranch{})
		if targetResult.kind == targetInvalidSpec {
			continue
		}
		for _, targets := range targetResult.choices {
			if modesValidForSpellAt(g, playerID, spellDef, modes) && targetsValidForSpell(g, playerID, spellDef, modes, targets, game.CastBranch{}) {
				return append([]int(nil), modes...), append([]game.Target(nil), targets...), true
			}
		}
	}
	return nil, nil, false
}

func moveExiledCardToGraveyard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	return moveCardBetweenZonesInBatch(g, playerID, cardID, zone.Exile, zone.Graveyard, false, 0)
}
