package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func graveyardStepSourceTriggers(g *game.Game, event game.Event) []pendingTriggeredAbility {
	if event.Kind != game.EventBeginningOfStep {
		return nil
	}
	if event.Step != game.StepUpkeep && event.Step != game.StepBeginningOfCombat && event.Step != game.StepEnd {
		return nil
	}
	return graveyardSourceTriggers(g, event)
}

// captureGraveyardDeathSourceTriggers is called after replacements determine the
// destination, but before any permanent in the event moves (CR 603.6c, 603.10a).
func captureGraveyardDeathSourceTriggers(g *game.Game, move *preparedPermanentZoneMove) []game.EventTriggeredAbility {
	if move.actualDestination != zone.Graveyard {
		return nil
	}
	event := move.event
	event.Kind = game.EventPermanentDied
	event.ToZone = zone.Graveyard
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	return snapshotEventTriggeredAbilities(g, graveyardSourceTriggers(g, event))
}

func graveyardSourceTriggers(g *game.Game, event game.Event) []pendingTriggeredAbility {
	var pending []pendingTriggeredAbility
	for _, player := range g.Players {
		if player.Eliminated {
			continue
		}
		for _, cardID := range player.Graveyard.All() {
			card, ok := g.GetCardInstance(cardID)
			if !ok || card.ID != cardID || card.Owner != player.ID {
				continue
			}
			def, ok := cardFaceDef(card, game.FaceFront)
			if !ok {
				continue
			}
			source := &game.Permanent{ObjectID: card.ID, CardInstanceID: card.ID, Owner: card.Owner, Controller: card.Owner}
			for i := range def.TriggeredAbilities {
				ability := &def.TriggeredAbilities[i]
				if ability.ZoneOfFunction != zone.Graveyard || ability.Trigger.Pattern.Event != event.Kind ||
					!graveyardSourceTriggerTypeMatches(ability.Trigger.Type, event.Kind) ||
					!triggerMatchesEventForController(g, source, card.Owner, &ability.Trigger.Pattern, event) ||
					!triggerInterveningIf(g, source, card.Owner, &ability.Trigger, &event) {
					continue
				}
				pending = append(pending, pendingTriggeredAbility{
					controller: card.Owner, sourceID: card.ID, sourceCardID: card.ID,
					sourceZone: zone.Graveyard, sourceZoneVersion: card.ZoneVersion, face: game.FaceFront,
					abilityIndex: i, inline: ability, event: event, hasEvent: true,
					triggerMultiplierCaptured: true,
				})
			}
		}
	}
	return pending
}

func graveyardSourceTriggerTypeMatches(trigger game.TriggerType, event game.EventKind) bool {
	if event == game.EventPermanentDied {
		return trigger == game.TriggerWhen || trigger == game.TriggerWhenever
	}
	return trigger == game.TriggerAt && event == game.EventBeginningOfStep
}
