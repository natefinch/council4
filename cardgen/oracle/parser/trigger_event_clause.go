package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func emitTriggerEventClauses(abilities []Ability, cardName string) {
	for i := range abilities {
		trigger := abilities[i].Trigger
		if trigger == nil || trigger.PhaseStep != nil || trigger.PlayerEvent != nil {
			continue
		}
		trigger.TriggerEvent = parseTriggerEventClause(
			trigger.eventTokens,
			trigger.Introduction.Kind,
			abilities[i].Atoms,
			cardName,
		)
	}
}

func parseTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	cardName string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhen && intro != TriggerIntroductionWhenever {
		return nil
	}
	var matched *TriggerEventClause
	matchCount := 0
	for _, parse := range []func([]shared.Token, TriggerIntroductionKind, Atoms, string) *TriggerEventClause{
		parseZoneChangeTriggerEventClause,
		parseSpellCastTriggerEventClause,
		parseAbilityActivatedTriggerEventClause,
		parseAttackBlockTriggerEventClause,
		parseDamageTriggerEventClause,
		parseCounterTriggerEventClause,
		parsePermanentStateTriggerEventClause,
		parseTappedForManaTriggerEventClause,
		parseSacrificeTriggerEventClause,
		parseMutateTriggerEventClause,
		parseBecameTargetTriggerEventClause,
		parseTokenCreatedTriggerEventClause,
		parseTokenCreateSacrificeUnionTriggerEventClause,
		parseEnterAttackUnionTriggerEventClause,
		parseAttackBecameTargetUnionTriggerEventClause,
	} {
		clause := parse(tokens, intro, atoms, cardName)
		if clause == nil {
			continue
		}
		matchCount++
		matched = clause
	}
	if matchCount != 1 || matched == nil {
		return nil
	}
	matched.Span = shared.SpanOf(tokens)
	return matched
}

type zoneSubjectResult struct {
	subject       TriggerEventSubject
	controller    TriggerController
	player        TriggerPlayerSelector
	excludeSelf   bool
	faceDown      bool
	oneOrMore     bool
	selfOrAnother bool
	ok            bool
}

type permanentSubjectResult struct {
	subject     TriggerEventSubject
	controller  TriggerController
	excludeSelf bool
	oneOrMore   bool
	ok          bool
}
