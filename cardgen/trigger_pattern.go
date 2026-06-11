package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerTriggerPattern is the only semantic TriggerPattern to runtime
// game.TriggerPattern lowering path.
func lowerTriggerPattern(pattern *oracle.TriggerPattern) (game.TriggerPattern, bool) {
	event, ok := lowerTriggerEvent(pattern.Event)
	if !ok {
		return game.TriggerPattern{}, false
	}
	controller, ok := lowerTriggerController(pattern.Controller)
	if !ok {
		return game.TriggerPattern{}, false
	}
	player, ok := lowerTriggerPlayer(pattern.Player)
	if !ok {
		return game.TriggerPattern{}, false
	}
	source, ok := lowerTriggerSource(pattern.Source)
	if !ok {
		return game.TriggerPattern{}, false
	}
	subject, ok := lowerTriggerSubject(pattern.Subject)
	if !ok {
		return game.TriggerPattern{}, false
	}
	subjectSelection, ok := lowerTriggerSelection(pattern.SubjectSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	cardSelection, ok := lowerTriggerSelection(pattern.CardSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	damageSelection, ok := lowerTriggerSelection(pattern.DamageRecipientSelection)
	if !ok || !triggerDamageSelectionSupported(damageSelection) {
		return game.TriggerPattern{}, false
	}
	damageRecipient, ok := lowerTriggerDamageRecipient(pattern.DamageRecipient)
	if !ok {
		return game.TriggerPattern{}, false
	}
	step, ok := lowerTriggerStep(pattern.Step)
	if !ok {
		return game.TriggerPattern{}, false
	}
	result := game.TriggerPattern{
		Event:                event,
		Controller:           controller,
		Source:               source,
		ExcludeSelf:          pattern.ExcludeSelf,
		Player:               player,
		Subject:              subject,
		SubjectSelection:     subjectSelection,
		CardSelection:        cardSelection,
		DamageRecipient:      damageRecipient,
		DamageRecipientTypes: damageSelection.RequiredTypes,
		Step:                 step,
		OneOrMore:            pattern.OneOrMore,
		RequireKickerPaid:    pattern.RequireKickerPaid,
		RequireHistoric:      pattern.RequireHistoric,
	}
	if pattern.CombatQualifier == oracle.TriggerCombatDamage {
		result.RequireCombatDamage = true
	} else if pattern.CombatQualifier != oracle.TriggerCombatAny {
		return game.TriggerPattern{}, false
	}
	if pattern.StackObject == oracle.TriggerStackObjectSpell {
		result.MatchStackObjectKind = true
		result.StackObjectKind = game.StackSpell
	} else if pattern.StackObject != oracle.TriggerStackObjectAny {
		return game.TriggerPattern{}, false
	}
	if pattern.Counter != oracle.TriggerCounterAny {
		result.MatchCounterKind = true
		switch pattern.Counter {
		case oracle.TriggerCounterPlusOnePlusOne:
			result.CounterKind = counter.PlusOnePlusOne
		case oracle.TriggerCounterMinusOneMinusOne:
			result.CounterKind = counter.MinusOneMinusOne
		default:
			return game.TriggerPattern{}, false
		}
	}
	if pattern.MatchFromZone {
		result.FromZone, ok = lowerTriggerZone(pattern.FromZone)
		if !ok {
			return game.TriggerPattern{}, false
		}
		result.MatchFromZone = true
	} else if pattern.FromZone != oracle.TriggerZoneNone {
		return game.TriggerPattern{}, false
	}
	if pattern.MatchToZone {
		result.ToZone, ok = lowerTriggerZone(pattern.ToZone)
		if !ok {
			return game.TriggerPattern{}, false
		}
		result.MatchToZone = true
	} else if pattern.ToZone != oracle.TriggerZoneNone {
		return game.TriggerPattern{}, false
	}
	return result, true
}

func lowerTriggerKind(kind oracle.TriggerKind) (game.TriggerType, bool) {
	switch kind {
	case oracle.TriggerWhen:
		return game.TriggerWhen, true
	case oracle.TriggerWhenever:
		return game.TriggerWhenever, true
	case oracle.TriggerAt:
		return game.TriggerAt, true
	default:
		return 0, false
	}
}

func lowerTriggerEvent(event oracle.TriggerEvent) (game.EventKind, bool) {
	switch event {
	case oracle.TriggerEventSpellCast:
		return game.EventSpellCast, true
	case oracle.TriggerEventPermanentEnteredBattlefield:
		return game.EventPermanentEnteredBattlefield, true
	case oracle.TriggerEventPermanentDied:
		return game.EventPermanentDied, true
	case oracle.TriggerEventCountersAdded:
		return game.EventCountersAdded, true
	case oracle.TriggerEventDamageDealt:
		return game.EventDamageDealt, true
	case oracle.TriggerEventCardDrawn:
		return game.EventCardDrawn, true
	case oracle.TriggerEventAttackerDeclared:
		return game.EventAttackerDeclared, true
	case oracle.TriggerEventBlockerDeclared:
		return game.EventBlockerDeclared, true
	case oracle.TriggerEventCardDiscarded:
		return game.EventCardDiscarded, true
	case oracle.TriggerEventCycled:
		return game.EventCycled, true
	case oracle.TriggerEventBeginningOfStep:
		return game.EventBeginningOfStep, true
	case oracle.TriggerEventLifeGained:
		return game.EventLifeGained, true
	case oracle.TriggerEventLifeLost:
		return game.EventLifeLost, true
	case oracle.TriggerEventPermanentTapped:
		return game.EventPermanentTapped, true
	case oracle.TriggerEventPermanentUntapped:
		return game.EventPermanentUntapped, true
	case oracle.TriggerEventObjectBecameTarget:
		return game.EventObjectBecameTarget, true
	case oracle.TriggerEventPermanentMutated:
		return game.EventPermanentMutated, true
	case oracle.TriggerEventAttackerBecameBlocked:
		return game.EventAttackerBecameBlocked, true
	default:
		return game.EventUnknown, false
	}
}

func lowerTriggerController(controller oracle.ControllerKind) (game.TriggerControllerFilter, bool) {
	switch controller {
	case oracle.ControllerAny:
		return game.TriggerControllerAny, true
	case oracle.ControllerYou:
		return game.TriggerControllerYou, true
	case oracle.ControllerOpponent:
		return game.TriggerControllerOpponent, true
	default:
		return game.TriggerControllerAny, false
	}
}

func lowerTriggerPlayer(player oracle.TriggerPlayerRelation) (game.TriggerPlayerFilter, bool) {
	switch player {
	case oracle.TriggerPlayerAny:
		return game.TriggerPlayerAny, true
	case oracle.TriggerPlayerYou:
		return game.TriggerPlayerYou, true
	case oracle.TriggerPlayerOpponent:
		return game.TriggerPlayerOpponent, true
	default:
		return game.TriggerPlayerAny, false
	}
}

func lowerTriggerSource(source oracle.TriggerSourceRelation) (game.TriggerSourceFilter, bool) {
	switch source {
	case oracle.TriggerSourceAny:
		return game.TriggerSourceAny, true
	case oracle.TriggerSourceSelf:
		return game.TriggerSourceSelf, true
	case oracle.TriggerSourceAttachedPermanent:
		return game.TriggerSourceAttachedPermanent, true
	default:
		return game.TriggerSourceAny, false
	}
}

func lowerTriggerSubject(subject oracle.TriggerSubject) (game.TriggerSubjectObject, bool) {
	switch subject {
	case oracle.TriggerSubjectDefault:
		return game.TriggerSubjectDefault, true
	case oracle.TriggerSubjectPermanent:
		return game.TriggerSubjectPermanent, true
	case oracle.TriggerSubjectBlockedAttacker:
		return game.TriggerSubjectBlockedAttacker, true
	case oracle.TriggerSubjectDamageSource:
		return game.TriggerSubjectDamageSource, true
	default:
		return game.TriggerSubjectDefault, false
	}
}

func lowerTriggerDamageRecipient(recipient oracle.TriggerDamageRecipient) (game.DamageRecipientKind, bool) {
	switch recipient {
	case oracle.TriggerDamageRecipientAny:
		return game.DamageRecipientNone, true
	case oracle.TriggerDamageRecipientPlayer:
		return game.DamageRecipientPlayer, true
	case oracle.TriggerDamageRecipientPermanent:
		return game.DamageRecipientPermanent, true
	default:
		return game.DamageRecipientNone, false
	}
}

func lowerTriggerStep(step oracle.TriggerStep) (game.Step, bool) {
	switch step {
	case oracle.TriggerStepNone:
		return game.StepNone, true
	case oracle.TriggerStepUpkeep:
		return game.StepUpkeep, true
	case oracle.TriggerStepDraw:
		return game.StepDraw, true
	case oracle.TriggerStepBeginningOfCombat:
		return game.StepBeginningOfCombat, true
	case oracle.TriggerStepEndOfCombat:
		return game.StepEndOfCombat, true
	case oracle.TriggerStepEnd:
		return game.StepEnd, true
	case oracle.TriggerStepPrecombatMain:
		return game.StepPrecombatMain, true
	case oracle.TriggerStepPostcombatMain:
		return game.StepPostcombatMain, true
	default:
		return game.StepNone, false
	}
}

func lowerTriggerZone(triggerZone oracle.TriggerZone) (zone.Type, bool) {
	switch triggerZone {
	case oracle.TriggerZoneGraveyard:
		return zone.Graveyard, true
	case oracle.TriggerZoneBattlefield:
		return zone.Battlefield, true
	case oracle.TriggerZoneHand:
		return zone.Hand, true
	case oracle.TriggerZoneExile:
		return zone.Exile, true
	case oracle.TriggerZoneLibrary:
		return zone.Library, true
	case oracle.TriggerZoneStack:
		return zone.Stack, true
	default:
		return 0, false
	}
}

func lowerTriggerSelection(selection oracle.TriggerSelection) (game.Selection, bool) {
	required, ok := lowerTriggerCardTypes(selection.RequiredTypes)
	if !ok {
		return game.Selection{}, false
	}
	requiredAny, ok := lowerTriggerCardTypes(selection.RequiredTypesAny)
	if !ok {
		return game.Selection{}, false
	}
	excluded, ok := lowerTriggerCardTypes(selection.ExcludedTypes)
	if !ok {
		return game.Selection{}, false
	}
	subtypes, ok := lowerTriggerSubtypes(selection.SubtypesAny)
	if !ok {
		return game.Selection{}, false
	}
	colors, ok := lowerTriggerColors(selection.ColorsAny)
	if !ok {
		return game.Selection{}, false
	}
	result := game.Selection{
		RequiredTypes:    required,
		RequiredTypesAny: requiredAny,
		ExcludedTypes:    excluded,
		SubtypesAny:      subtypes,
		ColorsAny:        colors,
		Colorless:        selection.Colorless,
		Multicolored:     selection.Multicolored,
		NonToken:         selection.NonToken,
	}
	if selection.MatchManaValue {
		result.ManaValue = opt.Val(compare.Int{
			Op:    compare.GreaterOrEqual,
			Value: selection.ManaValueAtLeast,
		})
	} else if selection.ManaValueAtLeast != 0 {
		return game.Selection{}, false
	}
	return result, true
}

func lowerTriggerCardTypes(cardTypes []oracle.TriggerCardType) ([]types.Card, bool) {
	if len(cardTypes) == 0 {
		return nil, true
	}
	result := make([]types.Card, 0, len(cardTypes))
	for _, cardType := range cardTypes {
		switch cardType {
		case oracle.TriggerCardTypeArtifact:
			result = append(result, types.Artifact)
		case oracle.TriggerCardTypeBattle:
			result = append(result, types.Battle)
		case oracle.TriggerCardTypeCreature:
			result = append(result, types.Creature)
		case oracle.TriggerCardTypeEnchantment:
			result = append(result, types.Enchantment)
		case oracle.TriggerCardTypeInstant:
			result = append(result, types.Instant)
		case oracle.TriggerCardTypeLand:
			result = append(result, types.Land)
		case oracle.TriggerCardTypePlaneswalker:
			result = append(result, types.Planeswalker)
		case oracle.TriggerCardTypeSorcery:
			result = append(result, types.Sorcery)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerTriggerSubtypes(subtypes []oracle.TriggerSubtype) ([]types.Sub, bool) {
	if len(subtypes) == 0 {
		return nil, true
	}
	result := make([]types.Sub, 0, len(subtypes))
	for _, subtype := range subtypes {
		switch subtype {
		case oracle.TriggerSubtypeSpirit:
			result = append(result, types.Spirit)
		case oracle.TriggerSubtypeArcane:
			result = append(result, types.Arcane)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerTriggerColors(colors []oracle.TriggerColor) ([]color.Color, bool) {
	if len(colors) == 0 {
		return nil, true
	}
	result := make([]color.Color, 0, len(colors))
	for _, triggerColor := range colors {
		switch triggerColor {
		case oracle.TriggerColorWhite:
			result = append(result, color.White)
		case oracle.TriggerColorBlue:
			result = append(result, color.Blue)
		case oracle.TriggerColorBlack:
			result = append(result, color.Black)
		case oracle.TriggerColorRed:
			result = append(result, color.Red)
		case oracle.TriggerColorGreen:
			result = append(result, color.Green)
		default:
			return nil, false
		}
	}
	return result, true
}

func triggerDamageSelectionSupported(selection game.Selection) bool {
	return len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		len(selection.Supertypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		len(selection.ExcludedColors) == 0 &&
		!selection.Colorless &&
		!selection.Multicolored &&
		selection.Controller == game.ControllerAny &&
		selection.Player == game.PlayerAny &&
		selection.Tapped == game.TriAny &&
		selection.CombatState == game.CombatStateAny &&
		selection.Keyword == game.KeywordNone &&
		selection.ExcludedKeyword == game.KeywordNone &&
		!selection.ManaValue.Exists &&
		!selection.Power.Exists &&
		!selection.Toughness.Exists &&
		!selection.ExcludeSource &&
		!selection.NonToken &&
		!selection.TokenOnly
}
