package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func libraryGraveyardSelfPattern(pattern game.TriggerPattern) bool {
	return pattern.Event == game.EventZoneChanged &&
		pattern.Source == game.TriggerSourceSelf && !pattern.ExcludeSelf &&
		!pattern.SubjectSelectionOrSelf && !pattern.OneOrMore &&
		pattern.Player == game.TriggerPlayerYou &&
		pattern.MatchFromZone && pattern.FromZone == zone.Library &&
		!pattern.ExcludeFromZone && len(pattern.FromZones) == 0 &&
		pattern.MatchToZone && pattern.ToZone == zone.Graveyard && !pattern.ExcludeToZone
}

func captureLibraryGraveyardSourceTriggers(g *game.Game, event game.Event) []game.EventTriggeredAbility {
	if event.Kind != game.EventZoneChanged || event.FromZone != zone.Library || event.ToZone != zone.Graveyard ||
		event.CardID == 0 || event.CardZoneVersion == 0 || event.PermanentID != 0 {
		return nil
	}
	card, ok := g.GetCardInstance(event.CardID)
	if !ok || card.ID != event.CardID || card.Owner != event.Player || card.ZoneVersion != event.CardZoneVersion {
		return nil
	}
	if card.Owner < game.Player1 || card.Owner > game.Player4 {
		return nil
	}
	owner := g.Players[card.Owner]
	if owner == nil || owner.Eliminated || !owner.Graveyard.Contains(card.ID) {
		return nil
	}
	def, ok := cardFaceDef(card, game.FaceFront)
	if !ok {
		return nil
	}
	source := &game.Permanent{ObjectID: card.ID, CardInstanceID: card.ID, Owner: card.Owner, Controller: card.Owner}
	var captured []game.EventTriggeredAbility
	for i := range def.TriggeredAbilities {
		ability := &def.TriggeredAbilities[i]
		if ability.ZoneOfFunction != zone.None ||
			ability.Trigger.Type != game.TriggerWhen && ability.Trigger.Type != game.TriggerWhenever ||
			!libraryGraveyardSelfPattern(ability.Trigger.Pattern) ||
			!triggerMatchesEventForController(g, source, card.Owner, &ability.Trigger.Pattern, event) ||
			!triggerInterveningIf(g, source, card.Owner, &ability.Trigger, &event) {
			continue
		}
		captured = append(captured, game.EventTriggeredAbility{
			Controller: card.Owner, SourceID: card.ID, SourceCardID: card.ID,
			SourceZone: zone.Graveyard, SourceZoneVersion: event.CardZoneVersion,
			Face: game.FaceFront, AbilityIndex: i, Ability: ability,
			TriggerMultiplierCaptured: true,
		})
	}
	return captured
}
