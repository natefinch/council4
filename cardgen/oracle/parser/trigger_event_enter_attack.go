package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseEnterAttackUnionTriggerEventClause recognizes the event-union triggers
// "<subject> enters or attacks", "<subject> enters or dies", and their mirrors
// ("<subject> attacks or enters", "<subject> dies or enters"), e.g. "Whenever
// this creature enters or attacks" or "When this creature enters or dies". Both
// verbs share a single subject, so the trigger fires when either the
// enters-the-battlefield event or the partner event happens (CR 603.2). The
// enter event is the primary clause and the partner event joins it through
// UnionKind. The dies partner needs no zone filter because the runtime emits a
// dedicated dies event only for battlefield-to-graveyard moves.
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
		partner, ok := enterUnionPartner(first, second)
		if !ok {
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
			UnionKind:   partner,
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

// enterGraveyardLeaveWords is the non-creature battlefield-to-graveyard verb
// phrase "is put into a graveyard from the battlefield", the union partner
// printed on permanents that are not creatures (Ichor Wellspring and the other
// artifact "wellspring"/"schematic" recursion payoffs). The runtime emits the
// dedicated dies event for any battlefield-to-graveyard move, creature or not, so
// the partner joins the enter event through the same UnionKind as "enters or
// dies".
var enterGraveyardLeaveWords = []string{"is", "put", "into", "a", "graveyard", "from", "the", "battlefield"}

// parseEnterGraveyardUnionTriggerEventClause recognizes the shared-subject
// event-union trigger "<subject> enters or is put into a graveyard from the
// battlefield", e.g. "When this artifact enters or is put into a graveyard from
// the battlefield". The enter event is the primary clause and the
// battlefield-to-graveyard move joins it through UnionKind (CR 603.2). This
// recognizer is deliberately disjoint from parseEnterAttackUnionTriggerEventClause:
// that one matches only the single-word partners "attacks" and "dies", while this
// one matches only the multi-word graveyard phrase, so the shared
// parseTriggerEventClause matchCount never sees two recognizers claim the same
// tokens.
func parseEnterGraveyardUnionTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	for i := 1; i+1 < len(tokens); i++ {
		if !equalWord(tokens[i], "enters") || !equalWord(tokens[i+1], "or") {
			continue
		}
		if !tokenWordsEqual(tokens[i+2:], enterGraveyardLeaveWords...) {
			return nil
		}
		subject := parsePermanentEventSubject(tokens[:i], false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		eventSpan := shared.SpanOf(tokens[i:])
		return &TriggerEventClause{
			Kind:        TriggerEventKindZoneChange,
			UnionKind:   TriggerEventKindDied,
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

// enterUnionPartner reports the union secondary for an "enters or <partner>"
// (or mirrored) pair of verbs sharing one subject. One verb must be "enters";
// the other must be a supported partner ("attacks" or "dies").
func enterUnionPartner(first, second shared.Token) (TriggerEventKind, bool) {
	switch {
	case equalWord(first, "enters") && equalWord(second, "attacks"),
		equalWord(first, "attacks") && equalWord(second, "enters"):
		return TriggerEventKindAttack, true
	case equalWord(first, "enters") && equalWord(second, "dies"),
		equalWord(first, "dies") && equalWord(second, "enters"):
		return TriggerEventKindDied, true
	default:
		return TriggerEventKindUnknown, false
	}
}
