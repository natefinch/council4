package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// CR 400.7f permits the original Aura's simultaneous departure or its subsequent
// unattached-Aura SBA departure, not any later graveyard incarnation.
func certifyAttachedDeathDeparture(g *game.Game, move *preparedPermanentZoneMove, unattachedSBA bool) {
	certifyCompletedAttachedDeathDeparture(g, move.event, move.snapshot, unattachedSBA)
}

func certifySimultaneousAttachedDeathDeparture(g *game.Game, event game.Event) {
	if event.SimultaneousID == 0 {
		return
	}
	if event.Kind == game.EventZoneChanged {
		if source, ok := lastKnownObject(g, event.PermanentID); ok {
			certifyCompletedAttachedDeathDeparture(g, event, source, false)
		}
	} else if event.Kind == game.EventPermanentDied {
		for _, departure := range g.Events {
			if departure.Kind != game.EventZoneChanged || departure.SimultaneousID != event.SimultaneousID {
				continue
			}
			if source, ok := lastKnownObject(g, departure.PermanentID); ok {
				certifyCompletedAttachedDeathDeparture(g, departure, source, false)
			}
		}
	}
}

func certifyCompletedAttachedDeathDeparture(g *game.Game, event game.Event, source game.ObjectSnapshot, unattachedSBA bool) {
	if event.Kind != game.EventZoneChanged || event.FromZone != zone.Battlefield ||
		event.ToZone != zone.Graveyard || event.CardID == 0 || event.CardZoneVersion == 0 ||
		event.PermanentID != source.ObjectID || event.CardID != source.CardID ||
		event.Player != source.Owner || source.TokenDef != nil ||
		!slices.Contains(source.Types, types.Enchantment) ||
		!slices.Contains(source.Subtypes, types.Aura) {
		return
	}
	for i, death := range g.Events {
		if i < g.TriggerEventCursor {
			continue
		}
		if death.Kind != game.EventPermanentDied || !death.TriggeredAbilitiesCaptured ||
			death.FromZone != zone.Battlefield || death.ToZone != zone.Graveyard ||
			death.PermanentID == 0 || death.CardZoneVersion == 0 {
			continue
		}
		if !unattachedSBA && (event.SimultaneousID == 0 || event.SimultaneousID != death.SimultaneousID) {
			continue
		}
		host, ok := lastKnownObject(g, death.PermanentID)
		if !ok || host.CardID != death.CardID || !slices.Contains(host.Attachments, event.PermanentID) {
			continue
		}
		for _, captured := range death.TriggeredAbilities {
			if captured.SourceID != event.PermanentID || captured.SourceCardID != event.CardID ||
				captured.Ability == nil || captured.Ability.Trigger.Pattern.Source != game.TriggerSourceAttachedPermanent {
				continue
			}
			g.RememberAttachedDeathDeparture(death.PermanentID, event.PermanentID, game.ZoneCardSnapshot{
				CardID: event.CardID, ZoneVersion: event.CardZoneVersion,
			}, event.Player)
		}
	}
}

func bindAttachedDeathDeparture(g *game.Game, trigger *pendingTriggeredAbility) {
	if !trigger.hasEvent || trigger.event.Kind != game.EventPermanentDied ||
		trigger.sourceZone != zone.None || trigger.inline == nil ||
		trigger.inline.Trigger.Pattern.Source != game.TriggerSourceAttachedPermanent {
		return
	}
	card, owner, ok := g.AttachedDeathDeparture(trigger.event.PermanentID, trigger.sourceID)
	source, sourceKnown := lastKnownObject(g, trigger.sourceID)
	host, hostKnown := lastKnownObject(g, trigger.event.PermanentID)
	if !ok || !sourceKnown || !hostKnown || source.CardID != trigger.sourceCardID ||
		card.CardID != trigger.sourceCardID || card.ZoneVersion == 0 || source.Owner != owner ||
		!slices.Contains(host.Attachments, trigger.sourceID) {
		return
	}
	trigger.sourceZone = zone.Graveyard
	trigger.sourceZoneVersion = card.ZoneVersion
}
