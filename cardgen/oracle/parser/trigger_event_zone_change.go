package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseZoneChangeTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	cardName string,
) *TriggerEventClause {
	for split := 1; split < len(tokens); split++ {
		tailTokens := tokens[split:]
		parsed := parseZoneChange(tailTokens)
		if !parsed.ok {
			continue
		}
		subject := parseZoneChangeSubject(tokens[:split], parsed.plural, atoms, cardName)
		if !subject.ok {
			continue
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindZoneChange,
			Subject:     subject.subject,
			Controller:  subject.controller,
			Player:      subject.player,
			OneOrMore:   subject.oneOrMore,
			ExcludeSelf: subject.excludeSelf,
			FaceDown:    subject.faceDown,
			Zone:        parsed.change.zone,
			ZoneChange:  parsed.change.kind,
			Tapped:      parsed.change.tapped,
		}
		if subject.selfOrAnother {
			// The union re-admits the source, so the self-excluding "another"
			// restriction must not reject it at runtime.
			clause.SelfOrAnother = true
			clause.ExcludeSelf = false
		}
		if subject.dealtDamageBySrc {
			// The damaged-by-source relative clause is only meaningful for the
			// dies zone change; reject it on any other zone change so the
			// wording never silently loses its restriction.
			if parsed.change.kind.Kind != TriggerEventZoneChangeDied {
				return nil
			}
			clause.DealtDamageBySourceThisTurn = true
		}
		if !mergeTriggerController(&clause.Controller, parsed.change.controller) {
			return nil
		}
		if !mergeTriggerPlayerSelector(&clause.Player, &parsed.change.player) {
			return nil
		}
		if parsed.change.kind.Kind == TriggerEventZoneChangeDied {
			if !selectionHasType(clause.Subject.Selection, TriggerCardTypeCreature) {
				clause.Subject.Selection.RequiredTypes = append(
					clause.Subject.Selection.RequiredTypes,
					TriggerCardTypeCreature,
				)
			}
		}
		if subject.permanentNoun &&
			zoneChangeAdmitsNonPermanentCards(parsed.change.zone) &&
			len(clause.Subject.Selection.RequiredTypes) == 0 &&
			len(clause.Subject.Selection.RequiredTypesAny) == 0 {
			clause.Subject.Selection.RequiredTypesAny = append(
				clause.Subject.Selection.RequiredTypesAny,
				triggerPermanentCardTypes...,
			)
		}
		return clause
	}
	return nil
}

// triggerPermanentCardTypes is the expansion of a "permanent card" trigger
// subject: the card types a permanent card can have (CR 110.4a). It mirrors the
// permanentCardTypes list used by battlefield card-type conditions.
var triggerPermanentCardTypes = []TriggerCardType{
	TriggerCardTypeArtifact,
	TriggerCardTypeBattle,
	TriggerCardTypeCreature,
	TriggerCardTypeEnchantment,
	TriggerCardTypeLand,
	TriggerCardTypePlaneswalker,
}

// zoneChangeAdmitsNonPermanentCards reports whether the zone change can move a
// non-permanent card, which is when a "permanent card" subject actually narrows
// the selection. A source or destination that is the battlefield guarantees the
// object is a permanent (only permanents exist there), so the permanent-type
// restriction is redundant and is not applied, preserving the prior output for
// battlefield-origin and enters-the-battlefield triggers. Non-battlefield or
// unconstrained ("from anywhere") sources can hold instants and sorceries, so
// the restriction is applied.
func zoneChangeAdmitsNonPermanentCards(z TriggerEventZoneContext) bool {
	if z.MatchToZone && z.ToZone.Kind == TriggerEventZoneBattlefield {
		return false
	}
	if z.MatchFromZone && !z.ExcludeFromZone && z.FromZone.Kind == TriggerEventZoneBattlefield {
		return false
	}
	return true
}

type parsedZoneChange struct {
	kind       TriggerEventZoneChange
	zone       TriggerEventZoneContext
	controller TriggerController
	player     TriggerPlayerSelector
	tapped     TriggerEventTappedState
}

type zoneChangeResult struct {
	change parsedZoneChange
	plural bool
	ok     bool
}

type eventVerbResult struct {
	remaining []shared.Token
	plural    bool
	ok        bool
}

func parseZoneChange(tokens []shared.Token) zoneChangeResult {
	span := shared.SpanOf(tokens)
	if verb := cutEventVerb(tokens, "enters", "enter"); verb.ok {
		return parseEnteredBattlefieldZoneChange(tokens, span, verb)
	}

	if verb := cutEventVerb(tokens, "dies", "die"); verb.ok && len(verb.remaining) == 0 {
		return matchedZoneChange(&parsedZoneChange{
			kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeDied, Span: span},
			zone: TriggerEventZoneContext{
				Span:          span,
				MatchFromZone: true,
				FromZone:      triggerEventZone(TriggerEventZoneBattlefield, zoneWordSpan(tokens, TriggerEventZoneBattlefield)),
				MatchToZone:   true,
				ToZone:        triggerEventZone(TriggerEventZoneGraveyard, zoneWordSpan(tokens, TriggerEventZoneGraveyard)),
			},
		}, verb.plural)
	}

	if verb := cutEventVerb(tokens, "leaves", "leave"); verb.ok {
		if tokenWordsEqual(verb.remaining, "the", "battlefield") ||
			tokenWordsEqual(verb.remaining, "the", "battlefield", "without", "dying") {
			change := parsedZoneChange{
				kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeMoved, Span: span},
				zone: TriggerEventZoneContext{
					Span:          span,
					MatchFromZone: true,
					FromZone:      triggerEventZone(TriggerEventZoneBattlefield, zoneWordSpan(tokens, TriggerEventZoneBattlefield)),
				},
			}
			if tokenWordsEqual(verb.remaining, "the", "battlefield", "without", "dying") {
				change.zone.ExcludeToZone = true
				change.zone.ToZone = triggerEventZone(
					TriggerEventZoneGraveyard,
					zoneWordSpan(tokens, TriggerEventZoneGraveyard),
				)
			}
			return matchedZoneChange(&change, verb.plural)
		}
		// "leave[s] [your / a / an opponent's] graveyard" departs the graveyard
		// for any zone, so the trigger only constrains the origin graveyard and
		// its owner; the destination zone is unconstrained.
		if originZone, player, ok := parseOriginZone(verb.remaining); ok && originZone.Kind == TriggerEventZoneGraveyard {
			change := parsedZoneChange{
				kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeMoved, Span: span},
				zone: TriggerEventZoneContext{
					Span:          span,
					MatchFromZone: true,
					FromZone:      originZone,
				},
				player: player,
			}
			return matchedZoneChange(&change, verb.plural)
		}
		return zoneChangeResult{}
	}

	if verb := cutEventVerb(tokens, "is", "are"); verb.ok {
		if put, ok := cutTokenPrefix(verb.remaining, "put", "into"); ok {
			return parsePutIntoZoneChange(tokens, put, verb.plural)
		}
		if exiled, ok := cutTokenPrefix(verb.remaining, "exiled", "from", "the", "battlefield"); ok && len(exiled) == 0 {
			return matchedZoneChange(&parsedZoneChange{
				kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeMoved, Span: span},
				zone: TriggerEventZoneContext{
					Span:          span,
					MatchFromZone: true,
					FromZone:      triggerEventZone(TriggerEventZoneBattlefield, zoneWordSpan(tokens, TriggerEventZoneBattlefield)),
					MatchToZone:   true,
					ToZone:        triggerEventZone(TriggerEventZoneExile, zoneWordSpan(tokens, TriggerEventZoneExile)),
				},
			}, verb.plural)
		}
		if returned, ok := cutTokenPrefix(verb.remaining, "returned", "to"); ok {
			zone, player, ok := parseDestinationZone(returned)
			if !ok || zone.Kind != TriggerEventZoneHand {
				return zoneChangeResult{}
			}
			return matchedZoneChange(&parsedZoneChange{
				kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeMoved, Span: span},
				zone: TriggerEventZoneContext{
					Span:          span,
					MatchFromZone: true,
					FromZone:      triggerEventZone(TriggerEventZoneBattlefield, zoneWordSpan(tokens, TriggerEventZoneBattlefield)),
					MatchToZone:   true,
					ToZone:        zone,
				},
				player: player,
			}, verb.plural)
		}
	}
	return zoneChangeResult{}
}

func parseEnteredBattlefieldZoneChange(
	tokens []shared.Token,
	span shared.Span,
	verb eventVerbResult,
) zoneChangeResult {
	change := parsedZoneChange{
		kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeEnteredBattlefield, Span: span},
		zone: TriggerEventZoneContext{
			Span:        span,
			MatchToZone: true,
			ToZone:      triggerEventZone(TriggerEventZoneBattlefield, zoneWordSpan(tokens, TriggerEventZoneBattlefield)),
		},
	}
	remaining := verb.remaining
	if len(remaining) == 0 || tokenWordsEqual(remaining, "the", "battlefield") {
		return matchedZoneChange(&change, verb.plural)
	}
	if tokenWordsEqual(remaining, "tapped") || tokenWordsEqual(remaining, "untapped") {
		kind := TriggerEventTappedStateTapped
		if tokenWordsEqual(remaining, "untapped") {
			kind = TriggerEventTappedStateUntapped
		}
		change.tapped = TriggerEventTappedState{Kind: kind, Span: shared.SpanOf(remaining)}
		return matchedZoneChange(&change, verb.plural)
	}
	if battlefield, ok := cutTokenPrefix(remaining, "the", "battlefield"); ok {
		remaining = battlefield
	}
	if under, ok := cutTokenPrefix(remaining, "under"); ok {
		controller, ok := parseEnteringController(under)
		if !ok {
			return zoneChangeResult{}
		}
		change.controller = controller
		return matchedZoneChange(&change, verb.plural)
	}
	if from, ok := cutTokenPrefix(remaining, "from"); ok {
		zone, player, ok := parseOriginZone(from)
		if !ok {
			return zoneChangeResult{}
		}
		change.zone.MatchFromZone = true
		change.zone.FromZone = zone
		change.player = player
		return matchedZoneChange(&change, verb.plural)
	}
	return zoneChangeResult{}
}

func parsePutIntoZoneChange(tokens, destination []shared.Token, plural bool) zoneChangeResult {
	// "put into <zone> from anywhere other than the battlefield" fires for moves
	// into the destination from any origin except the battlefield (mill, discard,
	// bounce-to-graveyard), excluding the common death move.
	if prefix, ok := stripTokenSuffix(destination, "from", "anywhere", "other", "than", "the", "battlefield"); ok {
		return parsePutIntoExcludingBattlefieldZoneChange(tokens, prefix, plural)
	}
	// "put into <zone> from anywhere" leaves the origin unconstrained, so the
	// trigger fires for moves from any zone into the destination (deaths, mill,
	// discard, bounce-to-graveyard, and so on).
	if prefix, ok := stripTokenSuffix(destination, "from", "anywhere"); ok {
		return parsePutIntoFromAnywhereZoneChange(tokens, prefix, plural)
	}
	// "put into <zone> from <origin zone>" constrains both endpoints, e.g.
	// "into your graveyard from your library" (a mill that puts land cards into
	// your graveyard). The origin zone is parsed by parseOriginZone; only
	// recognized origins (graveyard, hand, exile, library) reach this branch,
	// so an unrecognized origin like "the battlefield" falls through to the
	// default battlefield handling below.
	if before, after, ok := splitTokensOnFirstWord(destination, "from"); ok && len(before) > 0 && len(after) > 0 {
		if originZone, _, ok := parseOriginZone(after); ok {
			destZone, player, ok := parseDestinationZone(before)
			if !ok || destZone.Kind == TriggerEventZoneBattlefield {
				return zoneChangeResult{}
			}
			span := shared.SpanOf(tokens)
			return matchedZoneChange(&parsedZoneChange{
				kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeMoved, Span: span},
				zone: TriggerEventZoneContext{
					Span:          span,
					MatchFromZone: true,
					FromZone:      originZone,
					MatchToZone:   true,
					ToZone:        destZone,
				},
				player: player,
			}, plural)
		}
	}
	if prefix, ok := stripTokenSuffix(destination, "from", "the", "battlefield"); ok {
		destination = prefix
	}
	zone, player, ok := parseDestinationZone(destination)
	if !ok || zone.Kind == TriggerEventZoneBattlefield {
		return zoneChangeResult{}
	}
	span := shared.SpanOf(tokens)
	return matchedZoneChange(&parsedZoneChange{
		kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeMoved, Span: span},
		zone: TriggerEventZoneContext{
			Span:          span,
			MatchFromZone: true,
			FromZone:      triggerEventZone(TriggerEventZoneBattlefield, zoneWordSpan(tokens, TriggerEventZoneBattlefield)),
			MatchToZone:   true,
			ToZone:        zone,
		},
		player: player,
	}, plural)
}

func parsePutIntoFromAnywhereZoneChange(tokens, destination []shared.Token, plural bool) zoneChangeResult {
	zone, player, ok := parseDestinationZone(destination)
	if !ok || zone.Kind == TriggerEventZoneBattlefield {
		return zoneChangeResult{}
	}
	span := shared.SpanOf(tokens)
	return matchedZoneChange(&parsedZoneChange{
		kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeMoved, Span: span},
		zone: TriggerEventZoneContext{
			Span:        span,
			MatchToZone: true,
			ToZone:      zone,
		},
		player: player,
	}, plural)
}

func parsePutIntoExcludingBattlefieldZoneChange(tokens, destination []shared.Token, plural bool) zoneChangeResult {
	zone, player, ok := parseDestinationZone(destination)
	if !ok || zone.Kind == TriggerEventZoneBattlefield {
		return zoneChangeResult{}
	}
	span := shared.SpanOf(tokens)
	return matchedZoneChange(&parsedZoneChange{
		kind: TriggerEventZoneChange{Kind: TriggerEventZoneChangeMoved, Span: span},
		zone: TriggerEventZoneContext{
			Span:            span,
			MatchToZone:     true,
			ToZone:          zone,
			ExcludeFromZone: true,
			FromZone:        triggerEventZone(TriggerEventZoneBattlefield, zoneWordSpan(tokens, TriggerEventZoneBattlefield)),
		},
		player: player,
	}, plural)
}

func matchedZoneChange(change *parsedZoneChange, plural bool) zoneChangeResult {
	return zoneChangeResult{change: *change, plural: plural, ok: true}
}

func cutEventVerb(
	tokens []shared.Token,
	singular string,
	plural string,
) eventVerbResult {
	if rest, ok := cutTokenPrefix(tokens, singular); ok {
		return eventVerbResult{remaining: rest, ok: true}
	}
	if rest, ok := cutTokenPrefix(tokens, plural); ok {
		return eventVerbResult{remaining: rest, plural: true, ok: true}
	}
	return eventVerbResult{}
}

func cutTokenPrefix(tokens []shared.Token, words ...string) ([]shared.Token, bool) {
	if len(tokens) < len(words) {
		return nil, false
	}
	for i, word := range words {
		if word == "'" {
			if tokens[i].Kind != shared.Apostrophe {
				return nil, false
			}
			continue
		}
		if !equalWord(tokens[i], word) {
			return nil, false
		}
	}
	return tokens[len(words):], true
}

func tokenWordsEqual(tokens []shared.Token, words ...string) bool {
	rest, ok := cutTokenPrefix(tokens, words...)
	return ok && len(rest) == 0
}

func parseEnteringController(tokens []shared.Token) (TriggerController, bool) {
	switch {
	case tokenWordsEqual(tokens, "your", "control"):
		return ControllerYou, true
	case tokenWordsEqual(tokens, "an", "opponent's", "control"),
		tokenWordsEqual(tokens, "your", "opponents", "'", "control"):
		return ControllerOpponent, true
	default:
		return ControllerAny, false
	}
}

func parseOriginZone(tokens []shared.Token) (TriggerEventZone, TriggerPlayerSelector, bool) {
	switch {
	case tokenWordsEqual(tokens, "a", "graveyard"):
		return sourceSpannedZone(TriggerEventZoneGraveyard, tokens), TriggerPlayerSelector{}, true
	case tokenWordsEqual(tokens, "your", "graveyard"):
		return sourceSpannedZone(TriggerEventZoneGraveyard, tokens),
			playerSelectorFromKind(TriggerPlayerSelectorYou, shared.SpanOf(tokens)), true
	case tokenWordsEqual(tokens, "an", "opponent's", "graveyard"):
		return sourceSpannedZone(TriggerEventZoneGraveyard, tokens),
			playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens)), true
	case tokenWordsEqual(tokens, "your", "hand"):
		return sourceSpannedZone(TriggerEventZoneHand, tokens),
			playerSelectorFromKind(TriggerPlayerSelectorYou, shared.SpanOf(tokens)), true
	case tokenWordsEqual(tokens, "exile"):
		return sourceSpannedZone(TriggerEventZoneExile, tokens), TriggerPlayerSelector{}, true
	case tokenWordsEqual(tokens, "your", "library"):
		return sourceSpannedZone(TriggerEventZoneLibrary, tokens),
			playerSelectorFromKind(TriggerPlayerSelectorYou, shared.SpanOf(tokens)), true
	case tokenWordsEqual(tokens, "a", "library"):
		return sourceSpannedZone(TriggerEventZoneLibrary, tokens), TriggerPlayerSelector{}, true
	case tokenWordsEqual(tokens, "an", "opponent's", "library"):
		return sourceSpannedZone(TriggerEventZoneLibrary, tokens),
			playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens)), true
	default:
		return TriggerEventZone{}, TriggerPlayerSelector{}, false
	}
}

func parseDestinationZone(tokens []shared.Token) (TriggerEventZone, TriggerPlayerSelector, bool) {
	switch {
	case tokenWordsEqual(tokens, "a", "graveyard"),
		tokenWordsEqual(tokens, "its", "owner's", "graveyard"),
		tokenWordsEqual(tokens, "their", "owners", "'", "graveyards"):
		return sourceSpannedZone(TriggerEventZoneGraveyard, tokens), TriggerPlayerSelector{}, true
	case tokenWordsEqual(tokens, "your", "graveyard"):
		return sourceSpannedZone(TriggerEventZoneGraveyard, tokens),
			playerSelectorFromKind(TriggerPlayerSelectorYou, shared.SpanOf(tokens)), true
	case tokenWordsEqual(tokens, "an", "opponent's", "graveyard"):
		return sourceSpannedZone(TriggerEventZoneGraveyard, tokens),
			playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens)), true
	case tokenWordsEqual(tokens, "exile"):
		return sourceSpannedZone(TriggerEventZoneExile, tokens), TriggerPlayerSelector{}, true
	case tokenWordsEqual(tokens, "hand"),
		tokenWordsEqual(tokens, "a", "player's", "hand"),
		tokenWordsEqual(tokens, "its", "owner's", "hand"),
		tokenWordsEqual(tokens, "their", "owners", "'", "hands"):
		return sourceSpannedZone(TriggerEventZoneHand, tokens), TriggerPlayerSelector{}, true
	case tokenWordsEqual(tokens, "your", "hand"):
		return sourceSpannedZone(TriggerEventZoneHand, tokens),
			playerSelectorFromKind(TriggerPlayerSelectorYou, shared.SpanOf(tokens)), true
	default:
		return TriggerEventZone{}, TriggerPlayerSelector{}, false
	}
}

func sourceSpannedZone(kind TriggerEventZoneKind, tokens []shared.Token) TriggerEventZone {
	return triggerEventZone(kind, shared.SpanOf(tokens))
}

func splitTokensOnFirstWord(tokens []shared.Token, word string) (before, after []shared.Token, found bool) {
	for i := range tokens {
		if equalWord(tokens[i], word) {
			return tokens[:i], tokens[i+1:], true
		}
	}
	return nil, nil, false
}

func parseZoneChangeSubject(
	subjectTokens []shared.Token,
	plural bool,
	atoms Atoms,
	cardName string,
) zoneSubjectResult {
	_ = cardName
	result := zoneSubjectResult{controller: ControllerAny}
	remaining := subjectTokens
	var subtypeFromEntryChoice bool
	if rest, ok := stripTokenSuffix(remaining, "of", "the", "chosen", "type"); ok {
		remaining = rest
		subtypeFromEntryChoice = true
	}
	if plural {
		if span, count, ok := parseSelfSubject(remaining, atoms); ok && count == len(remaining) {
			if subtypeFromEntryChoice {
				return zoneSubjectResult{}
			}
			result.subject = TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: span}
			result.ok = true
			return result
		}
		if rest, ok := cutSyntaxWords(remaining, "one", "or", "more"); ok {
			remaining = rest
			result.oneOrMore = true
		}
	}
	if span, count, ok := parseSelfSubject(remaining, atoms); ok && count == len(remaining) {
		if subtypeFromEntryChoice {
			return zoneSubjectResult{}
		}
		result.subject = TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: span}
		result.ok = true
		return result
	}
	// "this creature or another <Selection> you control" unions the source
	// itself with a self-excluding selection subject. The remaining tokens after
	// the leading "<self> or" must parse as the "another <Selection>" form.
	if _, count, ok := parseSelfSubject(remaining, atoms); ok && count < len(remaining) && equalWord(remaining[count], "or") {
		remaining = remaining[count+1:]
		result.selfOrAnother = true
	}
	for i := 0; i+2 < len(remaining); i++ {
		if !equalWord(remaining[i], "other") || !equalWord(remaining[i+1], "than") {
			continue
		}
		span, count, ok := parseSelfSubject(remaining[i+2:], atoms)
		if !ok || count != len(remaining[i+2:]) {
			continue
		}
		_ = span
		remaining = remaining[:i]
		result.excludeSelf = true
		break
	}
	if attached, ok := parseAttachedEventSubject(remaining); ok {
		if subtypeFromEntryChoice {
			return zoneSubjectResult{}
		}
		result.subject = attached
		result.ok = true
		return result
	}
	if commander, controller, ok := parseCommanderEventSubject(remaining); ok {
		if subtypeFromEntryChoice || result.selfOrAnother {
			return zoneSubjectResult{}
		}
		result.controller = controller
		result.subject = commander
		result.ok = true
		return result
	}
	relations := stripZoneSubjectRelations(remaining)
	if !relations.ok {
		return zoneSubjectResult{}
	}
	remaining = relations.remaining
	result.controller = relations.controller
	result.player = relations.player
	if plural {
		if len(remaining) > 0 && equalWord(remaining[0], "other") {
			remaining = remaining[1:]
			result.excludeSelf = true
		}
	} else {
		switch {
		case len(remaining) > 0 && equalWord(remaining[0], "another"):
			remaining = remaining[1:]
			result.excludeSelf = true
		case len(remaining) > 0 && equalWord(remaining[0], "a"):
			remaining = remaining[1:]
		case len(remaining) > 0 && equalWord(remaining[0], "an"):
			remaining = remaining[1:]
		default:
			return zoneSubjectResult{}
		}
	}
	if len(remaining) > 0 && equalWord(remaining[0], "face-down") {
		remaining = remaining[1:]
		result.faceDown = true
	}
	if rest, ok := stripDealtDamageBySourceThisTurn(remaining, atoms); ok {
		remaining = rest
		result.dealtDamageBySrc = true
	}
	if len(remaining) == 0 {
		return zoneSubjectResult{}
	}
	// A "permanent [card]" subject imposes a permanent-card-type restriction that
	// only bites in non-battlefield source zones (a library or graveyard can hold
	// instant and sorcery cards). The caller expands it to the permanent-type
	// union when the zone change admits non-permanent cards; here we only record
	// that the noun was "permanent" so the base selection stays an any-card
	// selection, identical to the prior output for battlefield-origin triggers.
	result.permanentNoun = tokenWordsEqual(remaining, "permanent") ||
		tokenWordsEqual(remaining, "permanents") ||
		tokenWordsEqual(remaining, "permanent", "card") ||
		tokenWordsEqual(remaining, "permanent", "cards")
	if tokenWordsEqual(remaining, "card") || tokenWordsEqual(remaining, "cards") {
		// A bare "card"/"cards" subject imposes no type restriction, so it
		// resolves to an any-card selection that the runtime matches without a
		// type filter.
		result.subject = TriggerEventSubject{
			Kind:      TriggerEventSubjectSelection,
			Span:      shared.SpanOf(subjectTokens),
			Selection: TriggerSelection{SubtypeFromEntryChoice: subtypeFromEntryChoice},
		}
		if result.selfOrAnother && !result.excludeSelf {
			return zoneSubjectResult{}
		}
		result.ok = true
		return result
	}
	selection, ok := parseTriggerSelection(remaining)
	if !ok {
		return zoneSubjectResult{}
	}
	if !mergeTriggerController(&result.controller, selection.Controller) {
		return zoneSubjectResult{}
	}
	selection.Controller = ControllerAny
	selection.SubtypeFromEntryChoice = subtypeFromEntryChoice
	result.subject = TriggerEventSubject{
		Kind:      TriggerEventSubjectSelection,
		Span:      shared.SpanOf(subjectTokens),
		Selection: selection,
	}
	if result.selfOrAnother && !result.excludeSelf {
		// "<self> or" must be followed by a self-excluding "another <Selection>".
		return zoneSubjectResult{}
	}
	result.ok = true
	return result
}

// stripDealtDamageBySourceThisTurn removes a trailing "dealt damage by <source>
// this turn" relative clause from a dies-subject's tokens, returning the
// remaining base-noun tokens and whether the clause was present. The relative
// clause names the ability's own source through "this creature" or the card's
// own name (parseSelfSubject), modeling "a creature dealt damage by this
// creature this turn dies" (CR 603.2). It does not strip the clause when the
// named source is anything other than the ability source, so equipped- or
// enchanted-creature variants fail closed for now.
func stripDealtDamageBySourceThisTurn(tokens []shared.Token, atoms Atoms) ([]shared.Token, bool) {
	if !endsWithSyntaxWords(tokens, "this", "turn") {
		return tokens, false
	}
	body := tokens[:len(tokens)-2]
	idx := syntaxWordsIndex(body, "dealt", "damage", "by")
	if idx <= 0 {
		return tokens, false
	}
	sourceTokens := body[idx+3:]
	if _, count, ok := parseSelfSubject(sourceTokens, atoms); !ok || count != len(sourceTokens) {
		return tokens, false
	}
	return body[:idx], true
}

type zoneSubjectRelations struct {
	remaining  []shared.Token
	controller TriggerController
	player     TriggerPlayerSelector
	ok         bool
}

func stripZoneSubjectRelations(tokens []shared.Token) zoneSubjectRelations {
	controller := ControllerAny
	player := TriggerPlayerSelector{}
	for _, relation := range []struct {
		words      []string
		controller TriggerController
		player     TriggerPlayerSelectorKind
	}{
		{words: []string{"you", "control", "but", "don't", "own"}, controller: ControllerYou, player: TriggerPlayerSelectorOpponent},
		{words: []string{"your", "opponents", "control"}, controller: ControllerOpponent},
		{words: []string{"an", "opponent", "controls"}, controller: ControllerOpponent},
		{words: []string{"you", "don't", "control"}, controller: ControllerOpponent},
		{words: []string{"you", "control"}, controller: ControllerYou},
		{words: []string{"owned", "by", "another", "player"}, player: TriggerPlayerSelectorOpponent},
		{words: []string{"an", "opponent", "owns"}, player: TriggerPlayerSelectorOpponent},
		{words: []string{"you", "own"}, player: TriggerPlayerSelectorYou},
	} {
		prefix, ok := stripTokenSuffix(tokens, relation.words...)
		if !ok {
			continue
		}
		selector := playerSelectorFromKind(relation.player, shared.SpanOf(tokens[len(prefix):]))
		if !mergeTriggerController(&controller, relation.controller) ||
			!mergeTriggerPlayerSelector(&player, &selector) {
			return zoneSubjectRelations{}
		}
		return zoneSubjectRelations{
			remaining:  prefix,
			controller: controller,
			player:     player,
			ok:         len(prefix) > 0,
		}
	}
	return zoneSubjectRelations{
		remaining:  tokens,
		controller: controller,
		player:     player,
		ok:         len(tokens) > 0,
	}
}
