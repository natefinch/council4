package parser

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// StaticDeclarationKind identifies the static-declaration family the parser
// recognized for one composable clause.
type StaticDeclarationKind string

// Static declaration families recognized by the parser.
const (
	StaticDeclarationUnknown                  StaticDeclarationKind = ""
	StaticDeclarationContinuousPowerToughness StaticDeclarationKind = "StaticDeclarationContinuousPowerToughness"
	StaticDeclarationKeywordGrant             StaticDeclarationKind = "StaticDeclarationKeywordGrant"
	StaticDeclarationRule                     StaticDeclarationKind = "StaticDeclarationRule"
	StaticDeclarationCostModifier             StaticDeclarationKind = "StaticDeclarationCostModifier"
	StaticDeclarationCardAbilityGrant         StaticDeclarationKind = "StaticDeclarationCardAbilityGrant"
)

// StaticDeclarationSubjectKind identifies the affected group named by a typed
// static declaration. Group subjects carry their typed effect-subject value.
type StaticDeclarationSubjectKind string

// Static declaration subjects recognized by the parser.
const (
	StaticDeclarationSubjectUnknown        StaticDeclarationSubjectKind = ""
	StaticDeclarationSubjectSourceCreature StaticDeclarationSubjectKind = "StaticDeclarationSubjectSourceCreature"
	StaticDeclarationSubjectSourceSpell    StaticDeclarationSubjectKind = "StaticDeclarationSubjectSourceSpell"
	StaticDeclarationSubjectSourceNamed    StaticDeclarationSubjectKind = "StaticDeclarationSubjectSourceNamed"
	StaticDeclarationSubjectGroup          StaticDeclarationSubjectKind = "StaticDeclarationSubjectGroup"
	StaticDeclarationSubjectControllerHand StaticDeclarationSubjectKind = "StaticDeclarationSubjectControllerHand"
)

// StaticDeclarationCardFilterKind identifies the closed card filter that a
// controller-hand subject constrains.
type StaticDeclarationCardFilterKind string

// Static declaration card filters recognized by the parser.
const (
	StaticDeclarationCardFilterNone     StaticDeclarationCardFilterKind = ""
	StaticDeclarationCardFilterLand     StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterLand"
	StaticDeclarationCardFilterCreature StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterCreature"
	StaticDeclarationCardFilterHistoric StaticDeclarationCardFilterKind = "StaticDeclarationCardFilterHistoric"
)

// StaticDeclarationCostModifierKind identifies the closed cost-modifier shape a
// typed static declaration carries.
type StaticDeclarationCostModifierKind string

// Static declaration cost-modifier shapes recognized by the parser.
const (
	StaticDeclarationCostModifierUnknown          StaticDeclarationCostModifierKind = ""
	StaticDeclarationCostModifierAbilityReduction StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierAbilityReduction"
	StaticDeclarationCostModifierReplaceCost      StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierReplaceCost"
	StaticDeclarationCostModifierReplaceFirstCost StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierReplaceFirstCost"
)

// StaticDeclarationSubject is a source-spanned typed affected group.
type StaticDeclarationSubject struct {
	Kind       StaticDeclarationSubjectKind    `json:",omitempty"`
	Span       shared.Span                     `json:"-"`
	Group      EffectStaticSubjectSyntax       `json:",omitzero"`
	CardFilter StaticDeclarationCardFilterKind `json:",omitempty"`
}

// StaticDeclarationSyntax is one composable typed static declaration. The
// compiler maps these onto its semantic vocabulary mechanically; it inspects no
// Oracle source text to derive meaning.
type StaticDeclarationSyntax struct {
	Kind          StaticDeclarationKind    `json:",omitempty"`
	Span          shared.Span              `json:"-"`
	OperationSpan shared.Span              `json:"-"`
	Subject       StaticDeclarationSubject `json:",omitzero"`

	// HasCondition records whether a single supported-shaped condition clause
	// applies to this declaration; ConditionSpan links to that clause.
	HasCondition  bool        `json:",omitempty"`
	ConditionSpan shared.Span `json:"-"`

	// Continuous power/toughness payload.
	PowerDelta     SignedAmountSyntax `json:",omitzero"`
	ToughnessDelta SignedAmountSyntax `json:",omitzero"`
	Dynamic        bool               `json:",omitempty"`

	// Keyword-grant and card-ability-grant payload: the spans of the granted
	// keyword atoms in source order.
	KeywordSpans []shared.Span `json:"-"`

	// Rule payload.
	Rule StaticRuleSyntax `json:",omitzero"`

	// Cost-modifier payload.
	CostModifier        StaticDeclarationCostModifierKind `json:",omitempty"`
	CostReductionAmount int                               `json:",omitempty"`
	CostReplacement     string                            `json:",omitempty"`
}

func emitStaticDeclarations(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if ability.Modal != nil {
			continue
		}
		body := staticDeclarationBodyTokens(ability)
		if len(body) == 0 {
			continue
		}
		declarations := parseStaticDeclarations(body, ability.Atoms, ability.ConditionClauses)
		if len(declarations) > 0 {
			ability.StaticDeclarations = declarations
		}
	}
}

// staticDeclarationBodyTokens returns the ability's semantic tokens with reminder
// and quoted text removed, and any ability-word label and its em dash dropped.
func staticDeclarationBodyTokens(ability *Ability) []shared.Token {
	tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
	if ability.AbilityWord == nil {
		return tokens
	}
	for i := range tokens {
		if tokens[i].Kind == shared.EmDash {
			return tokens[i+1:]
		}
	}
	return tokens
}

func parseStaticDeclarations(tokens []shared.Token, atoms Atoms, conditions []ConditionClause) []StaticDeclarationSyntax {
	if declaration, ok := parseStaticCostModifierDeclaration(tokens, atoms, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCardAbilityGrantDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declarations, ok := parseStaticSubjectDeclarations(tokens, atoms, conditions); ok {
		return declarations
	}
	return nil
}

// staticDeclarationCondition returns the single condition clause that lies within
// the declaration body, if exactly one is present.
func staticDeclarationCondition(tokens []shared.Token, conditions []ConditionClause) (ConditionClause, bool) {
	body := shared.SpanOf(tokens)
	matched := -1
	for i := range conditions {
		if spanCovers(body, conditions[i].Span) {
			if matched >= 0 {
				return ConditionClause{}, false
			}
			matched = i
		}
	}
	if matched < 0 {
		return ConditionClause{}, false
	}
	return conditions[matched], true
}

// tokensOutsideCondition removes a condition clause's tokens from the body and
// drops a comma left dangling by a leading condition.
func tokensOutsideCondition(tokens []shared.Token, span shared.Span) []shared.Token {
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if !spanCovers(span, token.Span) {
			result = append(result, token)
		}
	}
	if len(result) > 0 && result[0].Kind == shared.Comma {
		result = result[1:]
	}
	return result
}

func staticOperationTokens(tokens []shared.Token, conditions []ConditionClause) ([]shared.Token, ConditionClause, bool) {
	condition, ok := staticDeclarationCondition(tokens, conditions)
	if !ok {
		return tokens, ConditionClause{}, false
	}
	return tokensOutsideCondition(tokens, condition.Span), condition, true
}

func parseStaticCostModifierDeclaration(
	tokens []shared.Token,
	atoms Atoms,
	conditions []ConditionClause,
) (StaticDeclarationSyntax, bool) {
	span := shared.SpanOf(tokens)
	opTokens, condition, hasCondition := staticOperationTokens(tokens, conditions)
	if len(opTokens) == 0 || opTokens[len(opTokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	keyword, ok := staticSoleBareCyclingKeyword(opTokens, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	declaration := StaticDeclarationSyntax{
		Kind: StaticDeclarationCostModifier,
		Span: span,
	}
	if hasCondition {
		declaration.HasCondition = true
		declaration.ConditionSpan = condition.Span
	}
	if reduction, ok := parseStaticAbilityReduction(opTokens, keyword); ok {
		declaration.CostModifier = StaticDeclarationCostModifierAbilityReduction
		declaration.CostReductionAmount = reduction
		declaration.OperationSpan = keyword.Span
		return declaration, true
	}
	if replacement, ok := parseStaticReplaceCyclingCost(opTokens, keyword); ok {
		declaration.CostModifier = StaticDeclarationCostModifierReplaceCost
		declaration.CostReplacement = replacement
		declaration.OperationSpan = keyword.Span
		return declaration, true
	}
	if replacement, ok := parseStaticReplaceFirstCyclingCost(opTokens, keyword); ok {
		declaration.CostModifier = StaticDeclarationCostModifierReplaceFirstCost
		declaration.CostReplacement = replacement
		declaration.OperationSpan = keyword.Span
		return declaration, true
	}
	return StaticDeclarationSyntax{}, false
}

// parseStaticAbilityReduction recognizes "Cycling abilities you activate cost up
// to {N} less to activate." and returns the generic reduction N.
func parseStaticAbilityReduction(tokens []shared.Token, keyword Keyword) (int, bool) {
	if len(tokens) != 12 ||
		keyword.NameSpan.Start.Offset != tokens[0].Span.Start.Offset ||
		!staticWordsAt(tokens, 1, "abilities", "you", "activate", "cost", "up", "to") ||
		tokens[7].Kind != shared.Symbol ||
		!staticWordsAt(tokens, 8, "less", "to", "activate") {
		return 0, false
	}
	return staticGenericSymbolValue(tokens[7].Text)
}

// parseStaticReplaceCyclingCost recognizes "you may pay {N} rather than pay
// cycling costs." and returns the replacement cost text.
func parseStaticReplaceCyclingCost(tokens []shared.Token, keyword Keyword) (string, bool) {
	if len(tokens) != 10 ||
		!staticWordsAt(tokens, 0, "you", "may", "pay") ||
		tokens[3].Kind != shared.Symbol ||
		!staticWordsAt(tokens, 4, "rather", "than", "pay") ||
		keyword.NameSpan.Start.Offset != tokens[7].Span.Start.Offset ||
		!staticWordsAt(tokens, 8, "costs") {
		return "", false
	}
	return staticReplacementCost(tokens[3].Text)
}

// parseStaticReplaceFirstCyclingCost recognizes "You may pay {N} rather than pay
// the cycling cost of the first card you cycle each turn" and returns the
// replacement cost text.
func parseStaticReplaceFirstCyclingCost(tokens []shared.Token, keyword Keyword) (string, bool) {
	if len(tokens) != 19 ||
		!staticWordsAt(tokens, 0, "you", "may", "pay") ||
		tokens[3].Kind != shared.Symbol ||
		!staticWordsAt(tokens, 4, "rather", "than", "pay", "the") ||
		keyword.NameSpan.Start.Offset != tokens[8].Span.Start.Offset ||
		!staticWordsAt(tokens, 9, "cost", "of", "the", "first", "card", "you", "cycle", "each", "turn") {
		return "", false
	}
	return staticReplacementCost(tokens[3].Text)
}

func parseStaticCardAbilityGrantDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 9 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "each") {
		return StaticDeclarationSyntax{}, false
	}
	filter := staticHandCardFilter(tokens[1])
	if filter == StaticDeclarationCardFilterNone {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 2, "card", "in", "your", "hand", "has") {
		return StaticDeclarationSyntax{}, false
	}
	keyword, width, ok := staticKeywordAt(tokens, 7, len(tokens)-1, atoms)
	if !ok || keyword.Kind != KeywordCycling ||
		keyword.Parameter.Kind != KeywordParameterManaCost || 7+width != len(tokens)-1 {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationCardAbilityGrant,
		Span:          shared.SpanOf(tokens),
		OperationSpan: keyword.Span,
		Subject: StaticDeclarationSubject{
			Kind:       StaticDeclarationSubjectControllerHand,
			Span:       shared.SpanOf(tokens[:6]),
			CardFilter: filter,
		},
		KeywordSpans: []shared.Span{keyword.Span},
	}, true
}

func staticHandCardFilter(token shared.Token) StaticDeclarationCardFilterKind {
	switch {
	case equalWord(token, "land"):
		return StaticDeclarationCardFilterLand
	case equalWord(token, "creature"):
		return StaticDeclarationCardFilterCreature
	case equalWord(token, "historic"):
		return StaticDeclarationCardFilterHistoric
	default:
		return StaticDeclarationCardFilterNone
	}
}

// staticSoleBareCyclingKeyword returns the single cycling keyword atom in the
// body when it is the only keyword and carries no parameter.
func staticSoleBareCyclingKeyword(tokens []shared.Token, atoms Atoms) (Keyword, bool) {
	keywords := atoms.KeywordsWithin(tokens)
	if len(keywords) != 1 ||
		keywords[0].Kind != KeywordCycling ||
		keywords[0].Parameter.Kind != KeywordParameterNone {
		return Keyword{}, false
	}
	return keywords[0], true
}

// staticGenericSymbolValue returns the generic value of a single {N} symbol.
func staticGenericSymbolValue(text string) (int, bool) {
	symbol, ok := staticTrimSymbol(text)
	if !ok {
		return 0, false
	}
	value, err := strconv.Atoi(symbol)
	if err != nil {
		return 0, false
	}
	return value, true
}

// staticReplacementCost returns the canonical mana cost text for a single {N}
// generic symbol, where {0} renders as the empty string.
func staticReplacementCost(text string) (string, bool) {
	value, ok := staticGenericSymbolValue(text)
	if !ok {
		return "", false
	}
	if value == 0 {
		return "", true
	}
	return text, true
}

func staticTrimSymbol(text string) (string, bool) {
	symbol, ok := strings.CutPrefix(text, "{")
	if !ok {
		return "", false
	}
	return strings.CutSuffix(symbol, "}")
}
