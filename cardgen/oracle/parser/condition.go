package parser

import (
	"slices"
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// ConditionIntroKind identifies the grammatical introducer that opens a typed
// condition clause.
type ConditionIntroKind string

// Condition introducers recognized by the parser.
const (
	ConditionIntroUnknown  ConditionIntroKind = ""
	ConditionIntroIf       ConditionIntroKind = "ConditionIntroIf"
	ConditionIntroUnless   ConditionIntroKind = "ConditionIntroUnless"
	ConditionIntroOnlyIf   ConditionIntroKind = "ConditionIntroOnlyIf"
	ConditionIntroAsLongAs ConditionIntroKind = "ConditionIntroAsLongAs"
)

// ConditionPredicateKind identifies the closed predicate family recognized in a
// condition clause. The compiler maps these mechanically onto its semantic
// predicate vocabulary.
type ConditionPredicateKind string

// Condition predicates recognized by the parser.
const (
	ConditionPredicateUnknown                              ConditionPredicateKind = ""
	ConditionPredicateControllerLifeAtLeast                ConditionPredicateKind = "ConditionPredicateControllerLifeAtLeast"
	ConditionPredicateControllerHandSizeAtLeast            ConditionPredicateKind = "ConditionPredicateControllerHandSizeAtLeast"
	ConditionPredicateControllerHandEmpty                  ConditionPredicateKind = "ConditionPredicateControllerHandEmpty"
	ConditionPredicateAnyPlayerLifeAtMost                  ConditionPredicateKind = "ConditionPredicateAnyPlayerLifeAtMost"
	ConditionPredicateOpponentCountAtLeast                 ConditionPredicateKind = "ConditionPredicateOpponentCountAtLeast"
	ConditionPredicateControls                             ConditionPredicateKind = "ConditionPredicateControls"
	ConditionPredicateGraveyardCardCountAtLeast            ConditionPredicateKind = "ConditionPredicateGraveyardCardCountAtLeast"
	ConditionPredicateGraveyardCardTypeCountAtLeast        ConditionPredicateKind = "ConditionPredicateGraveyardCardTypeCountAtLeast"
	ConditionPredicateCreaturePowerDiversityAtLeast        ConditionPredicateKind = "ConditionPredicateCreaturePowerDiversityAtLeast"
	ConditionPredicateEventSubjectWasKicked                ConditionPredicateKind = "ConditionPredicateEventSubjectWasKicked"
	ConditionPredicateEventSubjectWasCast                  ConditionPredicateKind = "ConditionPredicateEventSubjectWasCast"
	ConditionPredicateEventSubjectWasCastByController      ConditionPredicateKind = "ConditionPredicateEventSubjectWasCastByController"
	ConditionPredicateEventSubjectHadNoCounter             ConditionPredicateKind = "ConditionPredicateEventSubjectHadNoCounter"
	ConditionPredicateEventSubjectHadCounters              ConditionPredicateKind = "ConditionPredicateEventSubjectHadCounters"
	ConditionPredicatePriorInstructionNotAccepted          ConditionPredicateKind = "ConditionPredicatePriorInstructionNotAccepted"
	ConditionPredicatePriorInstructionAccepted             ConditionPredicateKind = "ConditionPredicatePriorInstructionAccepted"
	ConditionPredicateCounterPlacementOnControlledCreature ConditionPredicateKind = "ConditionPredicateCounterPlacementOnControlledCreature"
	ConditionPredicateControllerCounterPlacement           ConditionPredicateKind = "ConditionPredicateControllerCounterPlacement"
	ConditionPredicateDamageByControlledSource             ConditionPredicateKind = "ConditionPredicateDamageByControlledSource"
	ConditionPredicateTokenCreationUnderController         ConditionPredicateKind = "ConditionPredicateTokenCreationUnderController"
	ConditionPredicateSourceWouldDie                       ConditionPredicateKind = "ConditionPredicateSourceWouldDie"
	ConditionPredicateSourceWouldGoToGraveyard             ConditionPredicateKind = "ConditionPredicateSourceWouldGoToGraveyard"
	ConditionPredicateObjectMatches                        ConditionPredicateKind = "ConditionPredicateObjectMatches"
	ConditionPredicateObjectExists                         ConditionPredicateKind = "ConditionPredicateObjectExists"
)

// ConditionControlScope identifies which players' battlefields a "controls"
// predicate counts.
type ConditionControlScope string

// Control scopes recognized by the parser.
const (
	ConditionControlScopeController  ConditionControlScope = ""
	ConditionControlScopeAnyOpponent ConditionControlScope = "ConditionControlScopeAnyOpponent"
	ConditionControlScopeOpponents   ConditionControlScope = "ConditionControlScopeOpponents"
)

// ConditionComparison identifies the numeric comparison a count predicate uses.
type ConditionComparison string

// Condition comparisons recognized by the parser. ConditionComparisonNone marks
// a singular "a"/"an"/"another" selection with no explicit count.
const (
	ConditionComparisonNone    ConditionComparison = ""
	ConditionComparisonAtLeast ConditionComparison = "ConditionComparisonAtLeast"
	ConditionComparisonAtMost  ConditionComparison = "ConditionComparisonAtMost"
)

// ConditionTappedState is a typed tapped-state selection filter.
type ConditionTappedState string

// Tapped-state filters recognized by the parser.
const (
	ConditionTappedAny   ConditionTappedState = ""
	ConditionTappedTrue  ConditionTappedState = "ConditionTappedTrue"
	ConditionTappedFalse ConditionTappedState = "ConditionTappedFalse"
)

// ConditionSupertype identifies a supertype in a typed condition selection.
type ConditionSupertype string

// Condition supertypes recognized by the parser.
const (
	ConditionSupertypeUnknown ConditionSupertype = ""
	ConditionSupertypeBasic   ConditionSupertype = "ConditionSupertypeBasic"
	ConditionSupertypeSnow    ConditionSupertype = "ConditionSupertypeSnow"
)

// ConditionCounterKind identifies a counter mentioned by a condition clause.
type ConditionCounterKind string

// Condition counter kinds recognized by the parser.
const (
	ConditionCounterNone             ConditionCounterKind = ""
	ConditionCounterPlusOnePlusOne   ConditionCounterKind = "ConditionCounterPlusOnePlusOne"
	ConditionCounterMinusOneMinusOne ConditionCounterKind = "ConditionCounterMinusOneMinusOne"
)

// ConditionObjectBinding identifies the object a state predicate inspects.
type ConditionObjectBinding string

// Object bindings recognized by the parser.
const (
	ConditionObjectBindingNone           ConditionObjectBinding = ""
	ConditionObjectBindingSource         ConditionObjectBinding = "ConditionObjectBindingSource"
	ConditionObjectBindingEventPermanent ConditionObjectBinding = "ConditionObjectBindingEventPermanent"
)

// ConditionSelection is the source-independent permanent selection used by typed
// condition clauses. Subtype names are canonical typed identities.
type ConditionSelection struct {
	RequiredTypes     []TriggerCardType    `json:",omitempty"`
	Supertypes        []ConditionSupertype `json:",omitempty"`
	SubtypesAny       []types.Sub          `json:",omitempty"`
	ColorsAny         []TriggerColor       `json:",omitempty"`
	Colorless         bool                 `json:",omitempty"`
	Multicolored      bool                 `json:",omitempty"`
	TokenOnly         bool                 `json:",omitempty"`
	ExcludeSource     bool                 `json:",omitempty"`
	Tapped            ConditionTappedState `json:",omitempty"`
	PowerAtLeast      int                  `json:",omitempty"`
	MatchPowerAtLeast bool                 `json:",omitempty"`
	// TotalPowerAtLeast is the collective-power threshold for a "have total
	// power <n> or greater" qualifier, applied to the selected permanents as a
	// group rather than to each permanent individually. MatchTotalPowerAtLeast
	// marks the threshold present so a zero threshold remains expressible.
	TotalPowerAtLeast      int  `json:",omitempty"`
	MatchTotalPowerAtLeast bool `json:",omitempty"`
}

// ConditionClause is composable typed syntax for a supported condition. The
// parser owns the Oracle vocabulary, normalization, and grammar; the compiler
// maps these typed fields mechanically without inspecting source text.
type ConditionClause struct {
	Span      shared.Span            `json:"-"`
	Intro     ConditionIntroKind     `json:",omitempty"`
	Predicate ConditionPredicateKind `json:",omitempty"`

	// Scope, Comparison, and CompareValue describe a "controls" predicate. For
	// other predicates Threshold carries the literal numeric parameter.
	Scope        ConditionControlScope `json:",omitempty"`
	Comparison   ConditionComparison   `json:",omitempty"`
	CompareValue int                   `json:",omitempty"`
	Threshold    int                   `json:",omitempty"`

	Selection     ConditionSelection     `json:",omitzero"`
	Counter       ConditionCounterKind   `json:",omitempty"`
	ObjectBinding ConditionObjectBinding `json:",omitempty"`

	// SubjectSpan is set for source-death predicates so the compiler can confirm
	// the subject binds the source via a typed reference.
	SubjectSpan    shared.Span `json:"-"`
	HasSubjectSpan bool        `json:",omitempty"`
	// SubjectRefID is the parser-assigned NodeID of the reference that fills the
	// subject span for source-death predicates, or -1 when no reference does. The
	// compiler confirms the subject binds the source by matching this identity
	// instead of comparing the reference span to the subject span.
	SubjectRefID int `json:"-"`
}

func emitConditionClauses(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		if clauses := parseConditionClauses(tokens, ability.Atoms); len(clauses) > 0 {
			ability.ConditionClauses = clauses
		}
		if ability.Modal == nil {
			continue
		}
		for j := range ability.Modal.Options {
			mode := &ability.Modal.Options[j]
			tokens := eventHistorySemanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
			if clauses := parseConditionClauses(tokens, mode.Atoms); len(clauses) > 0 {
				mode.ConditionClauses = clauses
			}
		}
	}
}

func parseConditionClauses(tokens []shared.Token, atoms Atoms) []ConditionClause {
	var clauses []ConditionClause
	for i := 0; i < len(tokens); i++ {
		intro, width := conditionIntroAt(tokens, i)
		if intro == ConditionIntroUnknown {
			continue
		}
		end := conditionClauseEnd(tokens, i)
		if clause, ok := parseConditionClause(tokens[i:end], width, intro, atoms); ok {
			clause.Span = shared.SpanOf(tokens[i:end])
			clauses = append(clauses, clause)
		}
		i = end - 1
	}
	return clauses
}

func conditionIntroAt(tokens []shared.Token, index int) (kind ConditionIntroKind, width int) {
	switch {
	case equalWord(tokens[index], "if"):
		return ConditionIntroIf, 1
	case equalWord(tokens[index], "unless"):
		return ConditionIntroUnless, 1
	case index+1 < len(tokens) &&
		equalWord(tokens[index], "only") &&
		equalWord(tokens[index+1], "if"):
		return ConditionIntroOnlyIf, 2
	case index+2 < len(tokens) &&
		equalWord(tokens[index], "as") &&
		equalWord(tokens[index+1], "long") &&
		equalWord(tokens[index+2], "as"):
		return ConditionIntroAsLongAs, 3
	default:
		return ConditionIntroUnknown, 0
	}
}

func conditionClauseEnd(tokens []shared.Token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Kind == shared.Period || i > start && tokens[i].Kind == shared.Comma {
			return i
		}
	}
	return len(tokens)
}

func parseConditionClause(
	tokens []shared.Token,
	introWidth int,
	intro ConditionIntroKind,
	atoms Atoms,
) (ConditionClause, bool) {
	body := tokens[introWidth:]
	if len(body) == 0 {
		return ConditionClause{}, false
	}
	clause, ok := recognizeConditionPredicate(body, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	clause.Intro = intro
	return clause, true
}

func recognizeConditionPredicate(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	for _, recognize := range []func([]shared.Token, Atoms) (ConditionClause, bool){
		recognizePriorInstructionCondition,
		recognizeEventSubjectCondition,
		recognizeSourceStateCondition,
		recognizeControllerResourceCondition,
		recognizeGraveyardCondition,
		recognizeCounterPlacementCondition,
		recognizeDamageSourceCondition,
		recognizeTokenCreationCondition,
		recognizeControlsCondition,
		recognizeTotalPowerCondition,
		recognizeSourceDeathCondition,
	} {
		if clause, ok := recognize(body, atoms); ok {
			return clause, true
		}
	}
	return ConditionClause{}, false
}

func recognizePriorInstructionCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "don't") {
		return ConditionClause{Predicate: ConditionPredicatePriorInstructionNotAccepted}, true
	}
	if tokenWordsEqual(body, "you", "do") {
		return ConditionClause{Predicate: ConditionPredicatePriorInstructionAccepted}, true
	}
	return ConditionClause{}, false
}

func recognizeEventSubjectCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "cast", "it") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasCastByController}, true
	}
	if tokenWordsEqual(body, "it", "was", "kicked") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasKicked}, true
	}
	if tokenWordsEqual(body, "it", "was", "cast") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasCast}, true
	}
	if tokenWordsEqual(body, "it", "had", "counters", "on", "it") {
		return ConditionClause{
			Predicate:     ConditionPredicateEventSubjectHadCounters,
			ObjectBinding: ConditionObjectBindingEventPermanent,
		}, true
	}
	if clause, ok := recognizeEventSubjectCounterCondition(body, atoms); ok {
		return clause, true
	}
	return recognizeEventSubjectMatchCondition(body, atoms)
}

// recognizeEventSubjectCounterCondition handles "it had no <counter> counters"
// and the optional trailing "on it".
func recognizeEventSubjectCounterCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "it", "had", "no")
	if !ok {
		return ConditionClause{}, false
	}
	if trimmed, ok := stripTokenSuffix(rest, "on", "it"); ok {
		rest = trimmed
	}
	if !tokenSuffixWord(rest, "counters") {
		return ConditionClause{}, false
	}
	counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateEventSubjectHadNoCounter,
		Counter:   counterKind,
	}, true
}

// recognizeEventSubjectMatchCondition handles "it was a <selection>" and the
// "it's a <selection>" contraction, binding the event permanent.
func recognizeEventSubjectMatchCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "it", "was", "a")
	if !ok {
		if rest, ok = cutTokenPrefix(body, "it", "was", "an"); !ok {
			if rest, ok = cutTokenPrefix(body, "it's", "a"); !ok {
				if rest, ok = cutTokenPrefix(body, "it's", "an"); !ok {
					return ConditionClause{}, false
				}
			}
		}
	}
	selection, ok := parseConditionSelection(rest, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingEventPermanent,
		Selection:     selection,
	}, true
}

// recognizeSourceStateCondition handles "this <selection> is <state>" forms that
// inspect the source permanent.
func recognizeSourceStateCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "this")
	if !ok {
		return ConditionClause{}, false
	}
	isIndex := tokenWordIndex(rest, "is")
	if isIndex < 1 {
		return ConditionClause{}, false
	}
	subjectTokens := rest[:isIndex]
	stateTokens := rest[isIndex+1:]
	if tokenWordsEqual(stateTokens, "on", "the", "battlefield") {
		selection, ok := parseConditionSelection(subjectTokens, atoms)
		if !ok || !conditionSelectionEmptyExceptType(selection) {
			return ConditionClause{}, false
		}
		return ConditionClause{
			Predicate:     ConditionPredicateObjectExists,
			ObjectBinding: ConditionObjectBindingSource,
		}, true
	}
	selection, ok := parseConditionSelection(subjectTokens, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	if !applySourceState(stateTokens, atoms, &selection) {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingSource,
		Selection:     selection,
	}, true
}

func applySourceState(stateTokens []shared.Token, atoms Atoms, selection *ConditionSelection) bool {
	switch {
	case tokenWordsEqual(stateTokens, "untapped"):
		selection.Tapped = ConditionTappedFalse
		return true
	case tokenWordsEqual(stateTokens, "tapped"):
		selection.Tapped = ConditionTappedTrue
		return true
	}
	// "this permanent is an enchantment": the state is a typed card type.
	var rest []shared.Token
	if trimmed, ok := cutTokenPrefix(stateTokens, "a"); ok {
		rest = trimmed
	} else if trimmed, ok := cutTokenPrefix(stateTokens, "an"); ok {
		rest = trimmed
	} else {
		return false
	}
	typeSelection, ok := parseConditionSelection(rest, atoms)
	if !ok || len(typeSelection.RequiredTypes) == 0 {
		return false
	}
	selection.RequiredTypes = append(selection.RequiredTypes, typeSelection.RequiredTypes...)
	selection.SubtypesAny = append(selection.SubtypesAny, typeSelection.SubtypesAny...)
	selection.ColorsAny = append(selection.ColorsAny, typeSelection.ColorsAny...)
	selection.Supertypes = append(selection.Supertypes, typeSelection.Supertypes...)
	return true
}

func recognizeControllerResourceCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "have", "no", "cards", "in", "hand") {
		return ConditionClause{Predicate: ConditionPredicateControllerHandEmpty}, true
	}
	rest, ok := cutTokenPrefix(body, "you", "have")
	if ok {
		if count, tail, ok := parseLeadingCount(rest); ok && count.Comparison == ConditionComparisonAtLeast {
			switch {
			case tokenWordsEqual(tail, "cards", "in", "hand"):
				return ConditionClause{Predicate: ConditionPredicateControllerHandSizeAtLeast, Threshold: count.Value}, true
			case tokenWordsEqual(tail, "life"):
				return ConditionClause{Predicate: ConditionPredicateControllerLifeAtLeast, Threshold: count.Value}, true
			case tokenWordsEqual(tail, "opponents"):
				return ConditionClause{Predicate: ConditionPredicateOpponentCountAtLeast, Threshold: count.Value}, true
			}
		}
	}
	rest, ok = cutTokenPrefix(body, "a", "player", "has")
	if ok {
		if count, tail, ok := parseLeadingCount(rest); ok &&
			count.Comparison == ConditionComparisonAtMost &&
			tokenWordsEqual(tail, "life") {
			return ConditionClause{Predicate: ConditionPredicateAnyPlayerLifeAtMost, Threshold: count.Value}, true
		}
	}
	return ConditionClause{}, false
}

func recognizeGraveyardCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest := body
	if trimmed, ok := cutTokenPrefix(body, "there", "are"); ok {
		rest = trimmed
	}
	count, tail, ok := parseLeadingCount(rest)
	if !ok || count.Comparison != ConditionComparisonAtLeast {
		return ConditionClause{}, false
	}
	switch {
	case tokenWordsEqual(tail, "cards", "in", "your", "graveyard"),
		tokenWordsEqual(tail, "cards", "are", "in", "your", "graveyard"):
		return ConditionClause{Predicate: ConditionPredicateGraveyardCardCountAtLeast, Threshold: count.Value}, true
	case tokenWordsEqual(tail, "card", "types", "among", "cards", "in", "your", "graveyard"):
		return ConditionClause{Predicate: ConditionPredicateGraveyardCardTypeCountAtLeast, Threshold: count.Value}, true
	default:
		return ConditionClause{}, false
	}
}

func recognizeCounterPlacementCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "would", "put", "one", "or", "more", "counters", "on", "a", "permanent", "or", "player") {
		return ConditionClause{Predicate: ConditionPredicateControllerCounterPlacement}, true
	}
	rest, ok := cutTokenPrefix(body, "one", "or", "more")
	if !ok {
		return ConditionClause{}, false
	}
	tail, ok := stripTokenSuffix(rest, "counters", "would", "be", "put", "on", "a", "creature", "you", "control")
	if !ok || len(tail) == 0 {
		return ConditionClause{}, false
	}
	counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateCounterPlacementOnControlledCreature,
		Counter:   counterKind,
	}, true
}

func recognizeDamageSourceCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := stripTokenSuffix(body, "source", "you", "control", "would", "deal", "damage", "to", "a", "permanent", "or", "player")
	if !ok {
		return ConditionClause{}, false
	}
	var selection ConditionSelection
	if trimmed, ok := cutTokenPrefix(rest, "another"); ok {
		selection.ExcludeSource = true
		rest = trimmed
	} else if trimmed, ok := cutTokenPrefix(rest, "a"); ok {
		rest = trimmed
	} else {
		return ConditionClause{}, false
	}
	for _, token := range rest {
		color, ok := atoms.ColorAt(token.Span)
		if !ok {
			return ConditionClause{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, triggerColorFromAtom(color))
	}
	return ConditionClause{
		Predicate: ConditionPredicateDamageByControlledSource,
		Selection: selection,
	}, true
}

func recognizeTokenCreationCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "an", "effect", "would", "create", "one", "or", "more", "tokens", "under", "your", "control") {
		return ConditionClause{Predicate: ConditionPredicateTokenCreationUnderController}, true
	}
	return ConditionClause{}, false
}

func recognizeControlsCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	scope, rest, ok := cutControlScope(body)
	if !ok {
		return ConditionClause{}, false
	}
	determiner, ok := parseControlsDeterminer(rest)
	if !ok {
		return ConditionClause{}, false
	}
	count, exclude, tail := determiner.Count, determiner.Exclude, determiner.Rest
	if scope == ConditionControlScopeController &&
		count.Comparison == ConditionComparisonAtLeast &&
		tokenWordsEqual(tail, "creatures", "with", "different", "powers") {
		return ConditionClause{
			Predicate: ConditionPredicateCreaturePowerDiversityAtLeast,
			Threshold: count.Value,
		}, true
	}
	selection, ok := parseConditionSelection(tail, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	selection.ExcludeSource = selection.ExcludeSource || exclude
	return ConditionClause{
		Predicate:    ConditionPredicateControls,
		Scope:        scope,
		Comparison:   count.Comparison,
		CompareValue: count.Value,
		Selection:    selection,
	}, true
}

// recognizeTotalPowerCondition matches "<selection> you control have total
// power <n> or greater", a collective-power predicate (the "Formidable" ability
// word). The selected permanents the controller controls must have a combined
// power of at least the threshold. Only the controller scope is recognized.
func recognizeTotalPowerCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	haveIndex := tokenWordIndex(body, "have")
	if haveIndex <= 0 {
		return ConditionClause{}, false
	}
	rest, ok := cutTokenPrefix(body[haveIndex+1:], "total", "power")
	if !ok || len(rest) != 3 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok || !equalWord(rest[1], "or") || !equalWord(rest[2], "greater") {
		return ConditionClause{}, false
	}
	nounTokens, ok := stripTokenSuffix(body[:haveIndex], "you", "control")
	if !ok || len(nounTokens) == 0 {
		return ConditionClause{}, false
	}
	selection, ok := parseConditionSelection(nounTokens, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	selection.TotalPowerAtLeast = value
	selection.MatchTotalPowerAtLeast = true
	return ConditionClause{
		Predicate:  ConditionPredicateControls,
		Scope:      ConditionControlScopeController,
		Comparison: ConditionComparisonNone,
		Selection:  selection,
	}, true
}

func recognizeSourceDeathCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	for _, candidate := range []struct {
		suffix    []string
		predicate ConditionPredicateKind
	}{
		{[]string{"would", "die"}, ConditionPredicateSourceWouldDie},
		{[]string{"would", "be", "put", "into", "a", "graveyard", "from", "anywhere"}, ConditionPredicateSourceWouldGoToGraveyard},
	} {
		subject, ok := stripTokenSuffix(body, candidate.suffix...)
		if !ok || len(subject) == 0 {
			continue
		}
		subjectSpan := shared.SpanOf(subject)
		return ConditionClause{
			Predicate:      candidate.predicate,
			SubjectSpan:    subjectSpan,
			HasSubjectSpan: true,
			SubjectRefID:   atoms.ReferenceIDAt(subjectSpan),
		}, true
	}
	return ConditionClause{}, false
}

type conditionCount struct {
	Comparison ConditionComparison `json:",omitempty"`
	Value      int                 `json:",omitempty"`
}

// parseLeadingCount parses a leading "<n> or more|fewer|less" count, returning
// the typed comparison and the remaining tokens.
func parseLeadingCount(tokens []shared.Token) (conditionCount, []shared.Token, bool) {
	if len(tokens) < 3 {
		return conditionCount{}, nil, false
	}
	value, ok := conditionNumberValue(tokens[0])
	if !ok || !equalWord(tokens[1], "or") {
		return conditionCount{}, nil, false
	}
	switch {
	case equalWord(tokens[2], "more"):
		return conditionCount{Comparison: ConditionComparisonAtLeast, Value: value}, tokens[3:], true
	case equalWord(tokens[2], "fewer"), equalWord(tokens[2], "less"):
		return conditionCount{Comparison: ConditionComparisonAtMost, Value: value}, tokens[3:], true
	default:
		return conditionCount{}, nil, false
	}
}

// controlsDeterminer holds the parsed opening determiner of a controlled-permanent
// noun phrase.
type controlsDeterminer struct {
	Count   conditionCount `json:",omitzero"`
	Exclude bool           `json:",omitempty"`
	Rest    []shared.Token `json:"-"`
}

// parseControlsDeterminer parses the determiner that opens a controlled-permanent
// noun phrase: "a"/"an"/"another"/"no"/"<n> or more|fewer". It returns the count,
// whether the source is excluded, and the remaining noun tokens.
func parseControlsDeterminer(tokens []shared.Token) (controlsDeterminer, bool) {
	var rest []shared.Token
	count := conditionCount{Comparison: ConditionComparisonNone}
	exclude := false
	switch {
	case startsWithWord(tokens, "another"):
		exclude = true
		rest = tokens[1:]
	case startsWithWord(tokens, "a"), startsWithWord(tokens, "an"):
		rest = tokens[1:]
	case startsWithWord(tokens, "no"):
		count = conditionCount{Comparison: ConditionComparisonAtMost, Value: 0}
		rest = tokens[1:]
	default:
		parsed, tail, ok := parseLeadingCount(tokens)
		if !ok {
			return controlsDeterminer{}, false
		}
		count = parsed
		rest = tail
	}
	if trimmed, ok := cutTokenPrefix(rest, "other"); ok {
		exclude = true
		rest = trimmed
	}
	if len(rest) == 0 {
		return controlsDeterminer{}, false
	}
	return controlsDeterminer{Count: count, Exclude: exclude, Rest: rest}, true
}

// parseConditionSelection parses a permanent noun phrase into a typed selection,
// consuming card-type, subtype, color, and supertype atoms by span. It fails
// closed unless every token belongs to a recognized production.
func parseConditionSelection(tokens []shared.Token, atoms Atoms) (ConditionSelection, bool) {
	if len(tokens) == 0 {
		return ConditionSelection{}, false
	}
	var selection ConditionSelection
	// Trailing "with power <n> or greater" qualifier.
	if idx := tokenWordIndex(tokens, "with"); idx >= 0 {
		if !parseConditionPowerQualifier(tokens[idx+1:], &selection) {
			return ConditionSelection{}, false
		}
		tokens = tokens[:idx]
	}
	if len(tokens) == 0 {
		return ConditionSelection{}, false
	}
	// Leading tapped/untapped state.
	switch {
	case equalWord(tokens[0], "tapped"):
		selection.Tapped = ConditionTappedTrue
		tokens = tokens[1:]
	case equalWord(tokens[0], "untapped"):
		selection.Tapped = ConditionTappedFalse
		tokens = tokens[1:]
	default:
	}
	// Leading supertypes (basic/snow).
	for len(tokens) > 0 {
		supertype, ok := conditionSupertypeAtom(tokens[0].Span, atoms)
		if !ok {
			break
		}
		selection.Supertypes = append(selection.Supertypes, supertype)
		tokens = tokens[1:]
	}
	if len(tokens) == 0 {
		return selection, false
	}
	return parseConditionNoun(tokens, atoms, selection)
}

func parseConditionNoun(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if orIndex := tokenWordIndex(tokens, "or"); orIndex >= 0 {
		return parseConditionAlternativeNoun(tokens, orIndex, atoms, selection)
	}
	// Color-qualified "<colors> creature|permanent".
	if clause, ok := parseConditionColorQualified(tokens, atoms, selection); ok {
		return clause, true
	}
	// One or more distinct card-type words.
	cardTypes := make([]TriggerCardType, 0, len(tokens))
	allTypes := true
	for _, token := range tokens {
		cardType, ok := atoms.CardTypeAt(token.Span)
		if !ok {
			allTypes = false
			break
		}
		mapped := triggerCardTypeFromAtom(cardType)
		if slices.Contains(cardTypes, mapped) {
			return ConditionSelection{}, false
		}
		cardTypes = append(cardTypes, mapped)
	}
	if allTypes {
		selection.RequiredTypes = append(selection.RequiredTypes, cardTypes...)
		return selection, len(selection.RequiredTypes) > 0
	}
	// A bare permanent (no required type), e.g. "permanent" or "permanents".
	if tokenWordsEqual(tokens, "permanent") || tokenWordsEqual(tokens, "permanents") {
		return selection, true
	}
	// A bare token, e.g. "you control a token".
	if tokenWordsEqual(tokens, "token") || tokenWordsEqual(tokens, "tokens") {
		selection.TokenOnly = true
		return selection, true
	}
	// A subtype noun: creature, land, or "<name> planeswalker".
	return parseConditionSubtypeNoun(tokens, atoms, selection)
}

func parseConditionSubtypeNoun(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	span := shared.SpanOf(tokens)
	if subtype, ok := atoms.SubtypeAt(span); ok {
		selection.SubtypesAny = append(selection.SubtypesAny, subtype)
		return selection, true
	}
	if len(tokens) >= 2 && equalWord(tokens[len(tokens)-1], "planeswalker") {
		nameSpan := shared.SpanOf(tokens[:len(tokens)-1])
		if subtype, ok := conditionSubtypeAtom(nameSpan, atoms, CardTypePlaneswalker); ok {
			selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypePlaneswalker)
			selection.SubtypesAny = append(selection.SubtypesAny, subtype)
			return selection, true
		}
	}
	// A typed subtype noun "<name> creature", e.g. "a Griffin creature".
	if len(tokens) >= 2 &&
		(equalWord(tokens[len(tokens)-1], "creature") || equalWord(tokens[len(tokens)-1], "creatures")) {
		nameSpan := shared.SpanOf(tokens[:len(tokens)-1])
		if subtype, ok := conditionSubtypeAtom(nameSpan, atoms, CardTypeCreature); ok {
			selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeCreature)
			selection.SubtypesAny = append(selection.SubtypesAny, subtype)
			return selection, true
		}
	}
	return ConditionSelection{}, false
}

func parseConditionAlternativeNoun(tokens []shared.Token, orIndex int, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if clause, ok := parseConditionColorQualified(tokens, atoms, selection); ok {
		return clause, true
	}
	left := tokens[:orIndex]
	right := tokens[orIndex+1:]
	if trimmed, ok := cutTokenPrefix(right, "a"); ok {
		right = trimmed
	} else if trimmed, ok := cutTokenPrefix(right, "an"); ok {
		right = trimmed
	}
	leftSub, leftOK := conditionSubtypeAtom(shared.SpanOf(left), atoms, CardTypeLand)
	rightSub, rightOK := conditionSubtypeAtom(shared.SpanOf(right), atoms, CardTypeLand)
	if !leftOK || !rightOK {
		return ConditionSelection{}, false
	}
	selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeLand)
	selection.SubtypesAny = append(selection.SubtypesAny, leftSub, rightSub)
	return selection, true
}

// parseConditionColorQualified handles "<colors> creature(s)" and "<colors>
// permanent(s)", where colors are one or more color atoms joined by "or", or the
// "colorless"/"multicolored" qualifier.
func parseConditionColorQualified(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if len(tokens) < 2 {
		return ConditionSelection{}, false
	}
	last := tokens[len(tokens)-1]
	colorTokens := tokens[:len(tokens)-1]
	switch {
	case equalWord(last, "creature"), equalWord(last, "creatures"):
		selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeCreature)
	case equalWord(last, "permanent"), equalWord(last, "permanents"):
	default:
		return ConditionSelection{}, false
	}
	if tokenWordsEqual(colorTokens, "colorless") {
		selection.Colorless = true
		return selection, true
	}
	if tokenWordsEqual(colorTokens, "multicolored") {
		selection.Multicolored = true
		return selection, true
	}
	for _, token := range colorTokens {
		if equalWord(token, "or") {
			continue
		}
		color, ok := atoms.ColorAt(token.Span)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, triggerColorFromAtom(color))
	}
	if len(selection.ColorsAny) == 0 {
		return ConditionSelection{}, false
	}
	return selection, true
}

func parseConditionPowerQualifier(tokens []shared.Token, selection *ConditionSelection) bool {
	rest, ok := cutTokenPrefix(tokens, "power")
	if !ok || len(rest) != 3 {
		return false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok || !equalWord(rest[1], "or") || !equalWord(rest[2], "greater") {
		return false
	}
	selection.PowerAtLeast = value
	selection.MatchPowerAtLeast = true
	return true
}

func cutControlScope(tokens []shared.Token) (ConditionControlScope, []shared.Token, bool) {
	if rest, ok := cutTokenPrefix(tokens, "you", "control"); ok {
		return ConditionControlScopeController, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "an", "opponent", "controls"); ok {
		return ConditionControlScopeAnyOpponent, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "your", "opponents", "control"); ok {
		return ConditionControlScopeOpponents, rest, true
	}
	return ConditionControlScopeController, nil, false
}

func conditionCounterAtom(span shared.Span, atoms Atoms) (ConditionCounterKind, bool) {
	kind, _, ok := atoms.CounterIn(span)
	if !ok {
		return ConditionCounterNone, false
	}
	switch kind {
	case counter.PlusOnePlusOne:
		return ConditionCounterPlusOnePlusOne, true
	case counter.MinusOneMinusOne:
		return ConditionCounterMinusOneMinusOne, true
	default:
		return ConditionCounterNone, false
	}
}

func conditionSupertypeAtom(span shared.Span, atoms Atoms) (ConditionSupertype, bool) {
	supertype, ok := atoms.SupertypeAt(span)
	if !ok {
		return ConditionSupertypeUnknown, false
	}
	switch supertype {
	case SupertypeBasic:
		return ConditionSupertypeBasic, true
	case SupertypeSnow:
		return ConditionSupertypeSnow, true
	default:
		return ConditionSupertypeUnknown, false
	}
}

func conditionSubtypeAtom(span shared.Span, atoms Atoms, cardType CardType) (types.Sub, bool) {
	sub, ok := atoms.SubtypeAt(span)
	if !ok {
		return "", false
	}
	if !SubtypeMatchesCardType(sub, cardType) {
		return "", false
	}
	return sub, true
}

func conditionNumberValue(token shared.Token) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		return value, err == nil
	}
	if token.Kind == shared.Word {
		return CardinalWordValue(token.Text)
	}
	return 0, false
}

func triggerCardTypeFromAtom(cardType CardType) TriggerCardType {
	switch cardType {
	case CardTypeArtifact:
		return TriggerCardTypeArtifact
	case CardTypeBattle:
		return TriggerCardTypeBattle
	case CardTypeCreature:
		return TriggerCardTypeCreature
	case CardTypeEnchantment:
		return TriggerCardTypeEnchantment
	case CardTypeInstant:
		return TriggerCardTypeInstant
	case CardTypeLand:
		return TriggerCardTypeLand
	case CardTypePlaneswalker:
		return TriggerCardTypePlaneswalker
	case CardTypeSorcery:
		return TriggerCardTypeSorcery
	default:
		return TriggerCardTypeUnknown
	}
}

func triggerColorFromAtom(color Color) TriggerColor {
	switch color {
	case ColorWhite:
		return TriggerColorWhite
	case ColorBlue:
		return TriggerColorBlue
	case ColorBlack:
		return TriggerColorBlack
	case ColorRed:
		return TriggerColorRed
	case ColorGreen:
		return TriggerColorGreen
	default:
		return TriggerColorUnknown
	}
}

func conditionSelectionEmptyExceptType(selection ConditionSelection) bool {
	return len(selection.RequiredTypes) == 1 &&
		len(selection.Supertypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		!selection.Colorless &&
		!selection.ExcludeSource &&
		selection.Tapped == ConditionTappedAny &&
		!selection.MatchPowerAtLeast
}

func tokenSuffixWord(tokens []shared.Token, word string) bool {
	return len(tokens) > 0 && equalWord(tokens[len(tokens)-1], word)
}

func tokenWordIndex(tokens []shared.Token, word string) int {
	for i := range tokens {
		if equalWord(tokens[i], word) {
			return i
		}
	}
	return -1
}
