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
	tokens = stripInterveningWhileCondition(tokens, atoms)
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
		parseEnterGraveyardUnionTriggerEventClause,
		parseAttackBecameTargetUnionTriggerEventClause,
		parseBlockBecameBlockedUnionTriggerEventClause,
		parseAttackBlockUnionTriggerEventClause,
		parseSelfGraveyardOrAnotherUnionTriggerEventClause,
		parseClassBecameLevelTriggerEventClause,
		parseDoorUnlockTriggerEventClause,
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

// stripInterveningWhileCondition removes a trailing "while <state>" clause from a
// trigger event phrase when the state is a recognized intervening condition
// ("Whenever you tap a land for mana while you're the monarch, ...", Regal
// Behemoth), so the event parsers see only the bare event and the condition is
// picked up separately by emitConditionClauses. Event-specific "while" riders
// ("attacks while saddled") are left in place, since conditionIntroAt does not
// treat them as condition clauses.
func stripInterveningWhileCondition(tokens []shared.Token, atoms Atoms) []shared.Token {
	for i := 1; i < len(tokens); i++ {
		if !equalWord(tokens[i], "while") {
			continue
		}
		if len(parseConditionClauses(tokens[i:], atoms)) == 0 {
			continue
		}
		return tokens[:i]
	}
	return tokens
}

type zoneSubjectResult struct {
	subject          TriggerEventSubject
	controller       TriggerController
	player           TriggerPlayerSelector
	excludeSelf      bool
	faceDown         bool
	oneOrMore        bool
	selfOrAnother    bool
	dealtDamageBySrc bool
	ok               bool
}

type permanentSubjectResult struct {
	subject     TriggerEventSubject
	controller  TriggerController
	excludeSelf bool
	oneOrMore   bool
	ok          bool
}
