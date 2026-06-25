package rules

import (
	"github.com/natefinch/council4/mtg/game"
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
	if !e.chooseMay(g, agents, obj.Controller, "Cast card for its madness cost?", log) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
		return "declined"
	}
	manaCost, ok := game.BodyMadnessCost(ability)
	if !ok || !e.castMadnessSpellWithChoices(g, obj.Controller, card, manaCost, agents, log) {
		moveExiledCardToGraveyard(g, obj.Controller, cardID)
		return "resolved"
	}
	return "resolved"
}

func (e *Engine) castMadnessSpellWithChoices(g *game.Game, playerID game.PlayerID, card *game.CardInstance, manaCost cost.Mana, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Contains(card.ID) {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	modes, targets, ok := firstLegalSpellCastChoice(g, playerID, spellDef)
	if !ok {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, modes, targets)
	if !ok {
		panic("validated madness spell targets could not be segmented")
	}
	prefs := e.paymentPreferencesForCost(g, playerID, &manaCost, nil, 0, agents, log)
	riderSnapshot, _ := manaSpendRiderSnapshot(g, playerID)
	poolSpent, ok := paymentOrch.payGenericCostForSpell(g, payment.GenericRequest{
		PlayerID: playerID,
		Spell:    spellDef,
		Cost:     &manaCost,
		Prefs:    prefs,
	})
	if !ok {
		return false
	}
	if !player.Exile.Remove(card.ID) {
		return false
	}
	stackObj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     card.ID,
		Face:         game.FaceFront,
		Controller:   playerID,
		Targets:      append([]game.Target(nil), targets...),
		TargetCounts: targetCounts,
		ChosenModes:  append([]int(nil), modes...),
	}
	pushSpellToStack(g, stackObj, game.Event{
		SourceID:       card.ID,
		StackObjectID:  stackObj.ID,
		Controller:     playerID,
		CardID:         card.ID,
		CardTypes:      cardTypes(spellDef),
		CardSupertypes: cardSupertypes(spellDef),
		CardSubtypes:   cardSubtypes(spellDef),
		Colors:         spellColors(spellDef),
		ManaValue:      opt.Val(stackManaValue(spellDef, 0)),
		FromZone:       zone.Exile,
		ToZone:         zone.Stack,
	})
	stackObj.ColorsOfManaSpentToCast = distinctManaColorsSpent(poolSpent)
	resolveSpellCastManaSpendRiders(g, playerID, riderSnapshot, poolSpent, spellDef, stackObj)
	return true
}

func firstLegalSpellCastChoice(g *game.Game, playerID game.PlayerID, spellDef *game.CardDef) ([]int, []game.Target, bool) {
	if !isSupportedSpell(spellDef) {
		return nil, nil, false
	}
	for _, modes := range modeChoicesForSpellAt(g, playerID, spellDef) {
		targetResult := targetChoicesForSpell(g, playerID, spellDef, modes)
		if targetResult.kind == targetInvalidSpec {
			continue
		}
		for _, targets := range targetResult.choices {
			if modesValidForSpellAt(g, playerID, spellDef, modes) && targetsValidForSpell(g, playerID, spellDef, modes, targets) {
				return append([]int(nil), modes...), append([]game.Target(nil), targets...), true
			}
		}
	}
	return nil, nil, false
}

func moveExiledCardToGraveyard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Remove(cardID) {
		return false
	}
	player.Graveyard.Add(cardID)
	emitZoneChangeEvent(g, game.Event{
		Player:   playerID,
		CardID:   cardID,
		FromZone: zone.Exile,
		ToZone:   zone.Graveyard,
	})
	return true
}
