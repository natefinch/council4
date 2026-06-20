package parser

import (
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// StaticDeclarationKind identifies the static-declaration family the parser
// recognized for one composable clause.
type StaticDeclarationKind string

// Static declaration families recognized by the parser.
const (
	StaticDeclarationUnknown                      StaticDeclarationKind = ""
	StaticDeclarationContinuousPowerToughness     StaticDeclarationKind = "StaticDeclarationContinuousPowerToughness"
	StaticDeclarationContinuousBasePowerToughness StaticDeclarationKind = "StaticDeclarationContinuousBasePowerToughness"
	StaticDeclarationContinuousCharacteristic     StaticDeclarationKind = "StaticDeclarationContinuousCharacteristic"
	StaticDeclarationKeywordGrant                 StaticDeclarationKind = "StaticDeclarationKeywordGrant"
	StaticDeclarationRule                         StaticDeclarationKind = "StaticDeclarationRule"
	StaticDeclarationCostModifier                 StaticDeclarationKind = "StaticDeclarationCostModifier"
	StaticDeclarationCardAbilityGrant             StaticDeclarationKind = "StaticDeclarationCardAbilityGrant"
	StaticDeclarationPermanentAbilityGrant        StaticDeclarationKind = "StaticDeclarationPermanentAbilityGrant"
	StaticDeclarationControlGrant                 StaticDeclarationKind = "StaticDeclarationControlGrant"
	StaticDeclarationPlayerRule                   StaticDeclarationKind = "StaticDeclarationPlayerRule"
	StaticDeclarationLoseAbilitiesBecome          StaticDeclarationKind = "StaticDeclarationLoseAbilitiesBecome"
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
	StaticDeclarationSubjectController     StaticDeclarationSubjectKind = "StaticDeclarationSubjectController"
)

// StaticDeclarationPlayerRuleKind identifies the closed player-scoped rule a
// typed static declaration carries.
type StaticDeclarationPlayerRuleKind string

// Static declaration player rules recognized by the parser.
const (
	StaticDeclarationPlayerRuleUnknown           StaticDeclarationPlayerRuleKind = ""
	StaticDeclarationPlayerRuleNoMaximumHandSize StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleNoMaximumHandSize"
	StaticDeclarationPlayerRuleAttackTax         StaticDeclarationPlayerRuleKind = "StaticDeclarationPlayerRuleAttackTax"
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
	StaticDeclarationCostModifierSpellReduction   StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierSpellReduction"
	StaticDeclarationCostModifierSpellIncrease    StaticDeclarationCostModifierKind = "StaticDeclarationCostModifierSpellIncrease"
)

// StaticDeclarationSpellTypeKind identifies the closed spell-type filter a
// controller cast-cost modifier constrains.
type StaticDeclarationSpellTypeKind string

// Static declaration spell-type filters recognized by the parser.
const (
	StaticDeclarationSpellTypeAll              StaticDeclarationSpellTypeKind = ""
	StaticDeclarationSpellTypeArtifact         StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeArtifact"
	StaticDeclarationSpellTypeCreature         StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeCreature"
	StaticDeclarationSpellTypeEnchantment      StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeEnchantment"
	StaticDeclarationSpellTypeInstant          StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeInstant"
	StaticDeclarationSpellTypeSorcery          StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeSorcery"
	StaticDeclarationSpellTypeInstantOrSorcery StaticDeclarationSpellTypeKind = "StaticDeclarationSpellTypeInstantOrSorcery"
)

// StaticDeclarationSpellColorKind identifies the closed single-color filter a
// controller cast-cost modifier constrains. It is mutually exclusive with the
// spell-type filter: a declaration carries at most one of the two.
type StaticDeclarationSpellColorKind string

// Static declaration spell-color filters recognized by the parser.
const (
	StaticDeclarationSpellColorNone      StaticDeclarationSpellColorKind = ""
	StaticDeclarationSpellColorWhite     StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorWhite"
	StaticDeclarationSpellColorBlue      StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorBlue"
	StaticDeclarationSpellColorBlack     StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorBlack"
	StaticDeclarationSpellColorRed       StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorRed"
	StaticDeclarationSpellColorGreen     StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorGreen"
	StaticDeclarationSpellColorColorless StaticDeclarationSpellColorKind = "StaticDeclarationSpellColorColorless"
)

// StaticDeclarationSubject is a source-spanned typed affected group.
type StaticDeclarationSubject struct {
	Kind       StaticDeclarationSubjectKind    `json:",omitempty"`
	Span       shared.Span                     `json:"-"`
	Group      EffectStaticSubjectSyntax       `json:",omitzero"`
	CardFilter StaticDeclarationCardFilterKind `json:",omitempty"`
}

// StaticGrantedManaAbilitySyntax is one typed activated mana ability quoted by
// a static permanent-ability grant.
type StaticGrantedManaAbilitySyntax struct {
	Span     shared.Span `json:"-"`
	TapCost  bool        `json:",omitempty"`
	Amount   int         `json:",omitempty"`
	AnyColor bool        `json:",omitempty"`
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

	// Continuous base power/toughness (characteristic-setting) payload.
	BasePower     int  `json:",omitempty"`
	BaseToughness int  `json:",omitempty"`
	BasePTSet     bool `json:",omitempty"`

	// LoseAllAbilities marks a StaticDeclarationLoseAbilitiesBecome declaration
	// whose affected object loses all abilities ("loses all abilities"). For that
	// kind Colors, CardTypes, and Subtypes are SET (replacing the object's
	// existing colors, card types, and creature types) rather than added.
	LoseAllAbilities bool `json:",omitempty"`

	// Continuous characteristic addition payload: the colors, card types, and
	// subtypes a "<group> is/are ... in addition to ..." declaration grants. A
	// bare "<group> is/are <color>" with no "in addition" tail sets colors and
	// leaves ColorsAdd false; an explicit "in addition to its other colors" tail
	// sets ColorsAdd. Card types and subtypes are always additive.
	Colors    []Color     `json:"-"`
	CardTypes []CardType  `json:"-"`
	Subtypes  []types.Sub `json:"-"`
	ColorsAdd bool        `json:",omitempty"`

	// Keyword-grant and card-ability-grant payload: the spans of the granted
	// keyword atoms in source order.
	KeywordSpans []shared.Span `json:"-"`

	// Permanent-ability-grant payload.
	GrantedManaAbility *StaticGrantedManaAbilitySyntax `json:",omitempty"`

	// Rule payload.
	Rule StaticRuleSyntax `json:",omitzero"`

	// Cost-modifier payload.
	CostModifier        StaticDeclarationCostModifierKind `json:",omitempty"`
	CostReductionAmount int                               `json:",omitempty"`
	CostReplacement     string                            `json:",omitempty"`
	SpellType           StaticDeclarationSpellTypeKind    `json:",omitempty"`
	SpellColor          StaticDeclarationSpellColorKind   `json:",omitempty"`

	// Player-rule payload: the closed player-scoped rule this declaration grants
	// to the static ability's controller.
	PlayerRule       StaticDeclarationPlayerRuleKind `json:",omitempty"`
	AttackTaxGeneric int                             `json:",omitempty"`
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
		declarations := parseStaticDeclarations(body, ability.Quoted, ability.Atoms, ability.ConditionClauses)
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

func parseStaticDeclarations(tokens []shared.Token, quoted []Delimited, atoms Atoms, conditions []ConditionClause) []StaticDeclarationSyntax {
	if declaration, ok := parseStaticCostModifierDeclaration(tokens, atoms, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticSpellCostModifierDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticCardAbilityGrantDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticPermanentAbilityGrantDeclaration(tokens, quoted, conditions); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticControlGrantDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticPlayerRuleDeclaration(tokens); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declaration, ok := parseStaticLoseAbilitiesBecomeDeclaration(tokens, atoms); ok {
		return []StaticDeclarationSyntax{declaration}
	}
	if declarations, ok := parseStaticSubjectDeclarations(tokens, atoms, conditions); ok {
		return declarations
	}
	return nil
}

func parseStaticPermanentAbilityGrantDeclaration(
	tokens []shared.Token,
	quoted []Delimited,
	conditions []ConditionClause,
) (StaticDeclarationSyntax, bool) {
	if len(conditions) != 0 ||
		len(quoted) != 1 ||
		len(tokens) != 4 ||
		!staticWordsAt(tokens, 0, "lands", "you", "control", "have") {
		return StaticDeclarationSyntax{}, false
	}
	ability, ok := parseStaticGrantedManaAbility(quoted[0])
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	subjectSpan := shared.SpanOf(tokens[:3])
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPermanentAbilityGrant,
		Span:          shared.Span{Start: tokens[0].Span.Start, End: quoted[0].Span.End},
		OperationSpan: quoted[0].Span,
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectGroup,
			Span: subjectSpan,
			Group: EffectStaticSubjectSyntax{
				Kind: EffectStaticSubjectControlledLands,
				Span: subjectSpan,
			},
		},
		GrantedManaAbility: &ability,
	}, true
}

func parseStaticGrantedManaAbility(quoted Delimited) (StaticGrantedManaAbilitySyntax, bool) {
	tokens := quoted.Tokens
	if len(tokens) != 11 ||
		tokens[0].Kind != shared.Quote ||
		tokens[1].Kind != shared.Symbol ||
		tokens[1].Text != "{T}" ||
		tokens[2].Kind != shared.Colon ||
		!staticWordsAt(tokens, 3, "add", "one", "mana", "of", "any", "color") ||
		tokens[9].Kind != shared.Period ||
		tokens[10].Kind != shared.Quote {
		return StaticGrantedManaAbilitySyntax{}, false
	}
	return StaticGrantedManaAbilitySyntax{
		Span:     shared.SpanOf(tokens[1:10]),
		TapCost:  true,
		Amount:   1,
		AnyColor: true,
	}, true
}

// parseStaticControlGrantDeclaration recognizes the static source-tied control
// grant printed on control Auras: "You control enchanted creature." or "You
// control enchanted permanent." The affected group is the attached object; the
// new controller is the static ability's controller (you).
func parseStaticControlGrantDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 5 || tokens[4].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "control", "enchanted") {
		return StaticDeclarationSyntax{}, false
	}
	if !equalWord(tokens[3], "creature") && !equalWord(tokens[3], "permanent") {
		return StaticDeclarationSyntax{}, false
	}
	objectSpan := shared.SpanOf(tokens[2:4])
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationControlGrant,
		Span:          shared.SpanOf(tokens),
		OperationSpan: tokens[1].Span,
		Subject: StaticDeclarationSubject{
			Kind:  StaticDeclarationSubjectGroup,
			Span:  objectSpan,
			Group: EffectStaticSubjectSyntax{Kind: EffectStaticSubjectAttachedObject, Span: objectSpan},
		},
	}, true
}

type staticPlayerRuleParser func([]shared.Token) (StaticDeclarationSyntax, bool)

var staticPlayerRuleParsers = []staticPlayerRuleParser{
	parseStaticNoMaximumHandSizeDeclaration,
	parseStaticAttackTaxDeclaration,
}

func parseStaticPlayerRuleDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	for _, parse := range staticPlayerRuleParsers {
		if declaration, ok := parse(tokens); ok {
			return declaration, true
		}
	}
	return StaticDeclarationSyntax{}, false
}

// parseStaticNoMaximumHandSizeDeclaration recognizes the exact controller-scoped
// no-maximum-hand-size rule.
func parseStaticNoMaximumHandSizeDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 7 || tokens[6].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "you", "have", "no", "maximum", "hand", "size") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:6]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[0].Span,
		},
		PlayerRule: StaticDeclarationPlayerRuleNoMaximumHandSize,
	}, true
}

// parseStaticAttackTaxDeclaration recognizes the exact fixed-generic attack tax
// "Creatures can't attack you unless their controller pays {N} for each creature
// they control that's attacking you." The affected player is the static ability's
// controller; the cost is paid independently for each declared attacker.
func parseStaticAttackTaxDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 18 ||
		tokens[8].Kind != shared.Symbol ||
		tokens[17].Kind != shared.Period ||
		!staticWordsAt(tokens, 0, "creatures", "can't", "attack", "you", "unless", "their", "controller", "pays") ||
		!staticWordsAt(tokens, 9, "for", "each", "creature", "they", "control", "that's", "attacking", "you") {
		return StaticDeclarationSyntax{}, false
	}
	amount, ok := staticGenericSymbolValue(tokens[8].Text)
	if !ok || amount <= 0 {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationPlayerRule,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens[1:17]),
		Subject: StaticDeclarationSubject{
			Kind: StaticDeclarationSubjectController,
			Span: tokens[3].Span,
		},
		PlayerRule:       StaticDeclarationPlayerRuleAttackTax,
		AttackTaxGeneric: amount,
	}, true
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

// parseStaticSpellCostModifierDeclaration recognizes the static cast-cost
// modifier "[<filter>] spells you cast cost {N} less/more to cast." where the
// optional leading filter constrains the affected spells to a single card type,
// to instants and sorceries together, or to a single color (one of the five
// colors or colorless). The affected group is always the static ability's
// controller's spells. A type filter and a color filter are mutually exclusive.
func parseStaticSpellCostModifierDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	spellColor := staticSpellColorFilter(tokens)
	spellType := StaticDeclarationSpellTypeAll
	var rest []shared.Token
	if spellColor != StaticDeclarationSpellColorNone {
		rest = tokens[1:]
	} else {
		var ok bool
		spellType, rest, ok = staticSpellTypeFilter(tokens)
		if !ok {
			return StaticDeclarationSyntax{}, false
		}
	}
	if len(rest) != 9 ||
		!staticWordsAt(rest, 0, "spells", "you", "cast", "cost") ||
		rest[4].Kind != shared.Symbol ||
		!staticWordsAt(rest, 6, "to", "cast") {
		return StaticDeclarationSyntax{}, false
	}
	amount, ok := staticGenericSymbolValue(rest[4].Text)
	if !ok || amount <= 0 {
		return StaticDeclarationSyntax{}, false
	}
	var kind StaticDeclarationCostModifierKind
	switch {
	case equalWord(rest[5], "less"):
		kind = StaticDeclarationCostModifierSpellReduction
	case equalWord(rest[5], "more"):
		kind = StaticDeclarationCostModifierSpellIncrease
	default:
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCostModifier,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       shared.SpanOf(rest[3:8]),
		CostModifier:        kind,
		CostReductionAmount: amount,
		SpellType:           spellType,
		SpellColor:          spellColor,
	}, true
}

// staticSpellColorFilter recognizes a leading single-color filter word in a
// "<color> spells you cast cost ..." declaration ("White", "Blue", "Black",
// "Red", "Green", or "Colorless"). It returns the closed color filter, or
// StaticDeclarationSpellColorNone when the first token is not a recognized color
// word immediately followed by "spells". The color filter is mutually exclusive
// with the spell-type filter.
func staticSpellColorFilter(tokens []shared.Token) StaticDeclarationSpellColorKind {
	if len(tokens) < 2 || !equalWord(tokens[1], "spells") {
		return StaticDeclarationSpellColorNone
	}
	switch {
	case equalWord(tokens[0], "white"):
		return StaticDeclarationSpellColorWhite
	case equalWord(tokens[0], "blue"):
		return StaticDeclarationSpellColorBlue
	case equalWord(tokens[0], "black"):
		return StaticDeclarationSpellColorBlack
	case equalWord(tokens[0], "red"):
		return StaticDeclarationSpellColorRed
	case equalWord(tokens[0], "green"):
		return StaticDeclarationSpellColorGreen
	case equalWord(tokens[0], "colorless"):
		return StaticDeclarationSpellColorColorless
	default:
		return StaticDeclarationSpellColorNone
	}
}

// staticSpellTypeFilter strips an optional leading spell-type filter from a
// "spells you cast cost ..." declaration and returns the closed filter kind with
// the remaining tokens. It returns false when a leading word is present that is
// not a recognized single-type or instant-and-sorcery filter.
func staticSpellTypeFilter(tokens []shared.Token) (StaticDeclarationSpellTypeKind, []shared.Token, bool) {
	if len(tokens) == 0 {
		return StaticDeclarationSpellTypeAll, nil, false
	}
	if equalWord(tokens[0], "spells") {
		return StaticDeclarationSpellTypeAll, tokens, true
	}
	if len(tokens) >= 4 &&
		equalWord(tokens[0], "instant") &&
		equalWord(tokens[1], "and") &&
		equalWord(tokens[2], "sorcery") &&
		equalWord(tokens[3], "spells") {
		return StaticDeclarationSpellTypeInstantOrSorcery, tokens[3:], true
	}
	if len(tokens) < 2 || !equalWord(tokens[1], "spells") {
		return StaticDeclarationSpellTypeAll, nil, false
	}
	switch {
	case equalWord(tokens[0], "artifact"):
		return StaticDeclarationSpellTypeArtifact, tokens[1:], true
	case equalWord(tokens[0], "creature"):
		return StaticDeclarationSpellTypeCreature, tokens[1:], true
	case equalWord(tokens[0], "enchantment"):
		return StaticDeclarationSpellTypeEnchantment, tokens[1:], true
	case equalWord(tokens[0], "instant"):
		return StaticDeclarationSpellTypeInstant, tokens[1:], true
	case equalWord(tokens[0], "sorcery"):
		return StaticDeclarationSpellTypeSorcery, tokens[1:], true
	default:
		return StaticDeclarationSpellTypeAll, nil, false
	}
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
	if !ok || symbol == "" || (len(symbol) > 1 && symbol[0] == '0') {
		return 0, false
	}
	for i := range symbol {
		if symbol[i] < '0' || symbol[i] > '9' {
			return 0, false
		}
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
