package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
)

func emitTriggerEventClauses(abilities []Ability, cardName string) {
	for i := range abilities {
		trigger := abilities[i].Trigger
		if trigger == nil || trigger.PhaseStep != nil || trigger.PlayerEvent != nil {
			continue
		}
		trigger.TriggerEvent = parseTriggerEventClause(
			trigger.Event.Tokens,
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
		parseSacrificeTriggerEventClause,
		parseMutateTriggerEventClause,
		parseBecameTargetTriggerEventClause,
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
	subject     TriggerEventSubject
	controller  TriggerController
	player      TriggerPlayerSelector
	excludeSelf bool
	faceDown    bool
	oneOrMore   bool
	ok          bool
}

type permanentSubjectResult struct {
	subject     TriggerEventSubject
	controller  TriggerController
	excludeSelf bool
	oneOrMore   bool
	ok          bool
}

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
		if !mergeTriggerController(&clause.Controller, parsed.change.controller) {
			return nil
		}
		if !mergeTriggerPlayerSelector(&clause.Player, parsed.change.player) {
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
		return clause
	}
	return nil
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
		if !tokenWordsEqual(verb.remaining, "the", "battlefield") &&
			!tokenWordsEqual(verb.remaining, "the", "battlefield", "without", "dying") {
			return zoneChangeResult{}
		}
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

func parseZoneChangeSubject(
	subjectTokens []shared.Token,
	plural bool,
	atoms Atoms,
	cardName string,
) zoneSubjectResult {
	_ = cardName
	result := zoneSubjectResult{controller: ControllerAny}
	remaining := subjectTokens
	if plural {
		if span, count, ok := parseSelfSubject(remaining, atoms); ok && count == len(remaining) {
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
		result.subject = TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: span}
		result.ok = true
		return result
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
		result.subject = attached
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
	if len(remaining) == 0 {
		return zoneSubjectResult{}
	}
	selection, ok := parseTriggerSelection(remaining)
	if !ok {
		return zoneSubjectResult{}
	}
	if !mergeTriggerController(&result.controller, selection.Controller) {
		return zoneSubjectResult{}
	}
	selection.Controller = ControllerAny
	result.subject = TriggerEventSubject{
		Kind:      TriggerEventSubjectSelection,
		Span:      shared.SpanOf(subjectTokens),
		Selection: selection,
	}
	result.ok = true
	return result
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
		if !mergeTriggerController(&controller, relation.controller) ||
			!mergeTriggerPlayerSelector(&player, playerSelectorFromKind(relation.player, shared.SpanOf(tokens[len(prefix):]))) {
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

func parseSpellCastTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	_ Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	var actor TriggerEventActor
	var remaining []shared.Token
	switch {
	case len(tokens) >= 2 && equalWord(tokens[0], "you") && equalWord(tokens[1], "cast"):
		actor = TriggerEventActor{Kind: TriggerEventActorYou, Span: shared.SpanOf(tokens[:2])}
		remaining = tokens[2:]
	case len(tokens) >= 3 && equalWord(tokens[0], "a") && equalWord(tokens[1], "player") && equalWord(tokens[2], "casts"):
		actor = TriggerEventActor{Kind: TriggerEventActorPlayer, Span: shared.SpanOf(tokens[:3])}
		remaining = tokens[3:]
	case len(tokens) >= 3 && equalWord(tokens[0], "an") && equalWord(tokens[1], "opponent") && equalWord(tokens[2], "casts"):
		actor = TriggerEventActor{Kind: TriggerEventActorOpponent, Span: shared.SpanOf(tokens[:3])}
		remaining = tokens[3:]
	default:
		return nil
	}
	selection, ok := parseTriggerEventSpellSelection(remaining)
	if !ok || selection.FromZone.Kind != TriggerEventZoneNone && actor.Kind != TriggerEventActorYou {
		return nil
	}
	return &TriggerEventClause{
		Kind:           TriggerEventKindSpellCast,
		Actor:          actor,
		SpellSelection: selection,
	}
}

func parseTriggerEventSpellSelection(tokens []shared.Token) (TriggerEventSpellSelection, bool) {
	selection := TriggerEventSpellSelection{Span: shared.SpanOf(tokens)}
	switch {
	case syntaxWordsEqual(tokens, "a", "spell"):
		return selection, true
	case syntaxWordsEqual(tokens, "a", "kicked", "spell"):
		selection.Kicker = true
		return selection, true
	case syntaxWordsEqual(tokens, "a", "historic", "spell"):
		selection.Historic = true
		return selection, true
	case syntaxWordsEqual(tokens, "a", "spell", "from", "your", "graveyard"):
		selection.FromZone = TriggerEventZone{
			Kind: TriggerEventZoneGraveyard,
			Span: shared.SpanOf(tokens[3:]),
		}
		return selection, true
	case len(tokens) == 8 &&
		equalWord(tokens[0], "a") &&
		equalWord(tokens[1], "spell") &&
		equalWord(tokens[2], "with") &&
		equalWord(tokens[3], "mana") &&
		equalWord(tokens[4], "value") &&
		tokens[5].Kind == shared.Integer &&
		equalWord(tokens[6], "or") &&
		equalWord(tokens[7], "greater"):
		value := 0
		for _, r := range tokens[5].Text {
			if r < '0' || r > '9' {
				return TriggerEventSpellSelection{}, false
			}
			value = value*10 + int(r-'0')
		}
		selection.MatchManaValue = true
		selection.ManaValueAtLeast = value
		return selection, true
	case len(tokens) == 5 &&
		equalWord(tokens[0], "a") &&
		equalWord(tokens[1], "noncreature") &&
		tokens[2].Kind == shared.Comma &&
		equalWord(tokens[3], "nonland") &&
		equalWord(tokens[4], "spell"):
		selection.ExcludedTypes = []TriggerCardType{TriggerCardTypeCreature, TriggerCardTypeLand}
		return selection, true
	case len(tokens) == 5 &&
		(equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) &&
		equalWord(tokens[2], "or") &&
		equalWord(tokens[4], "spell"):
		leftType, leftTypeOK := triggerCardType(tokens[1].Text)
		rightType, rightTypeOK := triggerCardType(tokens[3].Text)
		if leftTypeOK && rightTypeOK && leftType != TriggerCardTypeUnknown && rightType != TriggerCardTypeUnknown {
			if leftType != TriggerCardTypeInstant && leftType != TriggerCardTypeSorcery ||
				rightType != TriggerCardTypeInstant && rightType != TriggerCardTypeSorcery ||
				leftType == rightType {
				return TriggerEventSpellSelection{}, false
			}
			selection.TypesAny = []TriggerCardType{leftType, rightType}
			return selection, true
		}
		leftSub, leftSubOK := recognizeSubtypePhrase(strings.ToLower(tokens[1].Text))
		rightSub, rightSubOK := recognizeSubtypePhrase(strings.ToLower(tokens[3].Text))
		if leftSubOK && rightSubOK {
			if leftSub != "Spirit" && leftSub != "Arcane" ||
				rightSub != "Spirit" && rightSub != "Arcane" ||
				leftSub == rightSub {
				return TriggerEventSpellSelection{}, false
			}
			selection.SubtypesAny = []TriggerSubtype{leftSub, rightSub}
			return selection, true
		}
		return TriggerEventSpellSelection{}, false
	case len(tokens) == 2 && equalWord(tokens[0], "an"):
		cardType, ok := triggerCardType(tokens[1].Text)
		if !ok || cardType != TriggerCardTypeInstant {
			return TriggerEventSpellSelection{}, false
		}
		selection.Types = []TriggerCardType{cardType}
		return selection, true
	case len(tokens) == 3 && (equalWord(tokens[0], "a") || equalWord(tokens[0], "an")) && equalWord(tokens[2], "spell"):
		word := strings.ToLower(tokens[1].Text)
		if cardType, ok := triggerCardType(word); ok && cardType != TriggerCardTypeUnknown {
			selection.Types = []TriggerCardType{cardType}
			return selection, true
		}
		if color, ok := recognizeColorWord(word); ok {
			selection.ColorsAny = []TriggerColor{triggerColor(tokens[1].Text)}
			_ = color
			return selection, true
		}
		switch word {
		case "colorless":
			selection.Colorless = true
			return selection, true
		case "multicolored":
			selection.Multicolored = true
			return selection, true
		}
		if rest, ok := strings.CutPrefix(word, "non"); ok {
			cardType, cardTypeOK := triggerCardType(rest)
			if cardTypeOK && cardType != TriggerCardTypeUnknown {
				selection.ExcludedTypes = []TriggerCardType{cardType}
				return selection, true
			}
		}
	default:
		return TriggerEventSpellSelection{}, false
	}
	return TriggerEventSpellSelection{}, false
}

func parseAbilityActivatedTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	_ Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	var actor TriggerEventActor
	var remaining []shared.Token
	switch {
	case len(tokens) >= 2 && equalWord(tokens[0], "you") && equalWord(tokens[1], "activate"):
		actor = TriggerEventActor{Kind: TriggerEventActorYou, Span: shared.SpanOf(tokens[:2])}
		remaining = tokens[2:]
	case len(tokens) >= 3 && equalWord(tokens[0], "an") && equalWord(tokens[1], "opponent") && equalWord(tokens[2], "activates"):
		actor = TriggerEventActor{Kind: TriggerEventActorOpponent, Span: shared.SpanOf(tokens[:3])}
		remaining = tokens[3:]
	case len(tokens) >= 3 && equalWord(tokens[0], "a") && equalWord(tokens[1], "player") && equalWord(tokens[2], "activates"):
		actor = TriggerEventActor{Kind: TriggerEventActorPlayer, Span: shared.SpanOf(tokens[:3])}
		remaining = tokens[3:]
	default:
		return nil
	}
	remaining, ok := stripTokenSuffix(remaining, "that", "isn't", "a", "mana", "ability")
	if !ok {
		return nil
	}
	clause := &TriggerEventClause{
		Kind:               TriggerEventKindAbilityActivated,
		Actor:              actor,
		ExcludeManaAbility: true,
	}
	if syntaxWordsEqual(remaining, "an", "ability") {
		return clause
	}
	remaining, ok = cutSyntaxWords(remaining, "an", "ability", "of")
	if !ok {
		return nil
	}
	selection, ok := parseSingleSelectionPhrase(remaining)
	if !ok {
		return nil
	}
	clause.SourceSelection = selection
	return clause
}

func parseAttackBlockTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if clause := parsePlayerAttackTriggerEventClause(tokens); clause != nil {
		return clause
	}
	if index := syntaxWordsIndex(tokens, "becomes", "blocked", "by"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		selection, ok := parseRelatedSelectionPhrase(tokens[index+3:])
		if !ok || !basicCreatureTriggerSelection(selection) {
			return nil
		}
		return &TriggerEventClause{
			Kind:             TriggerEventKindBecameBlocked,
			Subject:          subject.subject,
			Controller:       subject.controller,
			ExcludeSelf:      subject.excludeSelf,
			RelatedSelection: selection,
		}
	}
	if prefix, ok := stripTokenSuffix(tokens, "becomes", "blocked"); ok {
		subject := parsePermanentEventSubject(prefix, false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		return &TriggerEventClause{
			Kind:        TriggerEventKindBecameBlocked,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
		}
	}
	if index := syntaxWordsIndex(tokens, "blocks"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindBlock,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
		}
		if index+1 == len(tokens) {
			return clause
		}
		related, ok := parseRelatedSelectionPhrase(tokens[index+1:])
		if !ok || !selectionHasType(related, TriggerCardTypeCreature) {
			return nil
		}
		clause.RelatedSelection = related
		return clause
	}
	if index := syntaxWordsIndex(tokens, "block"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], true, atoms)
		if !subject.ok {
			return nil
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindBlock,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
		}
		if index+1 == len(tokens) {
			return clause
		}
		related, ok := parseRelatedSelectionPhrase(tokens[index+1:])
		if !ok || !selectionHasType(related, TriggerCardTypeCreature) {
			return nil
		}
		clause.RelatedSelection = related
		return clause
	}
	if index := syntaxWordsIndex(tokens, "attacks"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], false, atoms)
		if !subject.ok || subject.oneOrMore {
			return nil
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindAttack,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
		}
		if index+1 == len(tokens) {
			return clause
		}
		recipient, player, ok := parseAttackRecipient(tokens[index+1:])
		if !ok {
			return nil
		}
		clause.AttackRecipient = recipient
		clause.Player = player
		return clause
	}
	if index := syntaxWordsIndex(tokens, "attack"); index > 0 {
		subject := parsePermanentEventSubject(tokens[:index], true, atoms)
		if !subject.ok {
			return nil
		}
		clause := &TriggerEventClause{
			Kind:        TriggerEventKindAttack,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
		}
		if index+1 == len(tokens) {
			return clause
		}
		recipient, player, ok := parseAttackRecipient(tokens[index+1:])
		if !ok {
			return nil
		}
		clause.AttackRecipient = recipient
		clause.Player = player
		return clause
	}
	return nil
}

func parsePlayerAttackTriggerEventClause(tokens []shared.Token) *TriggerEventClause {
	var actor TriggerEventActor
	var rest []shared.Token
	switch {
	case len(tokens) >= 2 && syntaxWordsEqual(tokens[:2], "you", "attack"):
		actor = TriggerEventActor{Kind: TriggerEventActorYou, Span: tokens[0].Span}
		rest = tokens[2:]
	case len(tokens) >= 3 && syntaxWordsEqual(tokens[:3], "an", "opponent", "attacks"):
		actor = TriggerEventActor{Kind: TriggerEventActorOpponent, Span: shared.SpanOf(tokens[:2])}
		rest = tokens[3:]
	case len(tokens) >= 3 && syntaxWordsEqual(tokens[:3], "a", "player", "attacks"):
		actor = TriggerEventActor{Kind: TriggerEventActorPlayer, Span: shared.SpanOf(tokens[:2])}
		rest = tokens[3:]
	default:
		return nil
	}
	clause := &TriggerEventClause{
		Kind:      TriggerEventKindAttack,
		Actor:     actor,
		OneOrMore: true,
	}
	if len(rest) == 0 {
		return clause
	}
	if actor.Kind == TriggerEventActorYou && tokenWordsEqual(rest, "with", "one", "or", "more", "creatures") {
		return clause
	}
	recipient, player, ok := parseAttackRecipient(rest)
	if !ok || recipient.Kind != TriggerEventAttackRecipientPlayer {
		return nil
	}
	clause.OneOrMorePerAttackTarget = true
	clause.Player = player
	clause.AttackRecipient = recipient
	return clause
}

func parseDamageTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if syntaxWordsEqual(tokens, "you're", "dealt", "damage") ||
		syntaxWordsEqual(tokens, "you", "are", "dealt", "damage") ||
		syntaxWordsEqual(tokens, "you're", "dealt", "combat", "damage") ||
		syntaxWordsEqual(tokens, "you", "are", "dealt", "combat", "damage") ||
		syntaxWordsEqual(tokens, "you're", "dealt", "noncombat", "damage") ||
		syntaxWordsEqual(tokens, "you", "are", "dealt", "noncombat", "damage") {
		qualifier := TriggerEventCombatQualifier{Kind: TriggerEventCombatQualifierAny}
		switch {
		case slices.Contains(normalizedWords(tokens), "combat"):
			qualifier = TriggerEventCombatQualifier{Kind: TriggerEventCombatQualifierCombat, Span: tokens[len(tokens)-2].Span}
		case slices.Contains(normalizedWords(tokens), "noncombat"):
			qualifier = TriggerEventCombatQualifier{Kind: TriggerEventCombatQualifierNoncombat, Span: tokens[len(tokens)-2].Span}
		default:
		}
		player := playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
		return &TriggerEventClause{
			Kind:            TriggerEventKindDamageDealt,
			Player:          player,
			CombatQualifier: qualifier,
			DamageRecipient: TriggerEventDamageRecipient{
				Kind:   TriggerEventDamageRecipientPlayer,
				Span:   shared.SpanOf(tokens),
				Player: player,
			},
		}
	}
	for _, template := range []struct {
		words     []string
		qualifier TriggerEventCombatQualifierKind
		plural    bool
	}{
		{words: []string{"is", "dealt", "combat", "damage"}, qualifier: TriggerEventCombatQualifierCombat},
		{words: []string{"is", "dealt", "noncombat", "damage"}, qualifier: TriggerEventCombatQualifierNoncombat},
		{words: []string{"is", "dealt", "damage"}},
		{words: []string{"are", "dealt", "combat", "damage"}, qualifier: TriggerEventCombatQualifierCombat, plural: true},
		{words: []string{"are", "dealt", "noncombat", "damage"}, qualifier: TriggerEventCombatQualifierNoncombat, plural: true},
		{words: []string{"are", "dealt", "damage"}, plural: true},
	} {
		prefix, ok := stripTokenSuffix(tokens, template.words...)
		if !ok {
			continue
		}
		subject := parsePermanentEventSubject(prefix, template.plural, atoms)
		if !subject.ok {
			return nil
		}
		return &TriggerEventClause{
			Kind:        TriggerEventKindDamageDealt,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
			CombatQualifier: TriggerEventCombatQualifier{
				Kind: template.qualifier,
				Span: shared.SpanOf(tokens[len(prefix):]),
			},
			DamageRecipient: TriggerEventDamageRecipient{
				Kind: TriggerEventDamageRecipientPermanent,
				Span: shared.SpanOf(tokens[len(prefix):]),
			},
		}
	}
	for _, template := range []struct {
		words     []string
		qualifier TriggerEventCombatQualifierKind
		plural    bool
	}{
		{words: []string{"deals", "combat", "damage"}, qualifier: TriggerEventCombatQualifierCombat},
		{words: []string{"deals", "noncombat", "damage"}, qualifier: TriggerEventCombatQualifierNoncombat},
		{words: []string{"deals", "damage"}},
		{words: []string{"deal", "combat", "damage"}, qualifier: TriggerEventCombatQualifierCombat, plural: true},
		{words: []string{"deal", "noncombat", "damage"}, qualifier: TriggerEventCombatQualifierNoncombat, plural: true},
		{words: []string{"deal", "damage"}, plural: true},
	} {
		index := syntaxWordsIndex(tokens, template.words...)
		if index <= 0 {
			continue
		}
		sourceTokens := tokens[:index]
		recipientTokens := tokens[index+len(template.words):]
		clause := parseDamageSourcePattern(sourceTokens, template.plural, atoms)
		if clause == nil {
			return nil
		}
		clause.Kind = TriggerEventKindDamageDealt
		clause.CombatQualifier = TriggerEventCombatQualifier{
			Kind: template.qualifier,
			Span: shared.SpanOf(tokens[index : index+len(template.words)]),
		}
		if len(recipientTokens) == 0 {
			return clause
		}
		recipientTokens, ok := cutSyntaxWords(recipientTokens, "to")
		if !ok {
			return nil
		}
		recipient, player, ok := parseDamageRecipient(recipientTokens, atoms)
		if !ok {
			return nil
		}
		clause.DamageRecipient = recipient
		clause.Player = player
		return clause
	}
	return nil
}

func parseDamageSourcePattern(tokens []shared.Token, plural bool, atoms Atoms) *TriggerEventClause {
	working := tokens
	oneOrMore := false
	if rest, ok := cutSyntaxWords(working, "one", "or", "more"); ok {
		working = rest
		plural = true
		oneOrMore = true
	}
	if syntaxWordsEqual(working, "a", "source") {
		return &TriggerEventClause{
			DamageSource: TriggerEventSubject{
				Kind: TriggerEventSubjectDamageSource,
				Span: shared.SpanOf(tokens),
			},
			OneOrMore: oneOrMore,
		}
	}
	if syntaxWordsEqual(working, "a", "spell") {
		return &TriggerEventClause{
			DamageSourceIsStackObject: true,
			StackObject: TriggerEventStackObject{
				Kind: TriggerEventStackObjectSpell,
				Span: shared.SpanOf(working),
			},
			OneOrMore: oneOrMore,
		}
	}
	if selection, controller, ok := parseDamageSpellSource(working, plural); ok {
		return &TriggerEventClause{
			DamageSourceIsStackObject:  true,
			DamageSourceSpellSelection: selection,
			StackObject: TriggerEventStackObject{
				Kind: TriggerEventStackObjectSpell,
				Span: shared.SpanOf(working),
			},
			Controller: controller,
			OneOrMore:  oneOrMore,
		}
	}
	subject := parsePermanentEventSubject(tokens, plural, atoms)
	if !subject.ok {
		return nil
	}
	return &TriggerEventClause{
		DamageSource: subject.subject,
		Controller:   subject.controller,
		ExcludeSelf:  subject.excludeSelf,
		OneOrMore:    subject.oneOrMore || oneOrMore,
	}
}

func parseDamageSpellSource(
	tokens []shared.Token,
	plural bool,
) (TriggerEventSpellSelection, TriggerController, bool) {
	working, controller, ok := stripControllerSuffix(tokens)
	if !ok {
		return TriggerEventSpellSelection{}, ControllerAny, false
	}
	if plural && len(working) > 0 && equalWord(working[0], "other") {
		working = working[1:]
	}
	articleLength := 0
	if !plural {
		if len(working) == 0 ||
			(!equalWord(working[0], "a") && !equalWord(working[0], "an")) {
			return TriggerEventSpellSelection{}, ControllerAny, false
		}
		articleLength = 1
	}
	noun := "spell"
	if plural {
		noun = "spells"
	}
	if len(working) <= articleLength || !equalWord(working[len(working)-1], noun) {
		return TriggerEventSpellSelection{}, ControllerAny, false
	}
	phrase := working[articleLength : len(working)-1]
	selection := TriggerEventSpellSelection{Span: shared.SpanOf(working)}
	switch {
	case len(phrase) == 0:
		return selection, controller, true
	case syntaxWordsEqual(phrase, "noncreature"):
		selection.ExcludedTypes = []TriggerCardType{TriggerCardTypeCreature}
		return selection, controller, true
	case len(phrase) == 3 && equalWord(phrase[1], "or"):
		left, leftOK := triggerCardType(phrase[0].Text)
		right, rightOK := triggerCardType(phrase[2].Text)
		if !leftOK || !rightOK ||
			left != TriggerCardTypeInstant && left != TriggerCardTypeSorcery ||
			right != TriggerCardTypeInstant && right != TriggerCardTypeSorcery ||
			left == right {
			return TriggerEventSpellSelection{}, ControllerAny, false
		}
		selection.TypesAny = []TriggerCardType{left, right}
		return selection, controller, true
	case len(phrase) == 1:
		cardType, typeOK := triggerCardType(phrase[0].Text)
		if !typeOK || cardType == TriggerCardTypeUnknown {
			return TriggerEventSpellSelection{}, ControllerAny, false
		}
		selection.Types = []TriggerCardType{cardType}
		return selection, controller, true
	default:
		return TriggerEventSpellSelection{}, ControllerAny, false
	}
}

func parseDamageRecipient(
	tokens []shared.Token,
	atoms Atoms,
) (TriggerEventDamageRecipient, TriggerPlayerSelector, bool) {
	if span, count, ok := parseSelfSubject(tokens, atoms); ok && count == len(tokens) {
		return TriggerEventDamageRecipient{
			Kind:     TriggerEventDamageRecipientPermanent,
			Span:     span,
			IsSource: true,
		}, TriggerPlayerSelector{}, true
	}
	switch {
	case syntaxWordsEqual(tokens, "you"):
		player := playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
		return TriggerEventDamageRecipient{
			Kind:   TriggerEventDamageRecipientPlayer,
			Span:   tokens[0].Span,
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "an", "opponent"), syntaxWordsEqual(tokens, "one", "of", "your", "opponents"):
		player := playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens))
		return TriggerEventDamageRecipient{
			Kind:   TriggerEventDamageRecipientPlayer,
			Span:   shared.SpanOf(tokens),
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "a", "player"):
		player := playerSelectorFromKind(TriggerPlayerSelectorAny, shared.SpanOf(tokens))
		return TriggerEventDamageRecipient{
			Kind:   TriggerEventDamageRecipientPlayer,
			Span:   shared.SpanOf(tokens),
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "a", "player", "or", "planeswalker"):
		return TriggerEventDamageRecipient{
			Kind: TriggerEventDamageRecipientPlayer | TriggerEventDamageRecipientPermanent,
			Span: shared.SpanOf(tokens),
			Selection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker},
			},
		}, TriggerPlayerSelector{}, true
	case syntaxWordsEqual(tokens, "a", "player", "or", "battle"):
		return TriggerEventDamageRecipient{
			Kind: TriggerEventDamageRecipientPlayer | TriggerEventDamageRecipientPermanent,
			Span: shared.SpanOf(tokens),
			Selection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeBattle},
			},
		}, TriggerPlayerSelector{}, true
	case syntaxWordsEqual(tokens, "any", "target"):
		return TriggerEventDamageRecipient{
			Kind: TriggerEventDamageRecipientPlayer | TriggerEventDamageRecipientPermanent,
			Span: shared.SpanOf(tokens),
			Selection: TriggerSelection{
				RequiredTypesAny: []TriggerCardType{
					TriggerCardTypeCreature,
					TriggerCardTypePlaneswalker,
					TriggerCardTypeBattle,
				},
			},
		}, TriggerPlayerSelector{}, true
	}
	selection, ok := parseRelatedSelectionPhrase(tokens)
	if !ok {
		return TriggerEventDamageRecipient{}, TriggerPlayerSelector{}, false
	}
	player := TriggerPlayerSelector{}
	switch selection.Controller {
	case ControllerYou:
		player = playerSelectorFromKind(TriggerPlayerSelectorYou, shared.SpanOf(tokens))
	case ControllerOpponent:
		player = playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens))
	default:
	}
	return TriggerEventDamageRecipient{
		Kind:      TriggerEventDamageRecipientPermanent,
		Span:      shared.SpanOf(tokens),
		Player:    player,
		Selection: selection,
	}, player, true
}

func parseCounterTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	if index := syntaxWordsIndex(tokens, "counter", "is", "put", "on"); index > 1 && equalWord(tokens[0], "a") {
		if !syntaxWordsEqual(tokens[index+4:], "this", "creature") && !syntaxWordsEqual(tokens[index+4:], "this", "permanent") {
			return nil
		}
		counterKind, counterSpan, ok := triggerEventCounterIn(tokens, atoms)
		if !ok {
			return nil
		}
		subject := TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: shared.SpanOf(tokens[index+4:])}
		return &TriggerEventClause{
			Kind:    TriggerEventKindCounterAdded,
			Subject: subject,
			Counter: TriggerEventCounter{Kind: counterKind, Span: counterSpan},
		}
	}
	if index := syntaxWordsIndex(tokens, "counters", "are", "put", "on"); index > 3 && syntaxWordsEqual(tokens[:3], "one", "or", "more") {
		if !syntaxWordsEqual(tokens[index+4:], "this", "creature") && !syntaxWordsEqual(tokens[index+4:], "this", "permanent") {
			return nil
		}
		counterKind, counterSpan, ok := triggerEventCounterIn(tokens, atoms)
		if !ok {
			return nil
		}
		subject := TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: shared.SpanOf(tokens[index+4:])}
		return &TriggerEventClause{
			Kind:      TriggerEventKindCounterAdded,
			Subject:   subject,
			Counter:   TriggerEventCounter{Kind: counterKind, Span: counterSpan},
			OneOrMore: true,
		}
	}
	return nil
}

func triggerEventCounterIn(tokens []shared.Token, atoms Atoms) (TriggerEventCounterKind, shared.Span, bool) {
	kind, span, ok := atoms.CounterIn(shared.SpanOf(tokens))
	if !ok {
		return TriggerEventCounterAny, shared.Span{}, false
	}
	switch kind {
	case counter.PlusOnePlusOne:
		return TriggerEventCounterPlusOnePlusOne, span, true
	case counter.MinusOneMinusOne:
		return TriggerEventCounterMinusOneMinusOne, span, true
	default:
		return TriggerEventCounterAny, shared.Span{}, false
	}
}

func parsePermanentStateTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	for _, template := range []struct {
		suffix    []string
		kind      TriggerEventKind
		allowWhen bool
	}{
		{suffix: []string{"becomes", "tapped"}, kind: TriggerEventKindBecomesTapped, allowWhen: true},
		{suffix: []string{"becomes", "untapped"}, kind: TriggerEventKindBecomesUntapped, allowWhen: true},
		{suffix: []string{"is", "turned", "face", "up"}, kind: TriggerEventKindTurnedFaceUp, allowWhen: true},
	} {
		if intro != TriggerIntroductionWhenever && (intro != TriggerIntroductionWhen || !template.allowWhen) {
			continue
		}
		prefix, ok := stripTokenSuffix(tokens, template.suffix...)
		if !ok {
			continue
		}
		if span, count, ok := parseSelfSubject(prefix, atoms); ok && count == len(prefix) {
			return &TriggerEventClause{
				Kind:    template.kind,
				Subject: TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: span},
			}
		}
		if intro != TriggerIntroductionWhenever && template.kind != TriggerEventKindTurnedFaceUp {
			return nil
		}
		subject := parsePermanentEventSubject(prefix, false, atoms)
		if !subject.ok || subject.oneOrMore || subject.subject.Kind == TriggerEventSubjectSelf {
			return nil
		}
		return &TriggerEventClause{
			Kind:        template.kind,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
		}
	}
	return nil
}

func parseSacrificeTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	for _, actor := range []struct {
		words []string
		kind  TriggerEventActorKind
	}{
		{words: []string{"you", "sacrifice"}, kind: TriggerEventActorYou},
		{words: []string{"an", "opponent", "sacrifices"}, kind: TriggerEventActorOpponent},
		{words: []string{"a", "player", "sacrifices"}, kind: TriggerEventActorPlayer},
	} {
		remaining, ok := cutSyntaxWords(tokens, actor.words...)
		if !ok {
			continue
		}
		subject := parsePermanentEventSubject(remaining, false, atoms)
		if !subject.ok || subject.subject.Kind == TriggerEventSubjectAttached {
			return nil
		}
		return &TriggerEventClause{
			Kind:        TriggerEventKindSacrificed,
			Actor:       TriggerEventActor{Kind: actor.kind, Span: shared.SpanOf(tokens[:len(actor.words)])},
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
		}
	}
	return nil
}

func parseMutateTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	_ Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever || !syntaxWordsEqual(tokens, "this", "creature", "mutates") {
		return nil
	}
	return &TriggerEventClause{
		Kind: TriggerEventKindMutated,
		Subject: TriggerEventSubject{
			Kind: TriggerEventSubjectSelf,
			Span: shared.SpanOf(tokens[:2]),
		},
	}
}

func parseBecameTargetTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	index := syntaxWordsIndex(tokens, "becomes", "the", "target", "of")
	if index <= 0 {
		return nil
	}
	subject := parsePermanentEventSubject(tokens[:index], false, atoms)
	if !subject.ok || subject.oneOrMore {
		return nil
	}
	cause := tokens[index+4:]
	causeController := TriggerEventActorUnknown
	switch {
	case endsWithSyntaxWords(cause, "you", "control"):
		cause = cause[:len(cause)-2]
		causeController = TriggerEventActorYou
	case endsWithSyntaxWords(cause, "an", "opponent", "controls"):
		cause = cause[:len(cause)-3]
		causeController = TriggerEventActorOpponent
	default:
	}
	var stackObject TriggerEventStackObject
	switch {
	case syntaxWordsEqual(cause, "a", "spell"):
		stackObject = TriggerEventStackObject{Kind: TriggerEventStackObjectSpell, Span: shared.SpanOf(cause)}
	case syntaxWordsEqual(cause, "a", "spell", "or", "ability"):
		stackObject = TriggerEventStackObject{Kind: TriggerEventStackObjectAny, Span: shared.SpanOf(cause)}
	default:
		return nil
	}
	return &TriggerEventClause{
		Kind:            TriggerEventKindBecameTarget,
		Subject:         subject.subject,
		Controller:      subject.controller,
		ExcludeSelf:     subject.excludeSelf,
		StackObject:     stackObject,
		CauseController: causeController,
	}
}

func parsePermanentEventSubject(tokens []shared.Token, plural bool, atoms Atoms) permanentSubjectResult {
	result := permanentSubjectResult{controller: ControllerAny}
	remaining := tokens
	if rest, ok := cutSyntaxWords(remaining, "one", "or", "more"); ok {
		remaining = rest
		result.oneOrMore = true
		plural = true
	}
	if span, count, ok := parseSelfSubject(remaining, atoms); ok && count == len(remaining) {
		result.subject = TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: span}
		result.ok = true
		return result
	}
	if attached, ok := parseAttachedEventSubject(remaining); ok {
		result.subject = attached
		result.ok = true
		return result
	}
	var relationsOK bool
	remaining, result.controller, relationsOK = stripControllerSuffix(remaining)
	if !relationsOK {
		return permanentSubjectResult{}
	}
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
			return permanentSubjectResult{}
		}
	}
	if len(remaining) == 0 {
		return permanentSubjectResult{}
	}
	selection, ok := parseTriggerSelection(remaining)
	if !ok {
		return permanentSubjectResult{}
	}
	if !mergeTriggerController(&result.controller, selection.Controller) {
		return permanentSubjectResult{}
	}
	selection.Controller = ControllerAny
	result.subject = TriggerEventSubject{
		Kind:      TriggerEventSubjectSelection,
		Span:      shared.SpanOf(tokens),
		Selection: selection,
	}
	result.ok = true
	return result
}

func stripControllerSuffix(tokens []shared.Token) ([]shared.Token, TriggerController, bool) {
	for _, relation := range []struct {
		words      []string
		controller TriggerController
	}{
		{words: []string{"your", "opponents", "control"}, controller: ControllerOpponent},
		{words: []string{"an", "opponent", "controls"}, controller: ControllerOpponent},
		{words: []string{"you", "don't", "control"}, controller: ControllerOpponent},
		{words: []string{"you", "control"}, controller: ControllerYou},
	} {
		prefix, ok := stripTokenSuffix(tokens, relation.words...)
		if !ok {
			continue
		}
		return prefix, relation.controller, len(prefix) > 0
	}
	return tokens, ControllerAny, len(tokens) > 0
}

func parseAttackRecipient(tokens []shared.Token) (TriggerEventAttackRecipient, TriggerPlayerSelector, bool) {
	switch {
	case syntaxWordsEqual(tokens, "you"):
		player := playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
		return TriggerEventAttackRecipient{
			Kind:   TriggerEventAttackRecipientPlayer,
			Span:   tokens[0].Span,
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "an", "opponent"), syntaxWordsEqual(tokens, "one", "of", "your", "opponents"):
		player := playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens))
		return TriggerEventAttackRecipient{
			Kind:   TriggerEventAttackRecipientPlayer,
			Span:   shared.SpanOf(tokens),
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "a", "player"):
		player := playerSelectorFromKind(TriggerPlayerSelectorAny, shared.SpanOf(tokens))
		return TriggerEventAttackRecipient{
			Kind:   TriggerEventAttackRecipientPlayer,
			Span:   shared.SpanOf(tokens),
			Player: player,
		}, player, true
	case syntaxWordsEqual(tokens, "a", "player", "or", "planeswalker"):
		return TriggerEventAttackRecipient{
			Kind: TriggerEventAttackRecipientPlayer | TriggerEventAttackRecipientPlaneswalker,
			Span: shared.SpanOf(tokens),
			Selection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker},
			},
		}, TriggerPlayerSelector{}, true
	case syntaxWordsEqual(tokens, "a", "player", "or", "battle"):
		return TriggerEventAttackRecipient{
			Kind: TriggerEventAttackRecipientPlayer | TriggerEventAttackRecipientBattle,
			Span: shared.SpanOf(tokens),
			Selection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeBattle},
			},
		}, TriggerPlayerSelector{}, true
	case syntaxWordsEqual(tokens, "you", "or", "a", "planeswalker", "you", "control"):
		player := playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
		return TriggerEventAttackRecipient{
			Kind:   TriggerEventAttackRecipientPlayer | TriggerEventAttackRecipientPlaneswalker,
			Span:   shared.SpanOf(tokens),
			Player: player,
			Selection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker},
				Controller:    ControllerYou,
			},
		}, player, true
	case syntaxWordsEqual(tokens, "you", "or", "a", "battle", "you", "protect"):
		player := playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
		return TriggerEventAttackRecipient{
			Kind:   TriggerEventAttackRecipientPlayer | TriggerEventAttackRecipientBattle,
			Span:   shared.SpanOf(tokens),
			Player: player,
			Selection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeBattle},
			},
		}, player, true
	}
	selection, ok := parseRelatedSelectionPhrase(tokens)
	if !ok || len(selection.RequiredTypes) != 1 {
		return TriggerEventAttackRecipient{}, TriggerPlayerSelector{}, false
	}
	recipient := TriggerEventAttackRecipient{
		Span:      shared.SpanOf(tokens),
		Selection: selection,
	}
	switch selection.RequiredTypes[0] {
	case TriggerCardTypePlaneswalker:
		recipient.Kind = TriggerEventAttackRecipientPlaneswalker
	case TriggerCardTypeBattle:
		recipient.Kind = TriggerEventAttackRecipientBattle
	default:
		return TriggerEventAttackRecipient{}, TriggerPlayerSelector{}, false
	}
	player := TriggerPlayerSelector{}
	switch selection.Controller {
	case ControllerYou:
		player = playerSelectorFromKind(TriggerPlayerSelectorYou, shared.SpanOf(tokens))
	case ControllerOpponent:
		player = playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens))
	default:
	}
	recipient.Player = player
	return recipient, player, true
}

func parseSingleSelectionPhrase(tokens []shared.Token) (TriggerSelection, bool) {
	if len(tokens) == 0 {
		return TriggerSelection{}, false
	}
	remaining := tokens
	switch {
	case equalWord(remaining[0], "another"):
		return TriggerSelection{}, false
	case equalWord(remaining[0], "a"), equalWord(remaining[0], "an"):
		remaining = remaining[1:]
	default:
	}
	if len(remaining) == 0 {
		return TriggerSelection{}, false
	}
	return parseTriggerSelection(remaining)
}

func parseRelatedSelectionPhrase(tokens []shared.Token) (TriggerSelection, bool) {
	if len(tokens) > 0 && equalWord(tokens[0], "another") {
		tokens = tokens[1:]
	}
	return parseSingleSelectionPhrase(tokens)
}

func parseAttachedEventSubject(tokens []shared.Token) (TriggerEventSubject, bool) {
	if len(tokens) < 2 {
		return TriggerEventSubject{}, false
	}
	subject := TriggerEventSubject{
		Kind: TriggerEventSubjectAttached,
		Span: shared.SpanOf(tokens),
	}
	switch {
	case equalWord(tokens[0], "enchanted"):
		subject.AttachKind = TriggerEventAttachEnchanted
	case equalWord(tokens[0], "equipped"):
		subject.AttachKind = TriggerEventAttachEquipped
	case equalWord(tokens[0], "fortified"):
		subject.AttachKind = TriggerEventAttachFortified
	default:
		return TriggerEventSubject{}, false
	}
	selection, ok := parseTriggerSelection(tokens[1:])
	if !ok {
		return TriggerEventSubject{}, false
	}
	subject.Selection = selection
	return subject, true
}

func stripTokenSuffix(tokens []shared.Token, words ...string) ([]shared.Token, bool) {
	if len(tokens) < len(words) {
		return nil, false
	}
	tail := tokens[len(tokens)-len(words):]
	for i, word := range words {
		if word == "'" {
			if tail[i].Kind != shared.Apostrophe {
				return nil, false
			}
			continue
		}
		if !equalWord(tail[i], word) {
			return nil, false
		}
	}
	return tokens[:len(tokens)-len(words)], true
}

func tokenCountForSpan(tokens []shared.Token, span shared.Span) int {
	if len(tokens) == 0 || span == (shared.Span{}) || tokens[0].Span.Start.Offset != span.Start.Offset {
		return 0
	}
	for i := range tokens {
		if tokens[i].Span.End.Offset == span.End.Offset {
			return i + 1
		}
		if tokens[i].Span.End.Offset > span.End.Offset {
			return 0
		}
	}
	return 0
}

func parseSelfSubject(tokens []shared.Token, atoms Atoms) (shared.Span, int, bool) {
	if len(tokens) == 0 {
		return shared.Span{}, 0, false
	}
	if span, ok := atoms.SourceMarkerSpanStartingAt(tokens[0].Span); ok {
		if count := tokenCountForSpan(tokens, span); count > 0 {
			return span, count, true
		}
	}
	if span, ok := atoms.SourceNameSpanStartingAt(tokens[0].Span); ok {
		if count := tokenCountForSpan(tokens, span); count > 0 {
			return span, count, true
		}
	}
	return shared.Span{}, 0, false
}

func syntaxWordsIndex(tokens []shared.Token, words ...string) int {
	if len(words) == 0 || len(tokens) < len(words) {
		return -1
	}
	for start := 0; start+len(words) <= len(tokens); start++ {
		match := true
		for i, word := range words {
			if word == "'" {
				if tokens[start+i].Kind != shared.Apostrophe {
					match = false
					break
				}
				continue
			}
			if !equalWord(tokens[start+i], word) {
				match = false
				break
			}
		}
		if match {
			return start
		}
	}
	return -1
}

func endsWithSyntaxWords(tokens []shared.Token, words ...string) bool {
	if len(tokens) < len(words) {
		return false
	}
	return syntaxWordsEqual(tokens[len(tokens)-len(words):], words...)
}

func selectionHasType(selection TriggerSelection, kind TriggerCardType) bool {
	return slices.Contains(selection.RequiredTypes, kind) ||
		slices.Contains(selection.RequiredTypesAny, kind)
}

func basicCreatureTriggerSelection(selection TriggerSelection) bool {
	return len(selection.RequiredTypes) == 1 &&
		selection.RequiredTypes[0] == TriggerCardTypeCreature &&
		len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		len(selection.ExcludedColors) == 0 &&
		!selection.Colorless &&
		!selection.Multicolored &&
		!selection.NonToken &&
		!selection.TokenOnly &&
		selection.Controller == ControllerAny
}

func mergeTriggerController(current *TriggerController, additional TriggerController) bool {
	if additional == ControllerAny {
		return true
	}
	if *current != ControllerAny && *current != additional {
		return false
	}
	*current = additional
	return true
}

func mergeTriggerPlayerSelector(current *TriggerPlayerSelector, additional TriggerPlayerSelector) bool {
	if additional.Kind == TriggerPlayerSelectorUnknown {
		return true
	}
	if current.Kind != TriggerPlayerSelectorUnknown && current.Kind != additional.Kind {
		return false
	}
	if current.Kind == TriggerPlayerSelectorUnknown {
		*current = additional
	}
	return true
}

func playerSelectorFromKind(kind TriggerPlayerSelectorKind, span shared.Span) TriggerPlayerSelector {
	if kind == TriggerPlayerSelectorUnknown {
		return TriggerPlayerSelector{}
	}
	return TriggerPlayerSelector{Kind: kind, Span: span}
}

func triggerEventZone(kind TriggerEventZoneKind, span shared.Span) TriggerEventZone {
	return TriggerEventZone{Kind: kind, Span: span}
}

func zoneWordSpan(tokens []shared.Token, kind TriggerEventZoneKind) shared.Span {
	for i := len(tokens) - 1; i >= 0; i-- {
		switch kind {
		case TriggerEventZoneBattlefield:
			if equalWord(tokens[i], "battlefield") {
				return tokens[i].Span
			}
		case TriggerEventZoneGraveyard:
			if equalWord(tokens[i], "graveyard") || equalWord(tokens[i], "graveyards") {
				return tokens[i].Span
			}
		case TriggerEventZoneHand:
			if equalWord(tokens[i], "hand") || equalWord(tokens[i], "hands") {
				return tokens[i].Span
			}
		case TriggerEventZoneExile:
			if equalWord(tokens[i], "exile") || equalWord(tokens[i], "exiled") {
				return tokens[i].Span
			}
		case TriggerEventZoneLibrary:
			if equalWord(tokens[i], "library") || equalWord(tokens[i], "libraries") {
				return tokens[i].Span
			}
		case TriggerEventZoneStack:
			if equalWord(tokens[i], "stack") {
				return tokens[i].Span
			}
		case TriggerEventZoneCommand:
			if equalWord(tokens[i], "command") {
				return tokens[i].Span
			}
		default:
		}
	}
	return shared.Span{}
}
