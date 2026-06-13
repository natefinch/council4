package compiler

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// recognizeCondition assigns closed semantic data to an exact condition
// phrase. Unrecognized wording remains explicitly unsupported.
func recognizeCondition(
	condition *CompiledCondition,
	phrase []shared.Token,
	atoms parser.Atoms,
	eventHistories []parser.EventHistoryCondition,
) {
	condition.Predicate = ConditionPredicateUnsupported
	if recognizeEventHistoryCondition(condition, eventHistories) {
		return
	}
	remainder, ok := conditionRemainder(condition.Kind, condition.Text)
	if !ok {
		return
	}

	normalized := strings.ToLower(remainder)
	condition.Negated = condition.Kind == ConditionUnless
	if recognizeTriggerCompositionCondition(condition, normalized) {
		return
	}
	if recognizeCounterCondition(condition, normalized, phrase, atoms) {
		return
	}

	switch normalized {
	case "there are seven or more cards in your graveyard", "seven or more cards are in your graveyard":
		condition.Predicate = ConditionPredicateControllerGraveyardCardCountAtLeast
		condition.Threshold = 7
	case "there are four or more card types among cards in your graveyard":
		condition.Predicate = ConditionPredicateControllerGraveyardCardTypeCountAtLeast
		condition.Threshold = 4
	case "you control three or more artifacts":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Threshold = 3
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeArtifact}
	case "you have no cards in hand":
		condition.Predicate = ConditionPredicateControllerHandEmpty
	case "you have seven or more cards in hand":
		condition.Predicate = ConditionPredicateControllerHandSizeAtLeast
		condition.Threshold = 7
	case "you control a creature with power 4 or greater":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
		condition.Selection.PowerAtLeast = 4
		condition.Selection.MatchPowerAtLeast = true
	case "you control three or more creatures with different powers":
		condition.Predicate = ConditionPredicateControllerCreaturePowerDiversityAtLeast
		condition.Threshold = 3
	case "a player has 13 or less life":
		condition.Predicate = ConditionPredicateAnyPlayerLifeAtMost
		condition.Threshold = 13
	case "you have two or more opponents":
		condition.Predicate = ConditionPredicateOpponentCountAtLeast
		condition.Threshold = 2
	case "an opponent controls two or more lands":
		condition.Predicate = ConditionPredicateAnyOpponentControls
		condition.Threshold = 2
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeLand}
	case "your opponents control eight or more lands":
		condition.Predicate = ConditionPredicateOpponentsControl
		condition.Threshold = 8
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeLand}
	case "you control two or more basic lands":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Threshold = 2
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeLand}
		condition.Selection.Supertypes = []ConditionSupertype{ConditionSupertypeBasic}
	case "you control two or more other lands":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Threshold = 2
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeLand}
		condition.Selection.ExcludeSource = true
	case "you control two or fewer other lands":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Threshold = 3
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeLand}
		condition.Selection.ExcludeSource = true
		condition.Negated = condition.Kind != ConditionUnless
	case "it was kicked":
		condition.Predicate = ConditionPredicateEventSubjectWasKicked
	case "it was cast":
		condition.Predicate = ConditionPredicateEventSubjectWasCast
	case "you cast it":
		condition.Predicate = ConditionPredicateEventSubjectWasCastByController
	case "you don't":
		condition.Predicate = ConditionPredicatePriorInstructionNotAccepted
	case "you would put one or more counters on a permanent or player":
		condition.Predicate = ConditionPredicateControllerCounterPlacement
	case "another red source you control would deal damage to a permanent or player":
		condition.Predicate = ConditionPredicateDamageByControlledSource
		condition.Selection.ColorsAny = []ConditionColor{ConditionColorRed}
		condition.Selection.ExcludeSource = true
	case "a red source you control would deal damage to a permanent or player":
		condition.Predicate = ConditionPredicateDamageByControlledSource
		condition.Selection.ColorsAny = []ConditionColor{ConditionColorRed}
	case "a source you control would deal damage to a permanent or player":
		condition.Predicate = ConditionPredicateDamageByControlledSource
	case "an effect would create one or more tokens under your control":
		condition.Predicate = ConditionPredicateTokenCreationUnderController
	default:
		if strings.HasSuffix(normalized, " would die") {
			condition.Predicate = ConditionPredicateSourceWouldDie
			return
		}
		if strings.HasSuffix(normalized, " would be put into a graveyard from anywhere") {
			condition.Predicate = ConditionPredicateSourceWouldGoToGraveyard
			return
		}
		if threshold, ok := controllerLifeThreshold(normalized); ok {
			condition.Predicate = ConditionPredicateControllerLifeAtLeast
			condition.Threshold = threshold
			return
		}
		if selection, ok := controllerControlsSelection(remainder, phrase, atoms); ok {
			condition.Predicate = ConditionPredicateControllerControls
			condition.Selection = selection
			return
		}
		if selection, ok := controlledLandSubtypeSelection(remainder, phrase, atoms); ok {
			condition.Predicate = ConditionPredicateControllerControls
			condition.Selection = selection
		}
	}
}

func recognizeCounterCondition(condition *CompiledCondition, normalized string, phrase []shared.Token, atoms parser.Atoms) bool {
	counterValue, ok := conditionCounterAtom(shared.SpanOf(phrase), atoms)
	if !ok {
		return false
	}
	switch {
	case strings.HasPrefix(normalized, "it had no ") &&
		(strings.HasSuffix(normalized, " counters") || strings.HasSuffix(normalized, " counters on it")):
		condition.Predicate = ConditionPredicateEventSubjectHadNoCounter
		condition.Counter = counterValue
		return true
	case strings.HasPrefix(normalized, "one or more ") &&
		strings.HasSuffix(normalized, " counters would be put on a creature you control"):
		condition.Predicate = ConditionPredicateCounterPlacementOnControlledCreature
		condition.Counter = counterValue
		return true
	default:
		return false
	}
}

func conditionCounterAtom(span shared.Span, atoms parser.Atoms) (ConditionCounter, bool) {
	kind, _, ok := atoms.CounterIn(span)
	if !ok {
		return ConditionCounterUnknown, false
	}
	switch kind {
	case counter.PlusOnePlusOne:
		return ConditionCounterPlusOnePlusOne, true
	case counter.MinusOneMinusOne:
		return ConditionCounterMinusOneMinusOne, true
	default:
		return ConditionCounterUnknown, false
	}
}

func recognizeTriggerCompositionCondition(condition *CompiledCondition, normalized string) bool {
	switch normalized {
	case "you control two or more gates":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Threshold = 2
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeLand}
		condition.Selection.SubtypesAny = []string{string(types.Gate)}
	case "you control two or more tapped creatures":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Threshold = 2
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
		condition.Selection.Tapped = ConditionTriTrue
	case "you control a creature with power 5 or greater":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
		condition.Selection.PowerAtLeast = 5
		condition.Selection.MatchPowerAtLeast = true
	case "you control another creature with power 4 or greater":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
		condition.Selection.ExcludeSource = true
		condition.Selection.PowerAtLeast = 4
		condition.Selection.MatchPowerAtLeast = true
	case "you control an equipment":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeArtifact}
		condition.Selection.SubtypesAny = []string{string(types.Equipment)}
	case "you control no creatures":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Negated = true
		condition.Threshold = 1
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
	case "you control three or more creatures":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Threshold = 3
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
	case "you control a tapped creature":
		condition.Predicate = ConditionPredicateControllerControls
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
		condition.Selection.Tapped = ConditionTriTrue
	case "it was a creature", "it's a creature":
		condition.Predicate = ConditionPredicateObjectMatches
		condition.ObjectBinding = ReferenceBindingEventPermanent
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
	case "it was a human":
		condition.Predicate = ConditionPredicateObjectMatches
		condition.ObjectBinding = ReferenceBindingEventPermanent
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
		condition.Selection.SubtypesAny = []string{string(types.Human)}
	case "it had counters on it":
		condition.Predicate = ConditionPredicateEventSubjectHadCounters
		condition.ObjectBinding = ReferenceBindingEventPermanent
	case "this artifact is untapped":
		condition.Predicate = ConditionPredicateObjectMatches
		condition.ObjectBinding = ReferenceBindingSource
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeArtifact}
		condition.Selection.Tapped = ConditionTriFalse
	case "this creature is untapped":
		condition.Predicate = ConditionPredicateObjectMatches
		condition.ObjectBinding = ReferenceBindingSource
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
		condition.Selection.Tapped = ConditionTriFalse
	case "this permanent is an enchantment":
		condition.Predicate = ConditionPredicateObjectMatches
		condition.ObjectBinding = ReferenceBindingSource
		condition.Selection.RequiredTypes = []ConditionCardType{ConditionCardTypeEnchantment}
	case "this creature is on the battlefield":
		condition.Predicate = ConditionPredicateObjectExists
		condition.ObjectBinding = ReferenceBindingSource
	default:
		return false
	}
	return true
}

func recognizeEventHistoryCondition(
	condition *CompiledCondition,
	syntax []parser.EventHistoryCondition,
) bool {
	for i := range syntax {
		if syntax[i].Span != condition.Span {
			continue
		}
		pattern, ok := compileEventHistoryPattern(&syntax[i])
		if !ok {
			return false
		}
		window, ok := compileEventHistoryWindow(syntax[i].Window.Kind)
		if !ok {
			return false
		}
		condition.Predicate = ConditionPredicateEventHistory
		condition.Negated = syntax[i].Negated
		condition.EventHistoryPattern = &pattern
		condition.EventHistoryWindow = window
		return true
	}
	return false
}

func compileEventHistoryPattern(syntax *parser.EventHistoryCondition) (TriggerPattern, bool) {
	if syntax.TriggerEvent != nil && syntax.PlayerEvent != nil ||
		syntax.TriggerEvent == nil && syntax.PlayerEvent == nil {
		return TriggerPattern{}, false
	}
	if syntax.TriggerEvent != nil {
		return compileTriggerEventClause(syntax.TriggerEvent)
	}
	pattern := compilePlayerEventTriggerPattern(syntax.PlayerEvent, TriggerWhenever, nil)
	if pattern.Event == TriggerEventUnknown {
		return TriggerPattern{}, false
	}
	pattern.Kind = TriggerUnknown
	return pattern, true
}

func compileEventHistoryWindow(
	window parser.EventHistoryWindowKind,
) (ConditionEventHistoryWindow, bool) {
	switch window {
	case parser.EventHistoryWindowCurrentTurn:
		return ConditionEventHistoryWindowCurrentTurn, true
	case parser.EventHistoryWindowPreviousTurn:
		return ConditionEventHistoryWindowPreviousTurn, true
	default:
		return ConditionEventHistoryWindowCurrentTurn, false
	}
}

func bindConditionReferences(conditions []CompiledCondition, references []CompiledReference, trigger *CompiledTrigger) {
	for i := range conditions {
		switch conditions[i].Predicate {
		case ConditionPredicateSourceWouldDie:
			if !conditionSubjectBindsSource(conditions[i], references, " would die") {
				conditions[i].Predicate = ConditionPredicateUnsupported
			}
		case ConditionPredicateSourceWouldGoToGraveyard:
			if !conditionSubjectBindsSource(
				conditions[i],
				references,
				" would be put into a graveyard from anywhere",
			) {
				conditions[i].Predicate = ConditionPredicateUnsupported
			}
		case ConditionPredicateObjectMatches,
			ConditionPredicateObjectExists,
			ConditionPredicateEventSubjectHadCounters:
			binding, ok := conditionObjectBinding(conditions[i], references)
			if !ok ||
				binding == ReferenceBindingEventPermanent &&
					(trigger == nil || trigger.Pattern.OneOrMore || !triggerEventBindsPermanent(trigger.Pattern.Event)) ||
				conditions[i].Predicate == ConditionPredicateObjectExists && binding != ReferenceBindingSource ||
				conditions[i].Predicate == ConditionPredicateEventSubjectHadCounters && binding != ReferenceBindingEventPermanent {
				conditions[i].Predicate = ConditionPredicateUnsupported
				continue
			}
			conditions[i].ObjectBinding = binding
		default:
		}
	}
}

func conditionObjectBinding(condition CompiledCondition, references []CompiledReference) (ReferenceBinding, bool) {
	binding := condition.ObjectBinding
	found := binding == ReferenceBindingSource || binding == ReferenceBindingEventPermanent
	for _, reference := range references {
		if !spanContains(condition.Span, reference.Span) {
			continue
		}
		if reference.Binding != ReferenceBindingSource &&
			reference.Binding != ReferenceBindingEventPermanent {
			return ReferenceBindingUnsupported, false
		}
		if found && reference.Binding != binding {
			return ReferenceBindingUnsupported, false
		}
		binding = reference.Binding
		found = true
	}
	return binding, found
}

func conditionSubjectBindsSource(
	condition CompiledCondition,
	references []CompiledReference,
	suffix string,
) bool {
	const prefix = "If "
	if !strings.HasPrefix(condition.Text, prefix) || !strings.HasSuffix(condition.Text, suffix) {
		return false
	}
	subject := shared.Span{
		Start: condition.Span.Start,
		End:   condition.Span.End,
	}
	subject.Start.Offset += len(prefix)
	subject.Start.Column += len(prefix)
	subject.End.Offset -= len(suffix)
	subject.End.Column -= len(suffix)
	for _, reference := range references {
		if reference.Span == subject && reference.Binding == ReferenceBindingSource {
			return true
		}
	}
	return false
}

func conditionRemainder(kind ConditionKind, text string) (string, bool) {
	prefix := ""
	switch kind {
	case ConditionIf:
		prefix = "if "
	case ConditionUnless:
		prefix = "unless "
	case ConditionOnlyIf:
		prefix = "only if "
	case ConditionAsLongAs:
		prefix = "as long as "
	default:
		return "", false
	}
	if len(text) < len(prefix) || !strings.EqualFold(text[:len(prefix)], prefix) {
		return "", false
	}
	remainder := text[len(prefix):]
	return remainder, remainder != "" && strings.TrimSpace(remainder) == remainder
}

func controllerLifeThreshold(remainder string) (int, bool) {
	const (
		prefix = "you have "
		suffix = " or more life"
	)
	if !strings.HasPrefix(remainder, prefix) || !strings.HasSuffix(remainder, suffix) {
		return 0, false
	}
	value, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(remainder, prefix), suffix))
	return value, err == nil && value > 0
}

func controllerControlsSelection(remainder string, phrase []shared.Token, atoms parser.Atoms) (ConditionSelection, bool) {
	lowered := strings.ToLower(remainder)
	prefixes := []struct {
		text          string
		excludeSource bool
	}{
		{"you control a ", false},
		{"you control an ", false},
		{"you control another ", true},
	}
	for _, prefix := range prefixes {
		if !strings.HasPrefix(lowered, prefix.text) {
			continue
		}
		noun := remainder[len(prefix.text):]
		if strings.Contains(strings.ToLower(noun), " or ") {
			selection, ok := colorQualifiedConditionSelection(noun, phrase, atoms)
			if !ok {
				return ConditionSelection{}, false
			}
			selection.ExcludeSource = prefix.excludeSource
			return selection, true
		}
		selection, ok := conditionSelectionForNoun(noun, phrase, atoms)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ExcludeSource = prefix.excludeSource
		return selection, true
	}
	return ConditionSelection{}, false
}

func controlledLandSubtypeSelection(remainder string, phrase []shared.Token, atoms parser.Atoms) (ConditionSelection, bool) {
	const prefix = "you control "
	if len(remainder) < len(prefix) || !strings.EqualFold(remainder[:len(prefix)], prefix) {
		return ConditionSelection{}, false
	}
	parts := strings.Split(remainder[len(prefix):], " or ")
	if len(parts) == 0 || len(parts) > 2 {
		return ConditionSelection{}, false
	}
	selection := ConditionSelection{RequiredTypes: []ConditionCardType{ConditionCardTypeLand}}
	for _, part := range parts {
		name := strings.TrimPrefix(strings.TrimPrefix(part, "a "), "an ")
		span, ok := wordTokenSpan(phrase, name)
		if !ok {
			return ConditionSelection{}, false
		}
		subtype, ok := conditionSubtypeAtom(span, atoms, parser.CardTypeLand)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.SubtypesAny = append(selection.SubtypesAny, subtype)
	}
	return selection, true
}

func conditionSelectionForNoun(noun string, phrase []shared.Token, atoms parser.Atoms) (ConditionSelection, bool) {
	nounSpan, ok := phraseSuffixSpan(phrase, noun)
	if !ok {
		return ConditionSelection{}, false
	}
	nounTokens := tokensInSpan(phrase, nounSpan)
	switch strings.ToLower(noun) {
	case "artifact":
		return conditionSelectionForCardTypes(nounTokens, atoms, parser.CardTypeArtifact)
	case "artifact creature":
		return conditionSelectionForCardTypes(nounTokens, atoms, parser.CardTypeArtifact, parser.CardTypeCreature)
	case "battle":
		return conditionSelectionForCardTypes(nounTokens, atoms, parser.CardTypeBattle)
	case "creature":
		return conditionSelectionForCardTypes(nounTokens, atoms, parser.CardTypeCreature)
	case "enchantment":
		return conditionSelectionForCardTypes(nounTokens, atoms, parser.CardTypeEnchantment)
	case "land":
		return conditionSelectionForCardTypes(nounTokens, atoms, parser.CardTypeLand)
	case "planeswalker":
		return conditionSelectionForCardTypes(nounTokens, atoms, parser.CardTypePlaneswalker)
	case "snow land":
		if len(nounTokens) != 2 {
			return ConditionSelection{}, false
		}
		supertype, superOK := atoms.SupertypeAt(nounTokens[0].Span)
		cardType, typeOK := atoms.CardTypeAt(nounTokens[1].Span)
		if !superOK || supertype != parser.SupertypeSnow || !typeOK || cardType != parser.CardTypeLand {
			return ConditionSelection{}, false
		}
		return ConditionSelection{
			RequiredTypes: []ConditionCardType{ConditionCardTypeLand},
			Supertypes:    []ConditionSupertype{ConditionSupertypeSnow},
		}, true
	default:
	}
	if selection, ok := colorQualifiedConditionSelection(noun, phrase, atoms); ok {
		return selection, true
	}
	if subtype, ok := conditionSubtypeAtom(nounSpan, atoms, parser.CardTypeCreature); ok {
		return ConditionSelection{SubtypesAny: []string{subtype}}, true
	}
	if subtype, ok := conditionSubtypeAtom(nounSpan, atoms, parser.CardTypeLand); ok {
		return ConditionSelection{SubtypesAny: []string{subtype}}, true
	}
	if len(nounTokens) >= 2 && equalWord(nounTokens[len(nounTokens)-1], "planeswalker") {
		nameSpan := shared.SpanOf(nounTokens[:len(nounTokens)-1])
		if subtype, ok := conditionSubtypeAtom(nameSpan, atoms, parser.CardTypePlaneswalker); ok {
			return ConditionSelection{
				RequiredTypes: []ConditionCardType{ConditionCardTypePlaneswalker},
				SubtypesAny:   []string{subtype},
			}, true
		}
	}
	return ConditionSelection{}, false
}

func conditionSelectionForCardTypes(tokens []shared.Token, atoms parser.Atoms, cardTypes ...parser.CardType) (ConditionSelection, bool) {
	if len(tokens) != len(cardTypes) {
		return ConditionSelection{}, false
	}
	selection := ConditionSelection{RequiredTypes: make([]ConditionCardType, 0, len(cardTypes))}
	for i, want := range cardTypes {
		cardType, ok := atoms.CardTypeAt(tokens[i].Span)
		if !ok || cardType != want {
			return ConditionSelection{}, false
		}
		compiled, ok := conditionCardType(cardType)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.RequiredTypes = append(selection.RequiredTypes, compiled)
	}
	return selection, true
}

func conditionCardType(cardType parser.CardType) (ConditionCardType, bool) {
	switch cardType {
	case parser.CardTypeArtifact:
		return ConditionCardTypeArtifact, true
	case parser.CardTypeBattle:
		return ConditionCardTypeBattle, true
	case parser.CardTypeCreature:
		return ConditionCardTypeCreature, true
	case parser.CardTypeEnchantment:
		return ConditionCardTypeEnchantment, true
	case parser.CardTypeLand:
		return ConditionCardTypeLand, true
	case parser.CardTypePlaneswalker:
		return ConditionCardTypePlaneswalker, true
	default:
		return ConditionCardTypeUnknown, false
	}
}

func conditionSupertype(supertype parser.Supertype) (ConditionSupertype, bool) {
	switch supertype {
	case parser.SupertypeSnow:
		return ConditionSupertypeSnow, true
	default:
		return ConditionSupertypeUnknown, false
	}
}

func conditionColor(color parser.Color) (ConditionColor, bool) {
	switch color {
	case parser.ColorWhite:
		return ConditionColorWhite, true
	case parser.ColorBlue:
		return ConditionColorBlue, true
	case parser.ColorBlack:
		return ConditionColorBlack, true
	case parser.ColorRed:
		return ConditionColorRed, true
	case parser.ColorGreen:
		return ConditionColorGreen, true
	default:
		return ConditionColorUnknown, false
	}
}

func conditionSubtypeAtom(span shared.Span, atoms parser.Atoms, cardType parser.CardType) (string, bool) {
	sub, ok := atoms.SubtypeAt(span)
	if !ok {
		return "", false
	}
	if !parser.SubtypeMatchesCardType(sub, cardType) {
		return "", false
	}
	return string(sub), true
}

func colorQualifiedConditionSelection(noun string, phrase []shared.Token, atoms parser.Atoms) (ConditionSelection, bool) {
	lowered := strings.ToLower(noun)
	var selection ConditionSelection
	colorsText := ""
	switch {
	case strings.HasSuffix(lowered, " creature"):
		selection.RequiredTypes = []ConditionCardType{ConditionCardTypeCreature}
		colorsText = strings.TrimSuffix(lowered, " creature")
	case strings.HasSuffix(lowered, " permanent"):
		colorsText = strings.TrimSuffix(lowered, " permanent")
	default:
		return ConditionSelection{}, false
	}
	if colorsText == "colorless" {
		selection.Colorless = true
		return selection, true
	}
	for part := range strings.SplitSeq(colorsText, " or ") {
		color, ok := conditionColorAtom(phrase, atoms, part)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, color)
	}
	return selection, len(selection.ColorsAny) > 0
}

// conditionColorAtom resolves a color word to its condition color identity using
// the parser-emitted color atom that spans the matching token, rather than
// re-recognizing the color from spelling.
func conditionColorAtom(phrase []shared.Token, atoms parser.Atoms, word string) (ConditionColor, bool) {
	span, ok := wordTokenSpan(phrase, word)
	if !ok {
		return ConditionColorUnknown, false
	}
	color, ok := atoms.ColorAt(span)
	if !ok {
		return ConditionColorUnknown, false
	}
	switch color {
	case parser.ColorWhite:
		return ConditionColorWhite, true
	case parser.ColorBlue:
		return ConditionColorBlue, true
	case parser.ColorBlack:
		return ConditionColorBlack, true
	case parser.ColorRed:
		return ConditionColorRed, true
	case parser.ColorGreen:
		return ConditionColorGreen, true
	default:
		return ConditionColorUnknown, false
	}
}

// wordTokenSpan returns the span of the first word token in phrase whose text
// equals word (case-insensitively). The condition grammar identifies which word
// fills an atom slot; the span links that slot to its parser-emitted atom.
func wordTokenSpan(phrase []shared.Token, word string) (shared.Span, bool) {
	word = strings.TrimSpace(word)
	for _, token := range phrase {
		if token.Kind == shared.Word && strings.EqualFold(token.Text, word) {
			return token.Span, true
		}
	}
	return shared.Span{}, false
}

func phraseSuffixSpan(phrase []shared.Token, text string) (shared.Span, bool) {
	want := strings.ToLower(strings.TrimSpace(text))
	for start := range phrase {
		if strings.EqualFold(joinedSourceText(phrase[start:]), want) {
			return shared.SpanOf(phrase[start:]), true
		}
	}
	return shared.Span{}, false
}

func tokensInSpan(tokens []shared.Token, span shared.Span) []shared.Token {
	var result []shared.Token
	for _, token := range tokens {
		if token.Span.Start.Offset >= span.Start.Offset && token.Span.End.Offset <= span.End.Offset {
			result = append(result, token)
		}
	}
	return result
}
