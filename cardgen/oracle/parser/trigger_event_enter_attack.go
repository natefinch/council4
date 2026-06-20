package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseEnterAttackUnionTriggerEventClause recognizes the event-union trigger
// "<subject> enters or attacks" (and its mirror "<subject> attacks or enters"),
// e.g. "Whenever this creature enters or attacks". Both verbs share a single
// subject, so the trigger fires when either the enters-the-battlefield event or
// the attack event happens (CR 603.2). The enter event is the primary clause
// and the attack event joins it through UnionKind.
func parseEnterAttackUnionTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	for i := 1; i+2 < len(tokens); i++ {
		first, or, second := tokens[i], tokens[i+1], tokens[i+2]
		if !equalWord(or, "or") {
			continue
		}
		isEnterAttack := equalWord(first, "enters") && equalWord(second, "attacks")
		isAttackEnter := equalWord(first, "attacks") && equalWord(second, "enters")
		if !isEnterAttack && !isAttackEnter {
			continue
		}
		if i+3 != len(tokens) {
			return nil
		}
		subject := parsePermanentEventSubject(tokens[:i], false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		eventSpan := shared.SpanOf(tokens[i:])
		return &TriggerEventClause{
			Kind:        TriggerEventKindZoneChange,
			UnionKind:   TriggerEventKindAttack,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			ZoneChange: TriggerEventZoneChange{
				Kind: TriggerEventZoneChangeEnteredBattlefield,
				Span: eventSpan,
			},
			Zone: TriggerEventZoneContext{
				Span:        eventSpan,
				MatchToZone: true,
				ToZone:      triggerEventZone(TriggerEventZoneBattlefield, eventSpan),
			},
		}
	}
	return nil
}
