package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseSelfGraveyardOrAnotherUnionTriggerEventClause recognizes the two-verb
// self-or-another battlefield-to-graveyard union "this creature dies or another
// <Selection> [you control] is put into a graveyard from the battlefield" (Scrap
// Trawler), where the source's own departure to the graveyard is spelled with a
// dedicated self verb ("dies" or "is put into a graveyard from the battlefield")
// instead of sharing the partner clause's verb. The single-verb shared-subject
// form "this creature or another <Selection> you control is put into a graveyard
// from the battlefield" already parses through parseZoneChangeTriggerEventClause;
// this clause rewrites the redundant self verb away into that shared-subject form
// and delegates, then confirms the resulting union is a battlefield-to-graveyard
// move so the discarded self verb names the same event as the partner clause.
func parseSelfGraveyardOrAnotherUnionTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	cardName string,
) *TriggerEventClause {
	_, count, ok := parseSelfSubject(tokens, atoms)
	if !ok || count >= len(tokens) {
		return nil
	}
	selfTokens := tokens[:count]
	afterSelf, ok := cutSelfGraveyardLeaveVerb(tokens[count:])
	if !ok || len(afterSelf) == 0 || !equalWord(afterSelf[0], "or") {
		return nil
	}
	rewritten := make([]shared.Token, 0, len(selfTokens)+len(afterSelf))
	rewritten = append(rewritten, selfTokens...)
	rewritten = append(rewritten, afterSelf...)
	clause := parseZoneChangeTriggerEventClause(rewritten, intro, atoms, cardName)
	if clause == nil || !clause.SelfOrAnother {
		return nil
	}
	if !clause.Zone.MatchFromZone || clause.Zone.FromZone.Kind != TriggerEventZoneBattlefield ||
		!clause.Zone.MatchToZone || clause.Zone.ToZone.Kind != TriggerEventZoneGraveyard {
		return nil
	}
	clause.Span = shared.SpanOf(tokens)
	return clause
}

// cutSelfGraveyardLeaveVerb consumes a self-side battlefield-to-graveyard verb
// phrase — "dies" or "is put into a graveyard from the battlefield" — and returns
// the remaining tokens. Both phrases describe the same battlefield-to-graveyard
// move, so they share the union partner's zone change.
func cutSelfGraveyardLeaveVerb(tokens []shared.Token) ([]shared.Token, bool) {
	if verb := cutEventVerb(tokens, "dies", "die"); verb.ok {
		return verb.remaining, true
	}
	if verb := cutEventVerb(tokens, "is", "are"); verb.ok {
		if rest, ok := cutTokenPrefix(verb.remaining, "put", "into", "a", "graveyard", "from", "the", "battlefield"); ok {
			return rest, true
		}
	}
	return nil, false
}
