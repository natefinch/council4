package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

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

func mergeTriggerPlayerSelector(current, additional *TriggerPlayerSelector) bool {
	if additional.Kind == TriggerPlayerSelectorUnknown {
		return true
	}
	if current.Kind != TriggerPlayerSelectorUnknown && current.Kind != additional.Kind {
		return false
	}
	if current.Kind == TriggerPlayerSelectorUnknown {
		*current = *additional
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
