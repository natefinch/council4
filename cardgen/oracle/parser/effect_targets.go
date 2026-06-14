package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func parseTargets(tokens []shared.Token, atoms Atoms) []TargetSyntax {
	var targets []TargetSyntax
	for i, token := range tokens {
		if !equalWord(token, "target") {
			continue
		}
		start := i
		cardinality := TargetCardinalitySyntax{Min: 1, Max: 1}
		switch {
		case i >= 3 && effectWordsAt(tokens, i-3, "any", "number", "of"):
			start = i - 3
			cardinality = TargetCardinalitySyntax{Min: 0, Max: 99}
		case i >= 4 && effectWordsAt(tokens, i-4, "up", "to") &&
			(effectWordsAt(tokens, i-1, "another") || effectWordsAt(tokens, i-1, "other")):
			start = i - 4
			cardinality.Min = 0
			var ok bool
			cardinality.Max, ok = effectNumber(tokens[i-2], atoms)
			if !ok || cardinality.Max < 1 {
				cardinality = TargetCardinalitySyntax{}
			}
		case i >= 3 && effectWordsAt(tokens, i-3, "up", "to"):
			start = i - 3
			cardinality.Min = 0
			var ok bool
			cardinality.Max, ok = effectNumber(tokens[i-1], atoms)
			if !ok || cardinality.Max < 1 {
				cardinality = TargetCardinalitySyntax{}
			}
		case i >= 1:
			if count, ok := effectNumber(tokens[i-1], atoms); ok && count > 0 {
				start = i - 1
				cardinality = TargetCardinalitySyntax{Min: count, Max: count}
			} else if equalWord(tokens[i-1], "any") ||
				equalWord(tokens[i-1], "another") ||
				equalWord(tokens[i-1], "other") {
				start = i - 1
			}
		default:
		}
		end := targetSyntaxEnd(tokens, i+1)
		selectionTokens := append([]shared.Token(nil), tokens[start:i]...)
		selectionTokens = append(selectionTokens, tokens[i+1:end]...)
		selection := parseSelection(selectionTokens, atoms)
		if targetSelectionHasUnsupportedQualifier(selectionTokens, atoms) {
			selection = SelectionSyntax{Span: selection.Span, Text: selection.Text}
		}
		targets = append(targets, TargetSyntax{
			Span:        shared.SpanOf(tokens[start:end]),
			Text:        joinedEffectText(tokens[start:end]),
			Cardinality: cardinality,
			Selection:   selection,
			Exact:       exactRuntimeTargetSyntax(tokens[start:end], cardinality, selection),
		})
	}
	return targets
}

func exactRuntimeTargetSyntax(tokens []shared.Token, cardinality TargetCardinalitySyntax, selection SelectionSyntax) bool {
	if cardinality != (TargetCardinalitySyntax{Min: 1, Max: 1}) {
		return false
	}
	text := joinedEffectText(tokens)
	switch selection.Kind {
	case SelectionAny:
		return text == "any target"
	case SelectionPlayer:
		return strings.EqualFold(text, "target player")
	case SelectionOpponent:
		return strings.EqualFold(text, "target opponent")
	case SelectionActivatedAbility:
		return strings.EqualFold(text, "target activated ability")
	case SelectionTriggeredAbility:
		return strings.EqualFold(text, "target triggered ability")
	case SelectionActivatedOrTriggeredAbility:
		return strings.EqualFold(text, "target activated or triggered ability")
	case SelectionSpellActivatedOrTriggeredAbility:
		return strings.EqualFold(text, "target spell, activated ability, or triggered ability")
	case SelectionSpell:
		switch strings.ToLower(text) {
		case "target spell", "target instant spell", "target sorcery spell", "target creature spell",
			"target artifact spell", "target noncreature spell":
			return true
		}
		return false
	case SelectionCreature:
		if strings.EqualFold(text, "target creature spell") {
			return true
		}
	case SelectionArtifact:
		if strings.EqualFold(text, "target artifact spell") {
			return true
		}
	default:
	}

	noun := ""
	switch selection.Kind {
	case SelectionArtifact:
		noun = "artifact"
	case SelectionBattle:
		noun = "battle"
	case SelectionCreature:
		noun = "creature"
	case SelectionEnchantment:
		noun = "enchantment"
	case SelectionLand:
		noun = "land"
	case SelectionPermanent:
		noun = "permanent"
	case SelectionPlaneswalker:
		noun = "planeswalker"
	default:
		return false
	}
	if selection.Another || selection.Other ||
		(selection.Tapped && selection.Untapped) ||
		((selection.Tapped || selection.Untapped) && (selection.Attacking || selection.Blocking)) {
		return false
	}
	expected := "target "
	switch {
	case selection.Attacking && selection.Blocking:
		expected += "attacking or blocking "
	case selection.Attacking:
		expected += "attacking "
	case selection.Blocking:
		expected += "blocking "
	case selection.Tapped:
		expected += "tapped "
	case selection.Untapped:
		expected += "untapped "
	default:
	}
	expected += noun
	switch selection.Controller {
	case SelectionControllerAny:
	case SelectionControllerYou:
		expected += " you control"
	case SelectionControllerOpponent:
		expected += " an opponent controls"
	case SelectionControllerNotYou:
		expected += " you don't control"
	default:
		return false
	}
	return strings.EqualFold(text, expected)
}

func targetSelectionHasUnsupportedQualifier(tokens []shared.Token, atoms Atoms) bool {
	for _, token := range tokens {
		if token.Kind == shared.Integer || token.Kind == shared.Comma ||
			selectionGrammarWord(token) || selectionAtomCoversToken(atoms, token) {
			continue
		}
		return true
	}
	return false
}

func selectionGrammarWord(token shared.Token) bool {
	for _, word := range []string{
		"a", "an", "all", "any", "number", "of", "up", "to", "or", "and",
		"with", "without", "from", "in", "your", "you", "control", "controls", "don't",
		"opponent", "opponent's", "opponents", "activated", "triggered",
		"mana", "value", "power", "toughness", "equal", "less", "greater",
		"battlefield", "graveyard", "hand", "library", "exile", "command",
	} {
		if equalWord(token, word) {
			return true
		}
	}
	return false
}

func selectionAtomCoversToken(atoms Atoms, token shared.Token) bool {
	covered := func(span shared.Span) bool {
		return spanCovers(span, token.Span)
	}
	for _, atom := range atoms.Colors() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ExcludedColors() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ColorQualifiers() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.CardTypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ExcludedTypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Supertypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Subtypes() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.ObjectNouns() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Zones() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Cardinals() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.SelectionFlags() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.Controllers() {
		if covered(atom.Span) {
			return true
		}
	}
	for _, atom := range atoms.KeywordSelectors() {
		if covered(atom.Span) {
			return true
		}
	}
	return false
}

func targetSyntaxEnd(tokens []shared.Token, start int) int {
	if start+8 <= len(tokens) &&
		effectWordsAt(tokens, start, "spell") &&
		tokens[start+1].Kind == shared.Comma &&
		effectWordsAt(tokens, start+2, "activated", "ability") &&
		tokens[start+4].Kind == shared.Comma &&
		effectWordsAt(tokens, start+5, "or", "triggered", "ability") {
		return start + 8
	}
	end := start
	for end < len(tokens) {
		token := tokens[end]
		if token.Kind == shared.Comma || token.Kind == shared.Period || token.Kind == shared.Semicolon ||
			targetDestinationStartsAt(tokens, end) ||
			equalWord(token, "unless") ||
			(equalWord(token, "and") && end+2 < len(tokens) && equalWord(tokens[end+1], "you") && effectWordKind(tokens[end+2]) != EffectUnknown) ||
			(equalWord(token, "and") && end+1 < len(tokens) && effectWordKind(tokens[end+1]) != EffectUnknown) ||
			(end > start && effectWordKind(token) != EffectUnknown) ||
			(equalWord(token, "until") && end+1 < len(tokens)) ||
			(equalWord(token, "for") && effectWordsAt(tokens, end, "for", "as", "long", "as")) ||
			(equalWord(token, "as") && effectWordsAt(tokens, end, "as", "long", "as", "this")) {
			break
		}

		end++
	}

	return end
}

func targetDestinationStartsAt(tokens []shared.Token, index int) bool {
	for _, phrase := range [][]string{
		{"to", "its", "owner's", "hand"},
		{"to", "your", "hand"},
		{"to", "their", "hand"},
		{"to", "the", "battlefield"},
		{"onto", "the", "battlefield"},
		{"into", "your", "graveyard"},
		{"into", "your", "library"},
		{"on", "top", "of", "your", "library"},
		{"on", "the", "top", "of", "your", "library"},
		{"on", "bottom", "of", "your", "library"},
		{"on", "the", "bottom", "of", "your", "library"},
	} {
		if effectWordsAt(tokens, index, phrase...) {
			return true
		}
	}
	return false
}

func ambiguousZoneChoice(tokens []shared.Token, atoms Atoms, span shared.Span) bool {
	zones := atoms.Zones()
	for i, first := range zones {
		if !spanCovers(span, first.Span) {
			continue
		}
		for _, second := range zones[i+1:] {
			if first.Zone == second.Zone || !spanCovers(span, second.Span) {
				continue
			}
			for _, token := range tokens {
				if token.Span.Start.Offset >= first.Span.End.Offset &&
					token.Span.End.Offset <= second.Span.Start.Offset &&
					equalWord(token, "or") {
					return true
				}
			}
		}
	}
	return false
}

func parseSelection(tokens []shared.Token, atoms Atoms) SelectionSyntax {
	selection := SelectionSyntax{Span: shared.SpanOf(tokens), Text: joinedEffectText(tokens)}
	words := normalizedWords(tokens)
	switch {
	case slices.Equal(words, []string{"activated", "ability"}):
		selection.Kind = SelectionActivatedAbility
	case slices.Equal(words, []string{"triggered", "ability"}):
		selection.Kind = SelectionTriggeredAbility
	case slices.Equal(words, []string{"activated", "or", "triggered", "ability"}):
		selection.Kind = SelectionActivatedOrTriggeredAbility
	case slices.Equal(words, []string{"spell", "activated", "ability", "or", "triggered", "ability"}):
		selection.Kind = SelectionSpellActivatedOrTriggeredAbility
	default:
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && selection.Kind == SelectionUnknown {
			selection.Kind = selectionKindForNoun(noun)
		}
		if cardType, ok := atoms.CardTypeAt(token.Span); ok && !slices.Contains(selection.RequiredTypesAny, cardType) {
			selection.RequiredTypesAny = append(selection.RequiredTypesAny, cardType)
		}
		if cardType, ok := atoms.ExcludedCardTypeAt(token.Span); ok && !slices.Contains(selection.ExcludedTypes, cardType) {
			selection.ExcludedTypes = append(selection.ExcludedTypes, cardType)
		}
		if colorValue, ok := atoms.ColorAt(token.Span); ok && !slices.Contains(selection.ColorsAny, colorValue) {
			selection.ColorsAny = append(selection.ColorsAny, colorValue)
		}
		if colorValue, ok := atoms.ExcludedColorAt(token.Span); ok && !slices.Contains(selection.ExcludedColors, colorValue) {
			selection.ExcludedColors = append(selection.ExcludedColors, colorValue)
		}
		if supertype, ok := atoms.SupertypeAt(token.Span); ok && !slices.Contains(selection.Supertypes, supertype) {
			selection.Supertypes = append(selection.Supertypes, supertype)
		}
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && noun == ObjectNounSpell &&
			selection.Kind != SelectionActivatedAbility &&
			selection.Kind != SelectionTriggeredAbility &&
			selection.Kind != SelectionActivatedOrTriggeredAbility &&
			selection.Kind != SelectionSpellActivatedOrTriggeredAbility {
			selection.Kind = SelectionSpell
			break
		}
	}
	for _, token := range tokens {
		if noun, ok := atoms.ObjectNounAt(token.Span); ok && noun == ObjectNounAbility &&
			selection.Kind != SelectionActivatedAbility &&
			selection.Kind != SelectionTriggeredAbility &&
			selection.Kind != SelectionActivatedOrTriggeredAbility &&
			selection.Kind != SelectionSpellActivatedOrTriggeredAbility {
			selection.Kind = SelectionUnknown
			break
		}
	}
	span := shared.SpanOf(tokens)
	selection.SubtypesAny = atoms.SubtypesIn(span)
	if relation, ok := atoms.ControllerIn(span); ok {
		switch relation {
		case ControllerRelationYouControl:
			selection.Controller = SelectionControllerYou
		case ControllerRelationOpponentControls:
			selection.Controller = SelectionControllerOpponent
		case ControllerRelationYouDontControl:
			selection.Controller = SelectionControllerNotYou
		default:
		}
	}
	selection.Zone = firstZone(atoms, span, ZoneRoleFrom)
	if selection.Zone == zone.None {
		selection.Zone = firstZone(atoms, span, ZoneRolePlain)
	}
	switch {
	case effectContainsWords(words, "your", "graveyard"):
		selection.Controller = SelectionControllerYou
	case effectContainsWords(words, "opponent's", "graveyard"):
		selection.Controller = SelectionControllerOpponent
	default:
	}
	selection.All = slices.Contains(words, "all")
	selection.Another = atoms.SelectionFlagIn(span, SelectionFlagAnother)
	selection.Other = atoms.SelectionFlagIn(span, SelectionFlagOther)
	selection.Attacking = atoms.SelectionFlagIn(span, SelectionFlagAttacking)
	selection.Blocking = atoms.SelectionFlagIn(span, SelectionFlagBlocking)
	selection.Tapped = atoms.SelectionFlagIn(span, SelectionFlagTapped)
	selection.Untapped = atoms.SelectionFlagIn(span, SelectionFlagUntapped)
	if slices.Contains(words, "any") && selection.Kind == SelectionUnknown {
		selection.Kind = SelectionAny
	}
	if keyword, ok := atoms.KeywordSelectorIn(span, false); ok {
		selection.Keyword = keyword.Keyword
	}
	if !parseSelectionNumbers(tokens, atoms, &selection) {
		return SelectionSyntax{Span: span, Text: joinedEffectText(tokens)}
	}
	return selection
}

func parseSelectionNumbers(tokens []shared.Token, atoms Atoms, selection *SelectionSyntax) bool {
	for i := range tokens {
		if i+2 < len(tokens) && effectWordsAt(tokens, i, "mana", "value") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+2:], atoms)
			if !ok {
				return false
			}
			selection.ManaValue = comparison
			selection.MatchManaValue = true
			continue
		}
		if equalWord(tokens[i], "power") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+1:], atoms)
			if !ok {
				return false
			}
			selection.Power = comparison
			selection.MatchPower = true
			continue
		}
		if equalWord(tokens[i], "toughness") {
			comparison, ok := parseSelectionNumberComparison(tokens[i+1:], atoms)
			if !ok {
				return false
			}
			selection.Toughness = comparison
			selection.MatchToughness = true
		}
	}
	return true
}

func parseSelectionNumberComparison(tokens []shared.Token, atoms Atoms) (compare.Int, bool) {
	if len(tokens) == 0 {
		return compare.Int{}, false
	}
	if value, ok := effectNumber(tokens[0], atoms); ok {
		if len(tokens) >= 3 && equalWord(tokens[1], "or") {
			switch {
			case equalWord(tokens[2], "less"):
				return compare.Int{Op: compare.LessOrEqual, Value: value}, true
			case equalWord(tokens[2], "greater"):
				return compare.Int{Op: compare.GreaterOrEqual, Value: value}, true
			}
		}
		return compare.Int{Op: compare.Equal, Value: value}, true
	}
	if len(tokens) >= 3 && effectWordsAt(tokens, 0, "equal", "to") {
		if value, ok := effectNumber(tokens[2], atoms); ok {
			return compare.Int{Op: compare.Equal, Value: value}, true
		}
	}
	return compare.Int{}, false
}

func parseEffectStaticSubject(tokens []shared.Token, atoms Atoms) EffectStaticSubjectSyntax {
	subtype := func(index int) (types.Sub, bool) {
		if index >= len(tokens) {
			return "", false
		}
		value, ok := atoms.SubtypeAt(tokens[index].Span)
		return value, ok && SubtypeMatchesAnyRuntimeCardType(value, []types.Card{types.Creature, types.Kindred})
	}
	switch {
	case len(tokens) >= 3 &&
		(equalWord(tokens[0], "enchanted") || equalWord(tokens[0], "equipped")) &&
		equalWord(tokens[1], "creature") &&
		(equalWord(tokens[2], "gets") || equalWord(tokens[2], "has")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: shared.SpanOf(tokens[:2])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "other", "creatures", "you", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "creatures", "you", "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatures, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "creatures", "your", "opponents", "control") &&
		(equalWord(tokens[4], "get") || equalWord(tokens[4], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOpponentControlledCreatures, Span: shared.SpanOf(tokens[:4])}
	case len(tokens) >= 5 && effectWordsAt(tokens, 0, "each", "wall", "you", "control") &&
		(equalWord(tokens[4], "gets") || equalWord(tokens[4], "has")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledWalls, Span: shared.SpanOf(tokens[:4]), Subtype: types.Wall, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "walls", "you", "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledWalls, Span: shared.SpanOf(tokens[:3]), Subtype: types.Wall, SubtypeKnown: true}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "artifacts", "you", "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledArtifacts, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 4 && effectWordsAt(tokens, 0, "tokens", "you", "control") &&
		(equalWord(tokens[3], "get") || equalWord(tokens[3], "have")):
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledTokens, Span: shared.SpanOf(tokens[:3])}
	case len(tokens) >= 5 && equalWord(tokens[0], "other") && effectWordsAt(tokens, 2, "you", "control", "have"):
		value, ok := subtype(1)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectOtherControlledCreatureSubtype, Span: shared.SpanOf(tokens[:4]), Subtype: value, SubtypeText: tokens[1].Text, SubtypeKnown: ok}
	case len(tokens) >= 4 && effectWordsAt(tokens, 1, "you", "control", "have"):
		value, ok := subtype(0)
		return EffectStaticSubjectSyntax{Kind: EffectStaticSubjectControlledCreatureSubtype, Span: shared.SpanOf(tokens[:3]), Subtype: value, SubtypeText: tokens[0].Text, SubtypeKnown: ok}
	default:
		return EffectStaticSubjectSyntax{}
	}
}

func selectionKindForNoun(noun ObjectNoun) SelectionKind {
	switch noun {
	case ObjectNounArtifact:
		return SelectionArtifact
	case ObjectNounCard:
		return SelectionCard
	case ObjectNounCreature:
		return SelectionCreature
	case ObjectNounEnchantment:
		return SelectionEnchantment
	case ObjectNounLand:
		return SelectionLand
	case ObjectNounOpponent:
		return SelectionOpponent
	case ObjectNounPermanent:
		return SelectionPermanent
	case ObjectNounPlaneswalker:
		return SelectionPlaneswalker
	case ObjectNounPlayer:
		return SelectionPlayer
	case ObjectNounSpell:
		return SelectionSpell
	default:
		return SelectionUnknown
	}
}

func effectWordsAt(tokens []shared.Token, start int, words ...string) bool {
	if start < 0 || start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		if !equalWord(tokens[start+i], word) {
			return false
		}
	}
	return true
}

func effectContainsWords(words []string, sequence ...string) bool {
	for i := 0; i+len(sequence) <= len(words); i++ {
		if slices.Equal(words[i:i+len(sequence)], sequence) {
			return true
		}
	}
	return false
}

func joinedEffectText(tokens []shared.Token) string {
	var builder strings.Builder
	for i, token := range tokens {
		if i > 0 && token.Span.Start.Offset > tokens[i-1].Span.End.Offset {
			_ = builder.WriteByte(' ')
		}
		_, _ = builder.WriteString(token.Text)
	}
	return builder.String()
}
