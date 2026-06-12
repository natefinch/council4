package oracle

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game/types"
)

// recognizeCondition assigns closed semantic data to an exact condition
// phrase. Unrecognized wording remains explicitly unsupported.
func recognizeCondition(condition *CompiledCondition) {
	condition.Predicate = ConditionPredicateUnsupported
	remainder, ok := conditionRemainder(condition.Kind, condition.Text)
	if !ok {
		return
	}

	normalized := strings.ToLower(remainder)
	condition.Negated = condition.Kind == ConditionUnless
	if recognizeTriggerCompositionCondition(condition, normalized) {
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
	case "it had no +1/+1 counters", "it had no +1/+1 counters on it":
		condition.Predicate = ConditionPredicateEventSubjectHadNoCounter
		condition.Counter = ConditionCounterPlusOnePlusOne
	case "it had no -1/-1 counters", "it had no -1/-1 counters on it":
		condition.Predicate = ConditionPredicateEventSubjectHadNoCounter
		condition.Counter = ConditionCounterMinusOneMinusOne
	case "you don't":
		condition.Predicate = ConditionPredicatePriorInstructionNotAccepted
	case "one or more +1/+1 counters would be put on a creature you control":
		condition.Predicate = ConditionPredicateCounterPlacementOnControlledCreature
		condition.Counter = ConditionCounterPlusOnePlusOne
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
		if (strings.HasPrefix(normalized, "its controller pays {") ||
			strings.HasPrefix(normalized, "its controller pays{")) &&
			strings.HasSuffix(normalized, "}") {
			condition.Predicate = ConditionPredicateTargetControllerDoesNotPay
			return
		}
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
		if selection, ok := controllerControlsSelection(remainder); ok {
			condition.Predicate = ConditionPredicateControllerControls
			condition.Selection = selection
			return
		}
		if selection, ok := controlledLandSubtypeSelection(remainder); ok {
			condition.Predicate = ConditionPredicateControllerControls
			condition.Selection = selection
		}
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
	case "you attacked this turn":
		condition.Predicate = ConditionPredicateEventHistory
		condition.EventHistoryPattern = &TriggerPattern{
			Event:      TriggerEventAttackerDeclared,
			Controller: ControllerYou,
		}
		condition.EventHistoryWindow = ConditionEventHistoryWindowCurrentTurn
	case "a creature died this turn":
		condition.Predicate = ConditionPredicateEventHistory
		condition.EventHistoryPattern = &TriggerPattern{
			Event:            TriggerEventPermanentDied,
			SubjectSelection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}},
		}
		condition.EventHistoryWindow = ConditionEventHistoryWindowCurrentTurn
	case "you gained life this turn":
		condition.Predicate = ConditionPredicateEventHistory
		condition.EventHistoryPattern = &TriggerPattern{
			Event:  TriggerEventLifeGained,
			Player: TriggerPlayerYou,
		}
		condition.EventHistoryWindow = ConditionEventHistoryWindowCurrentTurn
	case "an opponent lost life this turn":
		condition.Predicate = ConditionPredicateEventHistory
		condition.EventHistoryPattern = &TriggerPattern{
			Event:  TriggerEventLifeLost,
			Player: TriggerPlayerOpponent,
		}
		condition.EventHistoryWindow = ConditionEventHistoryWindowCurrentTurn
	case "you lost life this turn":
		condition.Predicate = ConditionPredicateEventHistory
		condition.EventHistoryPattern = &TriggerPattern{
			Event:  TriggerEventLifeLost,
			Player: TriggerPlayerYou,
		}
		condition.EventHistoryWindow = ConditionEventHistoryWindowCurrentTurn
	case "an opponent lost life last turn":
		condition.Predicate = ConditionPredicateEventHistory
		condition.EventHistoryPattern = &TriggerPattern{
			Event:  TriggerEventLifeLost,
			Player: TriggerPlayerOpponent,
		}
		condition.EventHistoryWindow = ConditionEventHistoryWindowPreviousTurn
	case "you lost life last turn":
		condition.Predicate = ConditionPredicateEventHistory
		condition.EventHistoryPattern = &TriggerPattern{
			Event:  TriggerEventLifeLost,
			Player: TriggerPlayerYou,
		}
		condition.EventHistoryWindow = ConditionEventHistoryWindowPreviousTurn
	case "no spells were cast last turn":
		condition.Predicate = ConditionPredicateEventHistory
		condition.Negated = true
		condition.EventHistoryPattern = &TriggerPattern{Event: TriggerEventSpellCast}
		condition.EventHistoryWindow = ConditionEventHistoryWindowPreviousTurn
	default:
		return false
	}
	return true
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
	subject := Span{
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

func controllerControlsSelection(remainder string) (ConditionSelection, bool) {
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
		selection, ok := conditionSelectionForNoun(noun)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ExcludeSource = prefix.excludeSource
		return selection, true
	}
	return ConditionSelection{}, false
}

func controlledLandSubtypeSelection(remainder string) (ConditionSelection, bool) {
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
		subtype, ok := canonicalConditionSubtype(name, types.Land)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.SubtypesAny = append(selection.SubtypesAny, subtype)
	}
	return selection, true
}

func conditionSelectionForNoun(noun string) (ConditionSelection, bool) {
	switch strings.ToLower(noun) {
	case "artifact":
		return ConditionSelection{RequiredTypes: []ConditionCardType{ConditionCardTypeArtifact}}, true
	case "artifact creature":
		return ConditionSelection{RequiredTypes: []ConditionCardType{ConditionCardTypeArtifact, ConditionCardTypeCreature}}, true
	case "battle":
		return ConditionSelection{RequiredTypes: []ConditionCardType{ConditionCardTypeBattle}}, true
	case "creature":
		return ConditionSelection{RequiredTypes: []ConditionCardType{ConditionCardTypeCreature}}, true
	case "enchantment":
		return ConditionSelection{RequiredTypes: []ConditionCardType{ConditionCardTypeEnchantment}}, true
	case "land":
		return ConditionSelection{RequiredTypes: []ConditionCardType{ConditionCardTypeLand}}, true
	case "planeswalker":
		return ConditionSelection{RequiredTypes: []ConditionCardType{ConditionCardTypePlaneswalker}}, true
	case "snow land":
		return ConditionSelection{
			RequiredTypes: []ConditionCardType{ConditionCardTypeLand},
			Supertypes:    []ConditionSupertype{ConditionSupertypeSnow},
		}, true
	default:
	}
	if selection, ok := colorQualifiedConditionSelection(noun); ok {
		return selection, true
	}
	for _, cardType := range []types.Card{types.Creature, types.Land} {
		if subtype, ok := canonicalConditionSubtype(noun, cardType); ok {
			return ConditionSelection{SubtypesAny: []string{subtype}}, true
		}
	}
	const planeswalkerSuffix = " planeswalker"
	if strings.HasSuffix(strings.ToLower(noun), planeswalkerSuffix) {
		name := noun[:len(noun)-len(planeswalkerSuffix)]
		if subtype, ok := canonicalConditionSubtype(name, types.Planeswalker); ok {
			return ConditionSelection{
				RequiredTypes: []ConditionCardType{ConditionCardTypePlaneswalker},
				SubtypesAny:   []string{subtype},
			}, true
		}
	}
	return ConditionSelection{}, false
}

func colorQualifiedConditionSelection(noun string) (ConditionSelection, bool) {
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
		color, ok := conditionColor(part)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, color)
	}
	return selection, len(selection.ColorsAny) > 0
}

func conditionColor(word string) (ConditionColor, bool) {
	switch word {
	case "white":
		return ConditionColorWhite, true
	case "blue":
		return ConditionColorBlue, true
	case "black":
		return ConditionColorBlack, true
	case "red":
		return ConditionColorRed, true
	case "green":
		return ConditionColorGreen, true
	default:
		return ConditionColorUnknown, false
	}
}

func canonicalConditionSubtype(name string, cardType types.Card) (string, bool) {
	candidates := []string{name}
	if name != "" {
		candidates = append(candidates, strings.ToUpper(name[:1])+strings.ToLower(name[1:]))
	}
	for _, candidate := range candidates {
		subtype := types.Sub(candidate)
		if types.KnownSubtypeForType(cardType, subtype) {
			return string(subtype), true
		}
	}
	return "", false
}
