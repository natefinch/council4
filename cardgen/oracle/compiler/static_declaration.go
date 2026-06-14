package compiler

import (
	"slices"

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

// recognizeStaticDeclarations maps the typed static-declaration syntax the
// parser emitted for this ability onto closed semantic declarations. It consumes
// typed parser nodes and already-compiled semantic content only; it inspects no
// Oracle source text or tokens to derive meaning. Retained spans support exact
// source-consumption accounting and diagnostics.
func recognizeStaticDeclarations(compiled *CompiledAbility, syntax *parser.Ability) {
	if compiled.Kind != AbilityStatic {
		return
	}
	statics := syntax.StaticDeclarations
	if declarations, ok := recognizeTypedStaticRuleDeclarations(*compiled, syntax); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeMixedSourceStaticDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticPowerToughnessDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticKeywordGrantDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declaration, ok := recognizeStaticCostModifierDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticCardAbilityGrantDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if staticSyntaxIsHistoricCardGrant(*compiled, statics) {
		compiled.Static = &CompiledStaticSemantics{Blocker: StaticDeclarationBlockerHistoricCardSelection}
		return
	}
	if blocker := classifyStaticDeclarationBlocker(*compiled); blocker != StaticDeclarationBlockerNone {
		compiled.Static = &CompiledStaticSemantics{Blocker: blocker}
	}
}

// staticSyntaxKindsAre reports whether the parser emitted exactly the given
// declaration kinds in order.
func staticSyntaxKindsAre(statics []parser.StaticDeclarationSyntax, kinds ...parser.StaticDeclarationKind) bool {
	actual := make([]parser.StaticDeclarationKind, len(statics))
	for i := range statics {
		actual[i] = statics[i].Kind
	}
	return slices.Equal(actual, kinds)
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

func recognizeTypedStaticRuleDeclarations(ability CompiledAbility, syntax *parser.Ability) ([]StaticDeclaration, bool) {
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

func recognizeMixedSourceStaticDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics,
		parser.StaticDeclarationContinuousPowerToughness,
		parser.StaticDeclarationKeywordGrant,
		parser.StaticDeclarationRule) {
		return nil, false
	}
	rule, _, ok := semanticStaticRuleForSyntax(statics[2].Rule)
	if !ok || rule != StaticRuleMustAttack {
		return nil, false
	}
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
	effect := &ability.Content.Effects[0]
	if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
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

func recognizeStaticPowerToughnessDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	plain := staticSyntaxKindsAre(statics, parser.StaticDeclarationContinuousPowerToughness)
	withKeywords := staticSyntaxKindsAre(statics,
		parser.StaticDeclarationContinuousPowerToughness,
		parser.StaticDeclarationKeywordGrant)
	if !plain && !withKeywords {
		return nil, false
	}
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
	effect := &ability.Content.Effects[0]
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
		return nil, false
	}
	if statics[0].Dynamic != (effect.Amount.DynamicKind != DynamicAmountNone) {
		return nil, false
	}
	keywords := staticDeclarationGrantKeywords(ability.Content)
	if len(keywords) == 0 {
		if !plain {
			return nil, false
		}
	} else if !withKeywords {
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
	for i := range content.Effects {
		effect := &content.Effects[i]
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

func recognizeStaticKeywordGrantDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationKeywordGrant) {
		return nil, false
	}
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
	effect := &ability.Content.Effects[0]
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	if group.AffectedSource {
		if condition == nil {
			return nil, false
		}
	} else if condition != nil {
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

func staticDeclarationEffectGroup(ability CompiledAbility, effect *CompiledEffect) (staticDeclarationEffectGroupResult, bool) {
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

func staticPTDeclaration(span shared.Span, group StaticGroupReference, condition *CompiledCondition, effect *CompiledEffect) StaticDeclaration {
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

func recognizeStaticCostModifierDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCostModifier) {
		return StaticDeclaration{}, false
	}
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
	node := statics[0]
	cost := StaticCostModifierDeclaration{
		Kind:           StaticCostModifierAbility,
		AbilityKeyword: ability.Content.Keywords[0].Kind,
	}
	switch node.CostModifier {
	case parser.StaticDeclarationCostModifierAbilityReduction:
		if condition != nil {
			return StaticDeclaration{}, false
		}
		cost.GenericReduction = node.CostReductionAmount
	case parser.StaticDeclarationCostModifierReplaceCost:
		if condition == nil ||
			condition.Predicate != ConditionPredicateControllerHandSizeAtLeast ||
			condition.Threshold != 7 {
			return StaticDeclaration{}, false
		}
		cost.ReplaceManaCost = true
		cost.SetManaCost = node.CostReplacement
	case parser.StaticDeclarationCostModifierReplaceFirstCost:
		if condition != nil {
			return StaticDeclaration{}, false
		}
		cost.ReplaceManaCost = true
		cost.SetManaCost = node.CostReplacement
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

func recognizeStaticCardAbilityGrantDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCardAbilityGrant) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if !staticCardAbilityGrantGatingHolds(ability) {
		return StaticDeclaration{}, false
	}
	keyword := ability.Content.Keywords[0]
	group := StaticGroupReference{
		Span:   ability.Span,
		Domain: StaticGroupControllerHandCards,
	}
	var text string
	switch node.Subject.CardFilter {
	case parser.StaticDeclarationCardFilterLand:
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeLand}
		text = "Each land card in your hand has cycling " + keyword.Parameter + "."
	case parser.StaticDeclarationCardFilterCreature:
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		text = "Each creature card in your hand has cycling " + keyword.Parameter + "."
	default:
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCardAbilityGrant,
		Span:          ability.Span,
		OperationSpan: keyword.Span,
		Group:         group,
		CardGrant: &StaticCardAbilityGrantDeclaration{
			Keyword: keyword,
			Text:    text,
		},
	}, true
}

func staticSyntaxIsHistoricCardGrant(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) bool {
	return staticSyntaxKindsAre(statics, parser.StaticDeclarationCardAbilityGrant) &&
		statics[0].Subject.CardFilter == parser.StaticDeclarationCardFilterHistoric &&
		staticCardAbilityGrantGatingHolds(ability)
}

func staticCardAbilityGrantGatingHolds(ability CompiledAbility) bool {
	return ability.Cost == nil &&
		ability.Trigger == nil &&
		len(ability.Content.Modes) == 0 &&
		len(ability.Content.Targets) == 0 &&
		len(ability.Content.Conditions) == 0 &&
		len(ability.Content.References) == 0 &&
		len(ability.Content.Keywords) == 1 &&
		ability.Content.Keywords[0].Kind == parser.KeywordCycling &&
		ability.Content.Keywords[0].ParameterKind == parser.KeywordParameterManaCost
}

func spanContains(outer, inner shared.Span) bool {
	return outer.Start.Offset <= inner.Start.Offset && outer.End.Offset >= inner.End.Offset
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
