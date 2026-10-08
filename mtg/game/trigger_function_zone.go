package game

import "github.com/natefinch/council4/mtg/game/zone"

func validTriggeredSourceZone(ability *TriggeredAbility) bool {
	if ability.ZoneOfFunction == zone.None || ability.ZoneOfFunction == zone.Battlefield {
		return true
	}
	pattern := ability.Trigger.Pattern
	return ability.ZoneOfFunction == zone.Graveyard && ability.Trigger.Type == TriggerAt &&
		pattern.Event == EventBeginningOfStep &&
		(pattern.Step == StepUpkeep || pattern.Step == StepBeginningOfCombat || pattern.Step == StepEnd) &&
		pattern.Source == TriggerSourceAny && pattern.Subject == TriggerSubjectDefault &&
		!pattern.StepPlayerIsSourceEnchantedPlayer && pattern.StepPlayerSourceAttachedSelection.Empty()
}
