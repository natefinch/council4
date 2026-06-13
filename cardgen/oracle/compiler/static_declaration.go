package compiler

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// StaticDeclarationKind identifies a declaration category that never resolves.
type StaticDeclarationKind uint8

// Static declaration categories.
const (
	StaticDeclarationUnknown StaticDeclarationKind = iota
	StaticDeclarationContinuous
	StaticDeclarationRule
	StaticDeclarationCostModifier
	StaticDeclarationCardAbilityGrant
)

// StaticDeclarationBlocker identifies exact static wording whose declaration
// category is understood but whose semantic vocabulary is not yet representable.
type StaticDeclarationBlocker uint8

// Exact static declaration blockers.
const (
	StaticDeclarationBlockerNone StaticDeclarationBlocker = iota
	StaticDeclarationBlockerHistoricCardSelection
	StaticDeclarationBlockerCondition
	StaticDeclarationBlockerDuration
	StaticDeclarationBlockerGroup
	StaticDeclarationBlockerOperation
	StaticDeclarationBlockerShell
)

// StaticContinuousLayer identifies a semantic continuous-effect layer.
type StaticContinuousLayer uint8

// Static continuous-effect layers currently recognized by Card Generation.
const (
	StaticLayerUnknown StaticContinuousLayer = iota
	StaticLayerAbility
	StaticLayerPowerToughnessModify
)

// StaticContinuousOperation identifies a characteristic operation.
type StaticContinuousOperation uint8

// Static continuous-effect operations.
const (
	StaticContinuousUnknown StaticContinuousOperation = iota
	StaticContinuousModifyPowerToughness
	StaticContinuousGrantKeywords
)

// StaticRuleKind identifies a non-layer rules declaration.
type StaticRuleKind uint8

// StaticRuleDomain identifies the rules action constrained by a declaration.
type StaticRuleDomain uint8

// Static rule domains. Operations are added only when the runtime can represent
// them, while the closed domains keep recognition independent of wording.
const (
	StaticRuleDomainUnknown StaticRuleDomain = iota
	StaticRuleDomainAttack
	StaticRuleDomainBlock
	StaticRuleDomainCast
	StaticRuleDomainActivate
	StaticRuleDomainTarget
	StaticRuleDomainCountering
)

// Static rule declarations currently recognized by Card Generation.
const (
	StaticRuleUnknown StaticRuleKind = iota
	StaticRuleCantBlock
	StaticRuleCantBeBlocked
	StaticRuleMustAttack
	StaticRuleCantBeCountered
)

// StaticZone identifies where a static declaration functions.
type StaticZone uint8

// Static declaration zones.
const (
	StaticZoneBattlefield StaticZone = iota
	StaticZoneStack
	StaticZoneHand
)

// StaticGroupDomain identifies the closed candidate domain of an affected group.
type StaticGroupDomain uint8

// Static affected-group domains.
const (
	StaticGroupUnknown StaticGroupDomain = iota
	StaticGroupSource
	StaticGroupBattlefield
	StaticGroupAttachedObject
	StaticGroupSourceControllerPermanents
	StaticGroupControllerHandCards
)

// StaticCardType identifies card types used by a static Selection.
type StaticCardType uint8

// Static Selection card types.
const (
	StaticCardTypeUnknown StaticCardType = iota
	StaticCardTypeArtifact
	StaticCardTypeCreature
	StaticCardTypeLand
)

// StaticSelection is source-independent semantic data describing WHAT objects
// in a static declaration's group match.
type StaticSelection struct {
	RequiredTypes []StaticCardType
	SubtypesAny   []types.Sub
	Controller    ControllerKind
	TokenOnly     bool
}

// StaticGroupReference describes WHERE a static declaration finds objects and
// carries the Selection that describes WHAT matches there.
type StaticGroupReference struct {
	Span          shared.Span
	Domain        StaticGroupDomain
	Selection     StaticSelection
	ExcludeSource bool
}

// StaticContinuousDeclaration is one layer-preserving characteristic change.
type StaticContinuousDeclaration struct {
	Layer          StaticContinuousLayer
	Operation      StaticContinuousOperation
	PowerDelta     CompiledSignedAmount
	ToughnessDelta CompiledSignedAmount
	DynamicAmount  CompiledAmount
	Keywords       []CompiledKeyword
}

// StaticRuleDeclaration is one prohibition, requirement, or permission.
type StaticRuleDeclaration struct {
	Domain StaticRuleDomain
	Kind   StaticRuleKind
	Zone   StaticZone
}

// StaticCostModifierKind identifies which semantic cost category is modified.
type StaticCostModifierKind uint8

// Static cost modifier kinds.
const (
	StaticCostModifierUnknown StaticCostModifierKind = iota
	StaticCostModifierAbility
)

// StaticCostModifierDeclaration is a semantic cost change.
type StaticCostModifierDeclaration struct {
	Kind               StaticCostModifierKind
	AbilityKeyword     parser.KeywordKind
	GenericReduction   int
	SetManaCost        string
	ReplaceManaCost    bool
	FirstCycleEachTurn bool
}

// StaticCardAbilityGrantDeclaration grants a keyword ability to cards in a
// non-battlefield group.
type StaticCardAbilityGrantDeclaration struct {
	Keyword CompiledKeyword
	Text    string
}

// StaticDeclaration is source-spanned semantic data attached directly to a
// static ability. It is not Instruction content and never resolves.
type StaticDeclaration struct {
	Kind          StaticDeclarationKind
	Span          shared.Span
	OperationSpan shared.Span
	Group         StaticGroupReference
	Condition     *CompiledCondition

	// Exactly one variant payload matching Kind is non-nil.
	Continuous *StaticContinuousDeclaration
	Rule       *StaticRuleDeclaration
	Cost       *StaticCostModifierDeclaration
	CardGrant  *StaticCardAbilityGrantDeclaration
}

// CompiledStaticSemantics contains declarations recognized for a static
// ability, or the exact reason a declaration-shaped ability cannot be lowered.
type CompiledStaticSemantics struct {
	Declarations []StaticDeclaration
	Blocker      StaticDeclarationBlocker
}

func recognizeStaticDeclarations(compiled *CompiledAbility, syntax parser.Ability) {
	if compiled.Kind != AbilityStatic {
		return
	}
	if declarations, ok := recognizeTypedStaticRuleDeclarations(*compiled, syntax); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeMixedSourceStaticDeclarations(*compiled, syntax); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticPowerToughnessDeclarations(*compiled, syntax); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticKeywordGrantDeclarations(*compiled, syntax); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declaration, ok := recognizeStaticCostModifierDeclaration(*compiled); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticCardAbilityGrantDeclaration(*compiled, syntax); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if isHistoricCardAbilityGrant(*compiled, syntax) {
		compiled.Static = &CompiledStaticSemantics{Blocker: StaticDeclarationBlockerHistoricCardSelection}
		return
	}
	if blocker := classifyStaticDeclarationBlocker(*compiled); blocker != StaticDeclarationBlockerNone {
		compiled.Static = &CompiledStaticSemantics{Blocker: blocker}
	}
}

func classifyStaticDeclarationBlocker(ability CompiledAbility) StaticDeclarationBlocker {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 {
		return StaticDeclarationBlockerNone
	}
	if len(ability.Content.Effects) != 1 {
		return StaticDeclarationBlockerNone
	}
	effect := ability.Content.Effects[0]
	rule := staticRuleForEffect(effect.Kind) != StaticRuleUnknown
	if effect.Kind != EffectModifyPT && effect.Kind != EffectGrantKeyword && !rule {
		return StaticDeclarationBlockerNone
	}
	if effect.Duration != DurationNone {
		return StaticDeclarationBlockerDuration
	}
	if len(ability.Content.Conditions) > 1 || (rule && len(ability.Content.Conditions) != 0) {
		return StaticDeclarationBlockerCondition
	}
	if len(ability.Content.Conditions) == 1 &&
		ability.Content.Conditions[0].Predicate == ConditionPredicateUnsupported {
		return StaticDeclarationBlockerCondition
	}
	if rule {
		if len(ability.Content.References) != 1 ||
			ability.Content.References[0].Binding != ReferenceBindingSource {
			return StaticDeclarationBlockerGroup
		}
		return StaticDeclarationBlockerOperation
	}
	if effect.StaticSubject == StaticSubjectNone {
		if len(ability.Content.References) != 1 ||
			ability.Content.References[0].Binding != ReferenceBindingSource {
			return StaticDeclarationBlockerGroup
		}
	}
	if ability.AbilityWord != "" && !recognizedStaticAbilityWord(ability.AbilityWord) {
		return StaticDeclarationBlockerShell
	}
	return StaticDeclarationBlockerOperation
}

func recognizedStaticAbilityWord(word string) bool {
	switch word {
	case "",
		"Coven",
		"Delirium",
		"Domain",
		"Ferocious",
		"Hellbent",
		"Metalcraft",
		"Threshold":
		return true
	default:
		return false
	}
}

func recognizeTypedStaticRuleDeclarations(ability CompiledAbility, syntax parser.Ability) ([]StaticDeclaration, bool) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" ||
		len(syntax.Sentences) != 1 ||
		syntax.Sentences[0].StaticRule == nil ||
		len(syntax.Reminders) != 0 ||
		len(syntax.Quoted) != 0 {
		return nil, false
	}
	node := syntax.Sentences[0].StaticRule
	rule, zone, ok := semanticStaticRuleForSyntax(*node)
	if !ok {
		return nil, false
	}
	if len(ability.Content.Effects) != 1 ||
		staticRuleForEffect(ability.Content.Effects[0].Kind) != rule ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != ReferenceBindingSource {
		return nil, false
	}
	return []StaticDeclaration{staticRuleDeclaration(node.Span, node.Subject.Span, node.Operation.Span, rule, zone, nil)}, true
}

func semanticStaticRuleForSyntax(rule parser.StaticRuleSyntax) (StaticRuleKind, StaticZone, bool) {
	if rule.Subject.Kind == parser.StaticRuleSubjectSourceCreature &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationBlock &&
		len(rule.Qualifiers) == 0 {
		switch rule.Operation.Voice {
		case parser.StaticRuleVoiceActive:
			return StaticRuleCantBlock, StaticZoneBattlefield, true
		case parser.StaticRuleVoicePassive:
			return StaticRuleCantBeBlocked, StaticZoneBattlefield, true
		default:
			return StaticRuleUnknown, StaticZoneBattlefield, false
		}
	}
	if rule.Subject.Kind == parser.StaticRuleSubjectSourceCreature &&
		rule.Constraint.Kind == parser.StaticRuleConstraintRequirement &&
		rule.Operation.Kind == parser.StaticRuleOperationAttack &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierEachCombat, parser.StaticRuleQualifierIfAble) {
		return StaticRuleMustAttack, StaticZoneBattlefield, true
	}
	if rule.Subject.Kind == parser.StaticRuleSubjectSourceSpell &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationCounter &&
		rule.Operation.Voice == parser.StaticRuleVoicePassive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleCantBeCountered, StaticZoneStack, true
	}
	return StaticRuleUnknown, StaticZoneBattlefield, false
}

func staticRuleForEffect(kind EffectKind) StaticRuleKind {
	switch kind {
	case EffectCantBlock:
		return StaticRuleCantBlock
	case EffectCantBeBlocked:
		return StaticRuleCantBeBlocked
	case EffectMustAttack:
		return StaticRuleMustAttack
	case EffectCantBeCountered:
		return StaticRuleCantBeCountered
	default:
		return StaticRuleUnknown
	}
}

func staticRuleDeclaration(
	span, subjectSpan, operationSpan shared.Span,
	rule StaticRuleKind,
	zone StaticZone,
	condition *CompiledCondition,
) StaticDeclaration {
	return StaticDeclaration{
		Kind:          StaticDeclarationRule,
		Span:          span,
		OperationSpan: operationSpan,
		Group: StaticGroupReference{
			Span:   subjectSpan,
			Domain: StaticGroupSource,
		},
		Condition: condition,
		Rule: &StaticRuleDeclaration{
			Domain: staticRuleDomain(rule),
			Kind:   rule,
			Zone:   zone,
		},
	}
}

func staticRuleDomain(rule StaticRuleKind) StaticRuleDomain {
	switch rule {
	case StaticRuleMustAttack:
		return StaticRuleDomainAttack
	case StaticRuleCantBlock, StaticRuleCantBeBlocked:
		return StaticRuleDomainBlock
	case StaticRuleCantBeCountered:
		return StaticRuleDomainCountering
	default:
		return StaticRuleDomainUnknown
	}
}

func recognizeMixedSourceStaticDeclarations(ability CompiledAbility, syntax parser.Ability) ([]StaticDeclaration, bool) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectModifyPT ||
		ability.Content.Effects[0].Duration != DurationNone ||
		len(ability.Content.Conditions) != 1 ||
		ability.Content.Conditions[0].Predicate == ConditionPredicateUnsupported ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != ReferenceBindingSource ||
		len(ability.Content.Keywords) == 0 {
		return nil, false
	}
	effect := ability.Content.Effects[0]
	if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
		return nil, false
	}
	tokens := staticDeclarationTokens(syntax)
	tokens = staticTokensWithoutCondition(tokens, ability.Content.Conditions[0].Span)
	if !matchesMixedSourcePTKeywordMustAttack(tokens, effect, ability.Content.References[0], ability.Content.Keywords) {
		return nil, false
	}
	condition := &ability.Content.Conditions[0]
	group := StaticGroupReference{Span: ability.Content.References[0].Span, Domain: StaticGroupSource}
	return []StaticDeclaration{
		staticPTDeclaration(ability.Span, group, condition, effect),
		staticKeywordGrantDeclaration(ability.Span, group, condition, ability.Content.Keywords),
		staticRuleDeclaration(ability.Span, group.Span, ability.Span, StaticRuleMustAttack, StaticZoneBattlefield, condition),
	}, true
}

func matchesMixedSourcePTKeywordMustAttack(tokens []shared.Token, effect CompiledEffect, reference CompiledReference, keywords []CompiledKeyword) bool {
	subjectLength := tokensCoveredBySpan(tokens, reference.Span)
	prefixLength := subjectLength + 6
	if subjectLength == 0 ||
		len(tokens) < prefixLength+10 ||
		!equalWord(tokens[subjectLength], "gets") ||
		!staticTokensMatchSignedAmount(tokens[subjectLength+1], tokens[subjectLength+2], effect.PowerDelta) ||
		tokens[subjectLength+3].Kind != shared.Slash ||
		!staticTokensMatchSignedAmount(tokens[subjectLength+4], tokens[subjectLength+5], effect.ToughnessDelta) ||
		tokens[prefixLength].Kind != shared.Comma ||
		!equalWord(tokens[prefixLength+1], "has") {
		return false
	}
	ruleStart := len(tokens) - 8
	if ruleStart <= prefixLength+2 ||
		tokens[ruleStart].Kind != shared.Comma ||
		!equalWord(tokens[ruleStart+1], "and") ||
		!equalWord(tokens[ruleStart+2], "attacks") ||
		!equalWord(tokens[ruleStart+3], "each") ||
		!equalWord(tokens[ruleStart+4], "combat") ||
		!equalWord(tokens[ruleStart+5], "if") ||
		!equalWord(tokens[ruleStart+6], "able") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return false
	}
	return matchesStaticKeywordList(tokens[prefixLength+2:ruleStart], keywords)
}

func recognizeStaticPowerToughnessDeclarations(ability CompiledAbility, syntax parser.Ability) ([]StaticDeclaration, bool) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectModifyPT ||
		ability.Content.Effects[0].Duration != DurationNone ||
		len(ability.Content.Conditions) > 1 {
		return nil, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return nil, false
	}
	effect := ability.Content.Effects[0]
	keywords := staticDeclarationGrantKeywords(ability.Content)
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	dynamic := effect.Amount.DynamicKind != DynamicAmountNone
	if (!dynamic && (!effect.PowerDelta.Known || !effect.ToughnessDelta.Known)) ||
		(dynamic && (!effect.PowerDelta.Known || !effect.ToughnessDelta.Known)) {
		return nil, false
	}
	tokens := staticDeclarationTokens(syntax)
	if condition != nil {
		tokens = staticTokensWithoutCondition(tokens, condition.Span)
	}
	plain := matchesStaticPTBuffSyntax(tokens, effect, group.AffectedSource, ability.Content.References)
	withKeywords := len(keywords) > 0 &&
		matchesStaticPTBuffWithKeywordsSyntax(tokens, effect, group.AffectedSource, ability.Content.References, keywords)
	if (!plain && len(keywords) == 0) || (!withKeywords && len(keywords) > 0) {
		return nil, false
	}
	declarations := []StaticDeclaration{staticPTDeclaration(ability.Span, group.Group, condition, effect)}
	if withKeywords {
		declarations = append(declarations, staticKeywordGrantDeclaration(ability.Span, group.Group, condition, keywords))
	}
	return declarations, true
}

func staticDeclarationGrantKeywords(content AbilityContent) []CompiledKeyword {
	usesCyclingPredicate := false
	for _, effect := range content.Effects {
		if effect.Selector.Keyword == parser.KeywordCycling ||
			effect.Amount.Selector().Keyword == parser.KeywordCycling {
			usesCyclingPredicate = true
			break
		}
	}
	if !usesCyclingPredicate {
		return content.Keywords
	}
	filtered := make([]CompiledKeyword, 0, len(content.Keywords))
	for _, keyword := range content.Keywords {
		if keyword.Kind == parser.KeywordCycling && keyword.ParameterKind == parser.KeywordParameterNone {
			continue
		}
		filtered = append(filtered, keyword)
	}
	return filtered
}

func recognizeStaticKeywordGrantDeclarations(ability CompiledAbility, syntax parser.Ability) ([]StaticDeclaration, bool) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectGrantKeyword ||
		ability.Content.Effects[0].Duration != DurationNone ||
		len(ability.Content.Keywords) == 0 ||
		len(ability.Content.Conditions) > 1 {
		return nil, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return nil, false
	}
	effect := ability.Content.Effects[0]
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	tokens := staticDeclarationTokens(syntax)
	if group.AffectedSource {
		if condition == nil || !matchesSourceConditionalKeywordGrant(tokens, *condition, ability.Content.Keywords) {
			return nil, false
		}
	} else if condition != nil || !matchesStaticKeywordGrant(tokens, effect.StaticSubjectSpan, ability.Content.Keywords) {
		return nil, false
	}
	return []StaticDeclaration{staticKeywordGrantDeclaration(ability.Span, group.Group, condition, ability.Content.Keywords)}, true
}

func staticDeclarationCondition(conditions []CompiledCondition) (*CompiledCondition, bool) {
	if len(conditions) == 0 {
		return nil, true
	}
	if len(conditions) != 1 || conditions[0].Predicate == ConditionPredicateUnsupported {
		return nil, false
	}
	return &conditions[0], true
}

type staticDeclarationEffectGroupResult struct {
	Group          StaticGroupReference
	AffectedSource bool
}

func staticDeclarationEffectGroup(ability CompiledAbility, effect CompiledEffect) (staticDeclarationEffectGroupResult, bool) {
	if effect.StaticSubject != StaticSubjectNone {
		if len(ability.Content.References) != 0 {
			return staticDeclarationEffectGroupResult{}, false
		}
		group, ok := staticGroupForSubject(effect.StaticSubject, effect.StaticSubjectSpan, effect.StaticSubjectSub(), effect.StaticSubjectSubKnown())
		return staticDeclarationEffectGroupResult{Group: group}, ok
	}
	if len(ability.Content.References) == 1 && ability.Content.References[0].Binding == ReferenceBindingSource {
		return staticDeclarationEffectGroupResult{
			Group: StaticGroupReference{
				Span:   ability.Content.References[0].Span,
				Domain: StaticGroupSource,
			},
			AffectedSource: true,
		}, true
	}
	return staticDeclarationEffectGroupResult{}, false
}

func staticGroupForSubject(subject StaticSubjectKind, span shared.Span, subtype types.Sub, subtypeKnown bool) (StaticGroupReference, bool) {
	group := StaticGroupReference{Span: span}
	switch subject {
	case StaticSubjectAttachedObject:
		group.Domain = StaticGroupAttachedObject
	case StaticSubjectControlledCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
	case StaticSubjectOtherControlledCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.ExcludeSource = true
	case StaticSubjectControlledWalls:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.SubtypesAny = []types.Sub{types.Wall}
	case StaticSubjectControlledArtifacts:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeArtifact}
	case StaticSubjectControlledTokens:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.TokenOnly = true
	case StaticSubjectOpponentControlledCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.Controller = ControllerOpponent
	case StaticSubjectControlledCreatureSubtype:
		if !subtypeKnown {
			return StaticGroupReference{}, false
		}
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.SubtypesAny = []types.Sub{subtype}
	case StaticSubjectOtherControlledCreatureSubtype:
		if !subtypeKnown {
			return StaticGroupReference{}, false
		}
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.SubtypesAny = []types.Sub{subtype}
		group.ExcludeSource = true
	default:
		return StaticGroupReference{}, false
	}
	return group, true
}

func staticPTDeclaration(span shared.Span, group StaticGroupReference, condition *CompiledCondition, effect CompiledEffect) StaticDeclaration {
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          span,
		OperationSpan: effect.VerbSpan,
		Group:         group,
		Condition:     condition,
		Continuous: &StaticContinuousDeclaration{
			Layer:          StaticLayerPowerToughnessModify,
			Operation:      StaticContinuousModifyPowerToughness,
			PowerDelta:     effect.PowerDelta,
			ToughnessDelta: effect.ToughnessDelta,
			DynamicAmount:  effect.Amount,
		},
	}
}

func staticKeywordGrantDeclaration(span shared.Span, group StaticGroupReference, condition *CompiledCondition, keywords []CompiledKeyword) StaticDeclaration {
	operationSpan := keywords[0].Span
	operationSpan.End = keywords[len(keywords)-1].Span.End
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          span,
		OperationSpan: operationSpan,
		Group:         group,
		Condition:     condition,
		Continuous: &StaticContinuousDeclaration{
			Layer:     StaticLayerAbility,
			Operation: StaticContinuousGrantKeywords,
			Keywords:  append([]CompiledKeyword(nil), keywords...),
		},
	}
}

func recognizeStaticCostModifierDeclaration(ability CompiledAbility) (StaticDeclaration, bool) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.References) != 0 ||
		len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Kind != parser.KeywordCycling ||
		ability.Content.Keywords[0].ParameterKind != parser.KeywordParameterNone {
		return StaticDeclaration{}, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return StaticDeclaration{}, false
	}
	cost := StaticCostModifierDeclaration{
		Kind:           StaticCostModifierAbility,
		AbilityKeyword: ability.Content.Keywords[0].Kind,
	}
	switch ability.Text {
	case "Cycling abilities you activate cost up to {2} less to activate.":
		if condition != nil {
			return StaticDeclaration{}, false
		}
		cost.GenericReduction = 2
	case "As long as you have seven or more cards in hand, you may pay {0} rather than pay cycling costs.":
		if condition == nil || condition.Predicate != ConditionPredicateControllerHandSizeAtLeast || condition.Threshold != 7 {
			return StaticDeclaration{}, false
		}
		cost.ReplaceManaCost = true
		cost.SetManaCost = ""
	case "You may pay {0} rather than pay the cycling cost of the first card you cycle each turn.":
		if condition != nil {
			return StaticDeclaration{}, false
		}
		cost.ReplaceManaCost = true
		cost.SetManaCost = ""
		cost.FirstCycleEachTurn = true
	default:
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCostModifier,
		Span:          ability.Span,
		OperationSpan: ability.Span,
		Group: StaticGroupReference{
			Span:   ability.Span,
			Domain: StaticGroupControllerHandCards,
		},
		Condition: condition,
		Cost:      &cost,
	}, true
}

func recognizeStaticCardAbilityGrantDeclaration(ability CompiledAbility, syntax parser.Ability) (StaticDeclaration, bool) {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.References) != 0 ||
		len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Kind != parser.KeywordCycling ||
		ability.Content.Keywords[0].ParameterKind != parser.KeywordParameterManaCost {
		return StaticDeclaration{}, false
	}

	text := joinedSourceText(staticDeclarationTokens(syntax))
	parameter := ability.Content.Keywords[0].Parameter
	group := StaticGroupReference{
		Span:   ability.Span,
		Domain: StaticGroupControllerHandCards,
	}
	switch text {
	case "Each land card in your hand has cycling" + parameter + ".":
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeLand}
		text = "Each land card in your hand has cycling " + parameter + "."
	case "Each creature card in your hand has cycling" + parameter + ".":
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		text = "Each creature card in your hand has cycling " + parameter + "."
	default:
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCardAbilityGrant,
		Span:          ability.Span,
		OperationSpan: ability.Content.Keywords[0].Span,
		Group:         group,
		CardGrant: &StaticCardAbilityGrantDeclaration{
			Keyword: ability.Content.Keywords[0],
			Text:    text,
		},
	}, true
}

func isHistoricCardAbilityGrant(ability CompiledAbility, syntax parser.Ability) bool {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.References) != 0 ||
		len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Kind != parser.KeywordCycling ||
		ability.Content.Keywords[0].ParameterKind != parser.KeywordParameterManaCost {
		return false
	}
	return joinedSourceText(staticDeclarationTokens(syntax)) ==
		"Each historic card in your hand has cycling"+ability.Content.Keywords[0].Parameter+"."
}

func staticDeclarationTokens(syntax parser.Ability) []shared.Token {
	tokens := semanticTokens(syntax.Tokens, syntax.Reminders, syntax.Quoted)
	if syntax.AbilityWord == nil {
		return tokens
	}
	dash := slices.IndexFunc(tokens, func(token shared.Token) bool {
		return token.Kind == shared.EmDash
	})
	if dash < 0 {
		return tokens
	}
	return tokens[dash+1:]
}

func staticTokensWithoutCondition(tokens []shared.Token, span shared.Span) []shared.Token {
	filtered := slices.DeleteFunc(append([]shared.Token(nil), tokens...), func(token shared.Token) bool {
		return spanContains(span, token.Span)
	})
	if len(filtered) > 0 && filtered[0].Kind == shared.Comma {
		filtered = filtered[1:]
	}
	return filtered
}

func spanContains(outer, inner shared.Span) bool {
	return outer.Start.Offset <= inner.Start.Offset && outer.End.Offset >= inner.End.Offset
}

func tokensCoveredBySpan(tokens []shared.Token, span shared.Span) int {
	length := 0
	for length < len(tokens) && spanContains(span, tokens[length].Span) {
		length++
	}
	return length
}

func matchesStaticKeywordGrant(tokens []shared.Token, subject shared.Span, keywords []CompiledKeyword) bool {
	subjectLength := tokensCoveredBySpan(tokens, subject)
	if subjectLength == 0 ||
		len(tokens) < subjectLength+3 ||
		(!equalWord(tokens[subjectLength], "has") && !equalWord(tokens[subjectLength], "have")) ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return false
	}
	return matchesStaticKeywordList(tokens[subjectLength+1:len(tokens)-1], keywords)
}

func matchesSourceConditionalKeywordGrant(tokens []shared.Token, condition CompiledCondition, keywords []CompiledKeyword) bool {
	return matchesPrefixSourceConditionalKeywordGrant(tokens, condition, keywords) ||
		matchesPostfixSourceConditionalKeywordGrant(tokens, condition, keywords)
}

func matchesPrefixSourceConditionalKeywordGrant(tokens []shared.Token, condition CompiledCondition, keywords []CompiledKeyword) bool {
	conditionLength := tokensCoveredBySpan(tokens, condition.Span)
	if conditionLength == 0 ||
		len(tokens) < conditionLength+6 ||
		tokens[conditionLength].Kind != shared.Comma ||
		!equalWord(tokens[conditionLength+1], "this") ||
		!equalWord(tokens[conditionLength+2], "creature") ||
		!equalWord(tokens[conditionLength+3], "has") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return false
	}
	return matchesStaticKeywordList(tokens[conditionLength+4:len(tokens)-1], keywords)
}

func matchesPostfixSourceConditionalKeywordGrant(tokens []shared.Token, condition CompiledCondition, keywords []CompiledKeyword) bool {
	if len(tokens) < 8 ||
		!equalWord(tokens[0], "this") ||
		!equalWord(tokens[1], "creature") ||
		!equalWord(tokens[2], "has") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return false
	}
	conditionStart := slices.IndexFunc(tokens, func(token shared.Token) bool {
		return spanContains(condition.Span, token.Span)
	})
	if conditionStart <= 3 {
		return false
	}
	for _, token := range tokens[conditionStart : len(tokens)-1] {
		if !spanContains(condition.Span, token.Span) {
			return false
		}
	}
	return matchesStaticKeywordList(tokens[3:conditionStart], keywords)
}

func matchesStaticPTBuffSyntax(tokens []shared.Token, effect CompiledEffect, source bool, references []CompiledReference) bool {
	prefixLength, ok := staticPTBuffPrefix(tokens, effect, source, references)
	if !ok {
		return false
	}
	if effect.Amount.DynamicKind != DynamicAmountNone {
		return len(tokens) > prefixLength+1 &&
			tokens[len(tokens)-1].Kind == shared.Period &&
			strings.ToLower(joinedSourceText(tokens[prefixLength:len(tokens)-1])) == effect.Amount.Text
	}
	return len(tokens) == prefixLength+1 && tokens[prefixLength].Kind == shared.Period
}

func matchesStaticPTBuffWithKeywordsSyntax(tokens []shared.Token, effect CompiledEffect, source bool, references []CompiledReference, keywords []CompiledKeyword) bool {
	prefixLength, ok := staticPTBuffPrefix(tokens, effect, source, references)
	if !ok ||
		len(tokens) < prefixLength+4 ||
		!equalWord(tokens[prefixLength], "and") ||
		tokens[len(tokens)-1].Kind != shared.Period {
		return false
	}
	verb := "have"
	if source || effect.StaticSubject == StaticSubjectAttachedObject ||
		(effect.StaticSubject == StaticSubjectControlledWalls && equalWord(tokens[0], "each")) {
		verb = "has"
	}
	return equalWord(tokens[prefixLength+1], verb) &&
		matchesStaticKeywordList(tokens[prefixLength+2:len(tokens)-1], keywords)
}

func staticPTBuffPrefix(tokens []shared.Token, effect CompiledEffect, source bool, references []CompiledReference) (int, bool) {
	if source {
		if len(references) != 1 {
			return 0, false
		}
		subjectLength := tokensCoveredBySpan(tokens, references[0].Span)
		prefixLength := subjectLength + 6
		return prefixLength, subjectLength > 0 &&
			len(tokens) >= prefixLength &&
			equalWord(tokens[subjectLength], "gets") &&
			staticTokensMatchSignedAmount(tokens[subjectLength+1], tokens[subjectLength+2], effect.PowerDelta) &&
			tokens[subjectLength+3].Kind == shared.Slash &&
			staticTokensMatchSignedAmount(tokens[subjectLength+4], tokens[subjectLength+5], effect.ToughnessDelta)
	}
	subjectLength := tokensCoveredBySpan(tokens, effect.StaticSubjectSpan)
	prefixLength := subjectLength + 6
	return prefixLength, subjectLength > 0 &&
		len(tokens) >= prefixLength &&
		(equalWord(tokens[subjectLength], "get") || equalWord(tokens[subjectLength], "gets")) &&
		staticTokensMatchSignedAmount(tokens[subjectLength+1], tokens[subjectLength+2], effect.PowerDelta) &&
		tokens[subjectLength+3].Kind == shared.Slash &&
		staticTokensMatchSignedAmount(tokens[subjectLength+4], tokens[subjectLength+5], effect.ToughnessDelta)
}

func staticTokensMatchSignedAmount(sign, amount shared.Token, want CompiledSignedAmount) bool {
	expectedSign := shared.Plus
	if want.Negative {
		expectedSign = shared.Minus
	}
	return sign.Kind == expectedSign && amount.Kind == shared.Integer && amount.Text == strconv.Itoa(want.Value)
}

func matchesStaticKeywordList(tokens []shared.Token, keywords []CompiledKeyword) bool {
	elements := make([]string, 0, len(tokens))
	lastKeyword := -1
	for _, token := range tokens {
		keywordIndex := -1
		for i, keyword := range keywords {
			if spanContains(keyword.Span, token.Span) {
				keywordIndex = i
				break
			}
		}
		if keywordIndex >= 0 {
			if keywordIndex != lastKeyword {
				elements = append(elements, "keyword")
				lastKeyword = keywordIndex
			}
			continue
		}
		lastKeyword = -1
		switch {
		case token.Kind == shared.Comma:
			elements = append(elements, "comma")
		case equalWord(token, "and"):
			elements = append(elements, "and")
		default:
			return false
		}

	}
	if len(keywords) == 1 {
		return slices.Equal(elements, []string{"keyword"})
	}
	if len(keywords) == 2 {
		return slices.Equal(elements, []string{"keyword", "and", "keyword"})
	}
	position := 0
	for keywordIndex := range keywords {
		if position >= len(elements) || elements[position] != "keyword" {
			return false
		}

		position++
		if keywordIndex == len(keywords)-1 {
			return position == len(elements)
		}
		if keywordIndex == len(keywords)-2 {
			if position < len(elements) && elements[position] == "comma" {
				position++
			}
			if position >= len(elements) || elements[position] != "and" {
				return false
			}
			position++
			continue
		}
		if position >= len(elements) || elements[position] != "comma" {
			return false
		}
		position++
	}
	return false
}

func staticRuleQualifiersAre(qualifiers []parser.StaticRuleQualifier, kinds ...parser.StaticRuleQualifierKind) bool {
	if len(qualifiers) != len(kinds) {
		return false
	}
	for i := range qualifiers {
		if qualifiers[i].Kind != kinds[i] {
			return false
		}
	}
	return true
}
