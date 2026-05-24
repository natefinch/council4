package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

type damageEvent struct {
	sourceID       id.ID
	sourceObjectID id.ID
	controller     game.PlayerID
	player         game.PlayerID
	permanent      *game.Permanent
	amount         int
	combatDamage   bool
}

func applyDamagePrevention(g *game.Game, event damageEvent) int {
	if g == nil || event.amount <= 0 {
		return 0
	}
	amount := event.amount
	if event.permanent != nil && permanentProtectedFromSource(g, event.permanent, event.sourceID, event.sourceObjectID) {
		amount = 0
	}
	if amount > 0 && event.permanent != nil && event.permanent.Counters.Remove(counter.Shield, 1) > 0 {
		amount = 0
	}
	if prevented := event.amount - amount; prevented > 0 {
		emitDamagePreventedEvent(g, event, prevented)
	}
	return amount
}

func replaceDestroyPermanent(g *game.Game, permanent *game.Permanent) bool {
	if g == nil || permanent == nil {
		return false
	}
	if permanent.Counters.Remove(counter.Shield, 1) == 0 {
		return false
	}
	emitEvent(g, game.GameEvent{
		Kind:        game.EventDestroyReplaced,
		Controller:  permanent.Controller,
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
		FromZone:    game.ZoneBattlefield,
		ToZone:      game.ZoneGraveyard,
	})
	return true
}

func emitDamagePreventedEvent(g *game.Game, event damageEvent, prevented int) {
	preventedEvent := game.GameEvent{
		Kind:            game.EventDamagePrevented,
		SourceID:        event.sourceID,
		SourceObjectID:  event.sourceObjectID,
		Controller:      event.controller,
		Player:          event.player,
		Amount:          prevented,
		DamageRecipient: game.DamageRecipientPlayer,
		CombatDamage:    event.combatDamage,
	}
	if event.permanent != nil {
		preventedEvent.Player = event.permanent.Owner
		preventedEvent.PermanentID = event.permanent.ObjectID
		preventedEvent.CardID = event.permanent.CardInstanceID
		preventedEvent.TokenName = permanentTokenName(event.permanent)
		preventedEvent.TokenDef = event.permanent.TokenDef
		preventedEvent.DamageRecipient = game.DamageRecipientPermanent
	}
	emitEvent(g, preventedEvent)
}

func permanentProtectedFromSource(g *game.Game, permanent *game.Permanent, sourceID, sourceObjectID id.ID) bool {
	source := damageSourceDef(g, sourceID, sourceObjectID)
	return permanentProtectedFromSourceDef(g, permanent, source)
}

func permanentProtectedFromSourceDef(g *game.Game, permanent *game.Permanent, source *game.CardDef) bool {
	if permanent == nil || source == nil {
		return false
	}
	for _, color := range permanentProtectionColors(g, permanent) {
		if slices.Contains(source.Colors, color) {
			return true
		}
	}
	return false
}

func permanentProtectionColors(g *game.Game, permanent *game.Permanent) []mana.Color {
	def := permanentCardDef(g, permanent)
	if def == nil {
		return nil
	}
	var colors []mana.Color
	for i := range def.Abilities {
		ability := &def.Abilities[i]
		if !abilityHasKeyword(ability, game.Protection) {
			continue
		}
		colors = append(colors, ability.ProtectionFromColors...)
	}
	return colors
}

func damageSourceDef(g *game.Game, sourceID, sourceObjectID id.ID) *game.CardDef {
	if g == nil {
		return nil
	}
	if sourceID != 0 {
		if card := g.GetCardInstance(sourceID); card != nil {
			return card.Def
		}
	}
	return permanentCardDef(g, permanentByObjectID(g, sourceObjectID))
}
