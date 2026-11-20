package game

import "github.com/natefinch/council4/mtg/game/zone"

func validTriggeredSourceZone(ability *TriggeredAbility) bool {
	if ability.ZoneOfFunction == zone.None || ability.ZoneOfFunction == zone.Battlefield {
		return true
	}
	pattern := ability.Trigger.Pattern
	if ability.ZoneOfFunction != zone.Graveyard ||
		pattern.Source != TriggerSourceAny || pattern.Subject != TriggerSubjectDefault ||
		pattern.StepPlayerIsSourceEnchantedPlayer || !pattern.StepPlayerSourceAttachedSelection.Empty() {
		return false
	}
	switch pattern.Event {
	case EventBeginningOfStep:
		return ability.Trigger.Type == TriggerAt &&
			(pattern.Step == StepUpkeep || pattern.Step == StepBeginningOfCombat || pattern.Step == StepEnd)
	case EventPermanentDied:
		return ability.Trigger.Type == TriggerWhen || ability.Trigger.Type == TriggerWhenever
	default:
		return false
	}
}
