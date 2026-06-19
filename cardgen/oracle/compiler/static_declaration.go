package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/color"
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
	StaticDeclarationPlayerRule
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
	StaticLayerPowerToughnessSet
	StaticLayerType
	StaticLayerColor
	StaticLayerControl
)

// StaticContinuousOperation identifies a characteristic operation.
type StaticContinuousOperation uint8

// Static continuous-effect operations.
const (
	StaticContinuousUnknown StaticContinuousOperation = iota
	StaticContinuousModifyPowerToughness
	StaticContinuousSetBasePowerToughness
	StaticContinuousGrantKeywords
	StaticContinuousAddTypes
	StaticContinuousSetTypes
	StaticContinuousSetSubtypes
	StaticContinuousAddColors
	StaticContinuousSetColors
	StaticContinuousChangeControl
	StaticContinuousRemoveAllAbilities
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
	StaticRuleDomainAttackBlock
	StaticRuleDomainUntap
)

// Static rule declarations currently recognized by Card Generation.
const (
	StaticRuleUnknown StaticRuleKind = iota
	StaticRuleCantBlock
	StaticRuleCantBeBlocked
	StaticRuleCantAttack
	StaticRuleMustAttack
	StaticRuleMustBeBlocked
	StaticRuleCantBeCountered
	StaticRuleCantAttackOrBlock
	StaticRuleDoesntUntap
	// StaticRuleCantAttackYou prohibits attacking the source's controller or
	// their planeswalkers ("can't attack you or planeswalkers you control").
	StaticRuleCantAttackYou
	// StaticRuleCantBeBlockedByMoreThanOne bounds blocking the subject to at
	// most one creature ("can't be blocked by more than one creature").
	StaticRuleCantBeBlockedByMoreThanOne
	// StaticRuleCantBeBlockedByCreaturesWith is a restricted block prohibition
	// bounded by a blocker characteristic ("can't be blocked by creatures with
	// flying", "... power N or less", "... power N or greater"); the bounding
	// characteristic travels in StaticRuleDeclaration.Blocker.
	StaticRuleCantBeBlockedByCreaturesWith
)

// StaticBlockerRestrictionKind identifies the blocker characteristic bounding a
// restricted "can't be blocked by creatures with ..." prohibition.
type StaticBlockerRestrictionKind uint8

// Static blocker restriction kinds.
const (
	StaticBlockerRestrictionNone StaticBlockerRestrictionKind = iota
	StaticBlockerRestrictionFlying
	StaticBlockerRestrictionPowerOrLess
	StaticBlockerRestrictionPowerOrGreater
	// StaticBlockerRestrictionColor bounds the prohibition to blockers of the
	// restriction's Color ("can't be blocked by white creatures").
	StaticBlockerRestrictionColor
	// StaticBlockerRestrictionArtifact bounds the prohibition to artifact-creature
	// blockers ("can't be blocked by artifact creatures").
	StaticBlockerRestrictionArtifact
)

// StaticBlockerRestriction is the closed blocker characteristic bounding a
// restricted block prohibition. Amount is the power threshold for the
// power-comparison kinds; Color names the stopped blocker color for the color
// kind. Both are unused for kinds that do not need them.
type StaticBlockerRestriction struct {
	Kind   StaticBlockerRestrictionKind
	Amount int
	Color  color.Color
}

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
	StaticGroupControllerSpells
)

// StaticCardType identifies card types used by a static Selection.
type StaticCardType uint8

// Static Selection card types.
const (
	StaticCardTypeUnknown StaticCardType = iota
	StaticCardTypeArtifact
	StaticCardTypeCreature
	StaticCardTypeLand
	StaticCardTypeEnchantment
	StaticCardTypeInstant
	StaticCardTypeSorcery
)

// StaticCombatState constrains a static group's members by combat involvement.
type StaticCombatState uint8

// Static combat-state filters. The zero value applies no combat constraint.
const (
	StaticCombatStateAny StaticCombatState = iota
	StaticCombatStateAttacking
	StaticCombatStateBlocking
)

// StaticTapState constrains a static group's members by tapped state.
type StaticTapState uint8

// Static tap-state filters. The zero value applies no tap constraint.
const (
	StaticTapStateAny StaticTapState = iota
	StaticTapStateTapped
	StaticTapStateUntapped
)

// StaticSelection is source-independent semantic data describing WHAT objects
// in a static declaration's group match.
type StaticSelection struct {
	RequiredTypes []StaticCardType
	Supertypes    []types.Super
	SubtypesAny   []types.Sub
	ColorsAny     []color.Color
	Colorless     bool
	Multicolored  bool
	Controller    ControllerKind
	CombatState   StaticCombatState
	TapState      StaticTapState
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

	// Set base power/toughness payload (StaticContinuousSetBasePowerToughness).
	SetPower     int
	SetToughness int

	// Color characteristic payload (StaticContinuousAddColors / SetColors).
	Colors []color.Color

	// Type characteristic payload. AddTypes/AddSubtypes are additive
	// (StaticContinuousAddTypes); SetTypes/SetSubtypes replace the affected
	// object's card types and creature types (StaticContinuousSetTypes,
	// StaticContinuousSetSubtypes).
	AddTypes    []StaticCardType
	AddSubtypes []types.Sub
	SetTypes    []StaticCardType
	SetSubtypes []types.Sub
}

// StaticRuleDeclaration is one prohibition, requirement, or permission.
type StaticRuleDeclaration struct {
	Domain  StaticRuleDomain
	Kind    StaticRuleKind
	Zone    StaticZone
	Blocker StaticBlockerRestriction
}

// StaticCostModifierKind identifies which semantic cost category is modified.
type StaticCostModifierKind uint8

// Static cost modifier kinds.
const (
	StaticCostModifierUnknown StaticCostModifierKind = iota
	StaticCostModifierAbility
	StaticCostModifierSpell
)

// StaticCostModifierDeclaration is a semantic cost change.
type StaticCostModifierDeclaration struct {
	Kind               StaticCostModifierKind
	AbilityKeyword     parser.KeywordKind
	SpellTypes         []StaticCardType
	GenericReduction   int
	GenericIncrease    int
	SetManaCost        string
	ReplaceManaCost    bool
	FirstCycleEachTurn bool
}

// StaticPlayerRuleKind identifies a closed player-scoped static rule.
type StaticPlayerRuleKind uint8

// Static player rule kinds currently recognized by Card Generation.
const (
	StaticPlayerRuleUnknown StaticPlayerRuleKind = iota
	StaticPlayerRuleNoMaximumHandSize
)

// StaticPlayerRuleDeclaration is one player-scoped static rule applied to the
// static ability's controller.
type StaticPlayerRuleDeclaration struct {
	Kind StaticPlayerRuleKind
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
	Player     *StaticPlayerRuleDeclaration
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
	if declarations, ok := recognizeStaticLoseAbilitiesBecomeDeclaration(*compiled, statics); ok {
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
	if declarations, ok := recognizeStaticPowerToughnessRuleDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticKeywordGrantRuleDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticComposedContinuousDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declarations, ok := recognizeStaticKeywordGrantDeclarations(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: declarations}
		return
	}
	if declaration, ok := recognizeStaticSpellCostModifierDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
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
	if declaration, ok := recognizeStaticControlGrantDeclaration(*compiled, statics); ok {
		compiled.Static = &CompiledStaticSemantics{Declarations: []StaticDeclaration{declaration}}
		return
	}
	if declaration, ok := recognizeStaticPlayerRuleDeclaration(*compiled, statics); ok {
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
	group, ok := staticRuleGroupDomain(node.Subject.Kind)
	if !ok {
		return nil, false
	}
	if len(ability.Content.Effects) != 1 ||
		staticRuleForEffect(ability.Content.Effects[0].Kind) != rule ||
		len(ability.Content.References) != 1 ||
		ability.Content.References[0].Binding != ReferenceBindingSource {
		return nil, false
	}
	return []StaticDeclaration{staticRuleDeclaration(node.Span, node.Subject.Span, node.Operation.Span, rule, zone, group, staticBlockerRestrictionForSyntax(*node), nil)}, true
}

// staticRuleGroupDomain maps a parsed static rule subject to the affected group
// domain. Source subjects affect the object itself; an Aura or Equipment subject
// ("enchanted creature"/"equipped creature") affects the object it is attached to.
func staticRuleGroupDomain(kind parser.StaticRuleSubjectKind) (StaticGroupDomain, bool) {
	switch kind {
	case parser.StaticRuleSubjectSourceCreature, parser.StaticRuleSubjectSourceSpell:
		return StaticGroupSource, true
	case parser.StaticRuleSubjectAttachedObject:
		return StaticGroupAttachedObject, true
	default:
		return StaticGroupUnknown, false
	}
}

// isCreatureRuleSubject reports whether a static rule subject scopes a creature:
// either the source creature itself or the creature an Aura or Equipment is
// attached to. Combat and untap rule operations apply to either.
func isCreatureRuleSubject(kind parser.StaticRuleSubjectKind) bool {
	switch kind {
	case parser.StaticRuleSubjectSourceCreature, parser.StaticRuleSubjectAttachedObject:
		return true
	default:
		return false
	}
}

func semanticStaticRuleForSyntax(rule parser.StaticRuleSyntax) (StaticRuleKind, StaticZone, bool) {
	if isCreatureRuleSubject(rule.Subject.Kind) &&
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
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationBlock &&
		rule.Operation.Voice == parser.StaticRuleVoicePassive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierByMoreThanOne) {
		return StaticRuleCantBeBlockedByMoreThanOne, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationBlock &&
		rule.Operation.Voice == parser.StaticRuleVoicePassive &&
		staticBlockerRestrictionForSyntax(rule).Kind != StaticBlockerRestrictionNone {
		return StaticRuleCantBeBlockedByCreaturesWith, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationAttack &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleCantAttack, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationAttack &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierDefenderYou) {
		return StaticRuleCantAttackYou, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationAttackOrBlock &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleCantAttackOrBlock, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintProhibition &&
		rule.Operation.Kind == parser.StaticRuleOperationUntap &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		len(rule.Qualifiers) == 0 {
		return StaticRuleDoesntUntap, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintRequirement &&
		rule.Operation.Kind == parser.StaticRuleOperationAttack &&
		rule.Operation.Voice == parser.StaticRuleVoiceActive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierEachCombat, parser.StaticRuleQualifierIfAble) {
		return StaticRuleMustAttack, StaticZoneBattlefield, true
	}
	if isCreatureRuleSubject(rule.Subject.Kind) &&
		rule.Constraint.Kind == parser.StaticRuleConstraintRequirement &&
		rule.Operation.Kind == parser.StaticRuleOperationBlock &&
		rule.Operation.Voice == parser.StaticRuleVoicePassive &&
		staticRuleQualifiersAre(rule.Qualifiers, parser.StaticRuleQualifierIfAble) {
		return StaticRuleMustBeBlocked, StaticZoneBattlefield, true
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

// staticBlockerRestrictionForSyntax derives the closed blocker characteristic
// from a parsed passive block prohibition's qualifiers. A non-None result names
// a "can't be blocked by creatures with ..." restriction; an absent or
// unrecognized qualifier yields StaticBlockerRestrictionNone.
func staticBlockerRestrictionForSyntax(rule parser.StaticRuleSyntax) StaticBlockerRestriction {
	if len(rule.Qualifiers) != 1 {
		return StaticBlockerRestriction{}
	}
	qualifier := rule.Qualifiers[0]
	switch qualifier.Kind {
	case parser.StaticRuleQualifierBlockerFlying:
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionFlying}
	case parser.StaticRuleQualifierBlockerPowerOrLess:
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionPowerOrLess, Amount: qualifier.Amount}
	case parser.StaticRuleQualifierBlockerPowerOrGreater:
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionPowerOrGreater, Amount: qualifier.Amount}
	case parser.StaticRuleQualifierBlockerColor:
		runtimeColor, ok := compilerColor(qualifier.Color)
		if !ok {
			return StaticBlockerRestriction{}
		}
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionColor, Color: runtimeColor}
	case parser.StaticRuleQualifierBlockerArtifact:
		return StaticBlockerRestriction{Kind: StaticBlockerRestrictionArtifact}
	default:
		return StaticBlockerRestriction{}
	}
}

func staticRuleForEffect(kind EffectKind) StaticRuleKind {
	switch kind {
	case EffectCantBlock:
		return StaticRuleCantBlock
	case EffectCantBeBlocked:
		return StaticRuleCantBeBlocked
	case EffectCantAttack:
		return StaticRuleCantAttack
	case EffectMustAttack:
		return StaticRuleMustAttack
	case EffectMustBeBlocked:
		return StaticRuleMustBeBlocked
	case EffectCantBeCountered:
		return StaticRuleCantBeCountered
	case EffectCantBeBlockedByCreaturesWith:
		return StaticRuleCantBeBlockedByCreaturesWith
	case EffectCantBeBlockedByMoreThanOne:
		return StaticRuleCantBeBlockedByMoreThanOne
	case EffectCantAttackOrBlock:
		return StaticRuleCantAttackOrBlock
	case EffectDoesntUntap:
		return StaticRuleDoesntUntap
	default:
		return StaticRuleUnknown
	}
}

func staticRuleDeclaration(
	span, subjectSpan, operationSpan shared.Span,
	rule StaticRuleKind,
	zone StaticZone,
	group StaticGroupDomain,
	blocker StaticBlockerRestriction,
	condition *CompiledCondition,
) StaticDeclaration {
	return StaticDeclaration{
		Kind:          StaticDeclarationRule,
		Span:          span,
		OperationSpan: operationSpan,
		Group: StaticGroupReference{
			Span:   subjectSpan,
			Domain: group,
		},
		Condition: condition,
		Rule: &StaticRuleDeclaration{
			Domain:  staticRuleDomain(rule),
			Kind:    rule,
			Zone:    zone,
			Blocker: blocker,
		},
	}
}

func staticRuleDomain(rule StaticRuleKind) StaticRuleDomain {
	switch rule {
	case StaticRuleCantAttack, StaticRuleMustAttack, StaticRuleCantAttackYou:
		return StaticRuleDomainAttack
	case StaticRuleCantBlock, StaticRuleCantBeBlocked, StaticRuleMustBeBlocked, StaticRuleCantBeBlockedByMoreThanOne,
		StaticRuleCantBeBlockedByCreaturesWith:
		return StaticRuleDomainBlock
	case StaticRuleCantBeCountered:
		return StaticRuleDomainCountering
	case StaticRuleCantAttackOrBlock:
		return StaticRuleDomainAttackBlock
	case StaticRuleDoesntUntap:
		return StaticRuleDomainUntap
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
		staticRuleDeclaration(ability.Span, group.Span, ability.Span, StaticRuleMustAttack, StaticZoneBattlefield, StaticGroupSource, StaticBlockerRestriction{}, condition),
	}, true
}

// recognizeStaticLoseAbilitiesBecomeDeclaration maps the polymorph syntax
// "<subject> loses all abilities [and is/has ...]" onto layer-faithful semantic
// declarations: a remove-all-abilities ability-layer declaration, plus optional
// set-color, set-type, set-subtype, and base power/toughness declarations. The
// affected object's existing colors, card types, and creature types are replaced
// (set), so the colors and types travel as set operations rather than additions.
func recognizeStaticLoseAbilitiesBecomeDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationLoseAbilitiesBecome) {
		return nil, false
	}
	node := &statics[0]
	if !node.LoseAllAbilities {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		ability.AbilityWord != "" {
		return nil, false
	}
	group, ok := staticGroupForParserSubject(node.Subject)
	if !ok {
		return nil, false
	}
	declarations := []StaticDeclaration{{
		Kind:          StaticDeclarationContinuous,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Group:         group,
		Continuous: &StaticContinuousDeclaration{
			Layer:     StaticLayerAbility,
			Operation: StaticContinuousRemoveAllAbilities,
		},
	}}
	if len(node.Colors) != 0 {
		colors, ok := staticRuntimeColors(node.Colors)
		if !ok {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:     StaticLayerColor,
				Operation: StaticContinuousSetColors,
				Colors:    colors,
			},
		})
	}
	if len(node.CardTypes) != 0 || len(node.Subtypes) != 0 {
		cardTypes, ok := staticCardTypesFromParser(node.CardTypes)
		if !ok {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          node.Span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Continuous: &StaticContinuousDeclaration{
				Layer:       StaticLayerType,
				Operation:   StaticContinuousSetTypes,
				SetTypes:    cardTypes,
				SetSubtypes: slices.Clone(node.Subtypes),
			},
		})
	}
	if node.BasePTSet {
		declarations = append(declarations, staticBasePowerToughnessDeclaration(node.Span, node, group, nil))
	}
	return declarations, true
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

// recognizeStaticPowerToughnessRuleDeclarations maps a paragraph that composes a
// power/toughness modification (optionally with a keyword grant) and a single
// creature-scoped rule operation onto closed semantic declarations, e.g.
// "Enchanted creature gets +2/+2 and can't block." The resolving content carries
// only the power/toughness effect, so the rule operation derives from the typed
// parser node; the affected group derives from the resolving effect, keeping the
// mapping text-blind. Conditional compounds fail closed because static rule
// effects are recognized only without a condition.
func recognizeStaticPowerToughnessRuleDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	plain := staticSyntaxKindsAre(statics,
		parser.StaticDeclarationContinuousPowerToughness,
		parser.StaticDeclarationRule)
	withKeywords := staticSyntaxKindsAre(statics,
		parser.StaticDeclarationContinuousPowerToughness,
		parser.StaticDeclarationKeywordGrant,
		parser.StaticDeclarationRule)
	if !plain && !withKeywords {
		return nil, false
	}
	ruleNode := &statics[len(statics)-1]
	rule, zone, ok := semanticStaticRuleForSyntax(ruleNode.Rule)
	if !ok {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectModifyPT ||
		ability.Content.Effects[0].Duration != DurationNone {
		return nil, false
	}
	effect := &ability.Content.Effects[0]
	if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
		return nil, false
	}
	if statics[0].Dynamic != (effect.Amount.DynamicKind != DynamicAmountNone) {
		return nil, false
	}
	keywords := staticDeclarationGrantKeywords(ability.Content)
	if (len(keywords) != 0) != withKeywords {
		return nil, false
	}
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	ruleGroup, ok := staticRuleGroupDomain(ruleNode.Rule.Subject.Kind)
	if !ok || ruleGroup != group.Group.Domain {
		return nil, false
	}
	declarations := []StaticDeclaration{staticPTDeclaration(ability.Span, group.Group, nil, effect)}
	if withKeywords {
		declarations = append(declarations, staticKeywordGrantDeclaration(ability.Span, group.Group, nil, keywords))
	}
	declarations = append(declarations, staticRuleDeclaration(ability.Span, group.Group.Span, ruleNode.OperationSpan, rule, zone, group.Group.Domain, staticBlockerRestrictionForSyntax(ruleNode.Rule), nil))
	return declarations, true
}

// recognizeStaticKeywordGrantRuleDeclarations maps a paragraph that composes a
// keyword grant and a single creature-scoped rule operation, without any
// power/toughness change, onto closed semantic declarations, e.g. "Enchanted
// creature has trample and can't be blocked by more than one creature." The
// resolving content carries only the keyword-grant effect, so the rule operation
// derives from the typed parser node; the affected group derives from the
// resolving effect, keeping the mapping text-blind. Conditional compounds fail
// closed because static rule effects are recognized only without a condition.
func recognizeStaticKeywordGrantRuleDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationKeywordGrant, parser.StaticDeclarationRule) {
		return nil, false
	}
	ruleNode := &statics[len(statics)-1]
	rule, zone, ok := semanticStaticRuleForSyntax(ruleNode.Rule)
	if !ok {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 1 ||
		ability.Content.Effects[0].Kind != EffectGrantKeyword ||
		ability.Content.Effects[0].Duration != DurationNone {
		return nil, false
	}
	keywords := staticDeclarationGrantKeywords(ability.Content)
	if len(keywords) == 0 {
		return nil, false
	}
	effect := &ability.Content.Effects[0]
	group, ok := staticDeclarationEffectGroup(ability, effect)
	if !ok {
		return nil, false
	}
	ruleGroup, ok := staticRuleGroupDomain(ruleNode.Rule.Subject.Kind)
	if !ok || ruleGroup != group.Group.Domain {
		return nil, false
	}
	return []StaticDeclaration{
		staticKeywordGrantDeclaration(ability.Span, group.Group, nil, keywords),
		staticRuleDeclaration(ability.Span, group.Group.Span, ruleNode.OperationSpan, rule, zone, group.Group.Domain, staticBlockerRestrictionForSyntax(ruleNode.Rule), nil),
	}, true
}

// shared affected group with one or more layer-preserving characteristic changes
// onto closed semantic declarations. It recognizes power/toughness modification,
// base power/toughness setting, keyword grants, and color/type characteristic
// additions, requiring at least one base-power/toughness or characteristic node so
// the simpler single-family recognizers keep ownership of their shapes. The group
// and payload derive from the typed parser nodes and already-resolved content
// only; no Oracle text is inspected.
func recognizeStaticComposedContinuousDeclarations(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) ([]StaticDeclaration, bool) {
	if len(statics) == 0 {
		return nil, false
	}
	ptNodes := 0
	keywordNodes := 0
	newNodes := 0
	for i := range statics {
		switch statics[i].Kind {
		case parser.StaticDeclarationContinuousPowerToughness:
			ptNodes++
		case parser.StaticDeclarationKeywordGrant:
			keywordNodes++
		case parser.StaticDeclarationContinuousBasePowerToughness,
			parser.StaticDeclarationContinuousCharacteristic:
			newNodes++
		default:
			return nil, false
		}
	}
	if newNodes == 0 {
		return nil, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) > 1 {
		return nil, false
	}
	condition, ok := staticDeclarationCondition(ability.Content.Conditions)
	if !ok {
		return nil, false
	}
	subject := statics[0].Subject
	for i := range statics {
		if !staticSubjectsEquivalent(statics[i].Subject, subject) {
			return nil, false
		}
	}
	group, ok := staticGroupForParserSubject(subject)
	if !ok {
		return nil, false
	}
	// Cross-check the resolving content shape against the typed operations. The
	// "has base power and toughness" verb yields an empty keyword-grant effect
	// shell with no keywords, which is tolerated only when no keyword node is
	// present.
	modifyPT := 0
	for i := range ability.Content.Effects {
		switch ability.Content.Effects[i].Kind {
		case EffectModifyPT:
			modifyPT++
		case EffectGrantKeyword:
		default:
			return nil, false
		}
	}
	if modifyPT != ptNodes {
		return nil, false
	}
	if (keywordNodes > 0) != (len(ability.Content.Keywords) > 0) {
		return nil, false
	}
	if keywordNodes > 1 {
		return nil, false
	}
	var ptEffect *CompiledEffect
	for i := range ability.Content.Effects {
		if ability.Content.Effects[i].Kind == EffectModifyPT {
			ptEffect = &ability.Content.Effects[i]
		}
	}
	keywordsEmitted := false
	var declarations []StaticDeclaration
	for i := range statics {
		node := &statics[i]
		switch node.Kind {
		case parser.StaticDeclarationContinuousPowerToughness:
			if ptEffect == nil ||
				!ptEffect.PowerDelta.Known ||
				!ptEffect.ToughnessDelta.Known ||
				ptEffect.Duration != DurationNone {
				return nil, false
			}
			if node.Dynamic != (ptEffect.Amount.DynamicKind != DynamicAmountNone) {
				return nil, false
			}
			declarations = append(declarations, staticPTDeclaration(ability.Span, group, condition, ptEffect))
		case parser.StaticDeclarationKeywordGrant:
			if keywordsEmitted || len(ability.Content.Keywords) == 0 {
				return nil, false
			}
			keywordsEmitted = true
			declarations = append(declarations, staticKeywordGrantDeclaration(ability.Span, group, condition, ability.Content.Keywords))
		case parser.StaticDeclarationContinuousBasePowerToughness:
			if !node.BasePTSet {
				return nil, false
			}
			declarations = append(declarations, staticBasePowerToughnessDeclaration(ability.Span, node, group, condition))
		case parser.StaticDeclarationContinuousCharacteristic:
			characteristic, ok := staticCharacteristicDeclarations(ability.Span, node, group, condition)
			if !ok {
				return nil, false
			}
			declarations = append(declarations, characteristic...)
		default:
			return nil, false
		}
	}
	if len(declarations) == 0 {
		return nil, false
	}
	return declarations, true
}

// staticSubjectsEquivalent reports whether two typed parser subjects name the
// same affected group. It compares only typed identity fields and ignores source
// spans so recognition stays position-blind.
func staticSubjectsEquivalent(a, b parser.StaticDeclarationSubject) bool {
	return a.Kind == b.Kind &&
		a.CardFilter == b.CardFilter &&
		a.Group.Kind == b.Group.Kind &&
		a.Group.Subtype == b.Group.Subtype &&
		a.Group.SubtypeKnown == b.Group.SubtypeKnown &&
		a.Group.Colorless == b.Group.Colorless &&
		a.Group.Multicolored == b.Group.Multicolored &&
		slices.Equal(a.Group.Colors, b.Group.Colors)
}

// staticGroupForParserSubject maps a typed parser subject onto the affected group
// reference, failing closed for subjects whose runtime group is not representable.
func staticGroupForParserSubject(subject parser.StaticDeclarationSubject) (StaticGroupReference, bool) {
	switch subject.Kind {
	case parser.StaticDeclarationSubjectSourceCreature,
		parser.StaticDeclarationSubjectSourceNamed:
		return StaticGroupReference{Span: subject.Span, Domain: StaticGroupSource}, true
	case parser.StaticDeclarationSubjectGroup:
		kind := compileStaticSubjectKind(subject.Group.Kind)
		if kind == StaticSubjectNone {
			return StaticGroupReference{}, false
		}
		return staticGroupForSubject(kind, subject.Group.Span, subject.Group.Subtype, subject.Group.SubtypeKnown, staticColorFilter{
			Colors:       subject.Group.Colors,
			Colorless:    subject.Group.Colorless,
			Multicolored: subject.Group.Multicolored,
		})
	default:
		return StaticGroupReference{}, false
	}
}

// staticBasePowerToughnessDeclaration builds a base power/toughness setting
// declaration from the typed parser payload.
func staticBasePowerToughnessDeclaration(span shared.Span, node *parser.StaticDeclarationSyntax, group StaticGroupReference, condition *CompiledCondition) StaticDeclaration {
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          span,
		OperationSpan: node.OperationSpan,
		Group:         group,
		Condition:     condition,
		Continuous: &StaticContinuousDeclaration{
			Layer:        StaticLayerPowerToughnessSet,
			Operation:    StaticContinuousSetBasePowerToughness,
			SetPower:     node.BasePower,
			SetToughness: node.BaseToughness,
		},
	}
}

// staticCharacteristicDeclarations splits a "<group> is/are ... in addition"
// declaration into separate color and type layer declarations. Colors are set
// when no "in addition" tail is present and added otherwise; card types and
// subtypes are always additive. It fails closed for an unrepresentable color or
// card type.
func staticCharacteristicDeclarations(span shared.Span, node *parser.StaticDeclarationSyntax, group StaticGroupReference, condition *CompiledCondition) ([]StaticDeclaration, bool) {
	var declarations []StaticDeclaration
	if len(node.Colors) != 0 {
		colors, ok := staticRuntimeColors(node.Colors)
		if !ok {
			return nil, false
		}
		operation := StaticContinuousSetColors
		if node.ColorsAdd {
			operation = StaticContinuousAddColors
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Condition:     condition,
			Continuous: &StaticContinuousDeclaration{
				Layer:     StaticLayerColor,
				Operation: operation,
				Colors:    colors,
			},
		})
	}
	if len(node.CardTypes) != 0 || len(node.Subtypes) != 0 {
		cardTypes, ok := staticCardTypesFromParser(node.CardTypes)
		if !ok {
			return nil, false
		}
		declarations = append(declarations, StaticDeclaration{
			Kind:          StaticDeclarationContinuous,
			Span:          span,
			OperationSpan: node.OperationSpan,
			Group:         group,
			Condition:     condition,
			Continuous: &StaticContinuousDeclaration{
				Layer:       StaticLayerType,
				Operation:   StaticContinuousAddTypes,
				AddTypes:    cardTypes,
				AddSubtypes: slices.Clone(node.Subtypes),
			},
		})
	}
	if len(declarations) == 0 {
		return nil, false
	}
	return declarations, true
}

func staticRuntimeColors(colors []parser.Color) ([]color.Color, bool) {
	result := make([]color.Color, 0, len(colors))
	for _, value := range colors {
		runtime, ok := compilerColor(value)
		if !ok {
			return nil, false
		}
		result = append(result, runtime)
	}
	return result, true
}

func staticCardTypesFromParser(cardTypes []parser.CardType) ([]StaticCardType, bool) {
	result := make([]StaticCardType, 0, len(cardTypes))
	for _, value := range cardTypes {
		mapped, ok := staticCardTypeFromParser(value)
		if !ok {
			return nil, false
		}
		result = append(result, mapped)
	}
	return result, true
}

func staticCardTypeFromParser(value parser.CardType) (StaticCardType, bool) {
	switch value {
	case parser.CardTypeArtifact:
		return StaticCardTypeArtifact, true
	case parser.CardTypeCreature:
		return StaticCardTypeCreature, true
	case parser.CardTypeLand:
		return StaticCardTypeLand, true
	case parser.CardTypeEnchantment:
		return StaticCardTypeEnchantment, true
	case parser.CardTypeInstant:
		return StaticCardTypeInstant, true
	case parser.CardTypeSorcery:
		return StaticCardTypeSorcery, true
	default:
		return StaticCardTypeUnknown, false
	}
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
		group, ok := staticGroupForSubject(effect.StaticSubject, effect.StaticSubjectSpan, effect.StaticSubjectSub(), effect.StaticSubjectSubKnown(), staticColorFilter{
			Colors:       effect.StaticSubjectColorsAny(),
			Colorless:    effect.StaticSubjectColorless(),
			Multicolored: effect.StaticSubjectMulticolored(),
		})
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

func staticGroupForSubject(subject StaticSubjectKind, span shared.Span, subtype types.Sub, subtypeKnown bool, colors staticColorFilter) (StaticGroupReference, bool) {
	group := StaticGroupReference{Span: span}
	switch subject {
	case StaticSubjectAttachedObject:
		group.Domain = StaticGroupAttachedObject
	case StaticSubjectAllCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
	case StaticSubjectAllOtherCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.ExcludeSource = true
	case StaticSubjectAttackingCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.CombatState = StaticCombatStateAttacking
	case StaticSubjectBlockingCreatures:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.CombatState = StaticCombatStateBlocking
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
	case StaticSubjectAllCreatureSubtype:
		if !subtypeKnown {
			return StaticGroupReference{}, false
		}
		group.Domain = StaticGroupBattlefield
		group.Selection.SubtypesAny = []types.Sub{subtype}
	case StaticSubjectOtherCreatureSubtype:
		if !subtypeKnown {
			return StaticGroupReference{}, false
		}
		group.Domain = StaticGroupBattlefield
		group.Selection.SubtypesAny = []types.Sub{subtype}
		group.ExcludeSource = true
	case StaticSubjectControlledAttackingCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.CombatState = StaticCombatStateAttacking
	case StaticSubjectControlledCreatureTokens:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.TokenOnly = true
	case StaticSubjectBattlefieldCreatureTokens:
		group.Domain = StaticGroupBattlefield
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.TokenOnly = true
	case StaticSubjectControlledLegendaryCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.Supertypes = []types.Super{types.Legendary}
	case StaticSubjectControlledUntappedCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.TapState = StaticTapStateUntapped
	case StaticSubjectOtherControlledTappedCreatures:
		group.Domain = StaticGroupSourceControllerPermanents
		group.Selection.RequiredTypes = []StaticCardType{StaticCardTypeCreature}
		group.Selection.TapState = StaticTapStateTapped
		group.ExcludeSource = true
	default:
		return StaticGroupReference{}, false
	}
	if !applyStaticColorFilter(&group.Selection, colors) {
		return StaticGroupReference{}, false
	}
	return group, true
}

// staticColorFilter is the closed color constraint an affected creature group
// may carry ("Other red creatures you control ..."). The zero value applies no
// color constraint.
type staticColorFilter struct {
	Colors       []parser.Color
	Colorless    bool
	Multicolored bool
}

// applyStaticColorFilter sets the Selection's color predicate from a typed color
// filter, failing closed for any color word that has no runtime representation.
func applyStaticColorFilter(selection *StaticSelection, colors staticColorFilter) bool {
	for _, value := range colors.Colors {
		runtime, ok := compilerColor(value)
		if !ok {
			return false
		}
		selection.ColorsAny = append(selection.ColorsAny, runtime)
	}
	selection.Colorless = colors.Colorless
	selection.Multicolored = colors.Multicolored
	return true
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

// recognizeStaticSpellCostModifierDeclaration maps the typed spell cast-cost
// modifier syntax onto a closed semantic cost declaration. The affected group is
// the static ability's controller's spells; the optional spell-type filter is a
// closed set of card types matched as a disjunction at runtime.
func recognizeStaticSpellCostModifierDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationCostModifier) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.CostModifier != parser.StaticDeclarationCostModifierSpellReduction &&
		node.CostModifier != parser.StaticDeclarationCostModifierSpellIncrease {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.References) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.Conditions) != 0 {
		return StaticDeclaration{}, false
	}
	spellTypes, ok := staticSpellTypeCardTypes(node.SpellType)
	if !ok {
		return StaticDeclaration{}, false
	}
	if node.CostReductionAmount <= 0 {
		return StaticDeclaration{}, false
	}
	cost := StaticCostModifierDeclaration{
		Kind:       StaticCostModifierSpell,
		SpellTypes: spellTypes,
	}
	if node.CostModifier == parser.StaticDeclarationCostModifierSpellIncrease {
		cost.GenericIncrease = node.CostReductionAmount
	} else {
		cost.GenericReduction = node.CostReductionAmount
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationCostModifier,
		Span:          ability.Span,
		OperationSpan: node.OperationSpan,
		Group: StaticGroupReference{
			Span:   ability.Span,
			Domain: StaticGroupControllerSpells,
		},
		Cost: &cost,
	}, true
}

// staticSpellTypeCardTypes maps a closed spell-type filter onto the card types
// whose disjunction the runtime matches. An all-spells filter returns no types.
func staticSpellTypeCardTypes(filter parser.StaticDeclarationSpellTypeKind) ([]StaticCardType, bool) {
	switch filter {
	case parser.StaticDeclarationSpellTypeAll:
		return nil, true
	case parser.StaticDeclarationSpellTypeArtifact:
		return []StaticCardType{StaticCardTypeArtifact}, true
	case parser.StaticDeclarationSpellTypeCreature:
		return []StaticCardType{StaticCardTypeCreature}, true
	case parser.StaticDeclarationSpellTypeEnchantment:
		return []StaticCardType{StaticCardTypeEnchantment}, true
	case parser.StaticDeclarationSpellTypeInstant:
		return []StaticCardType{StaticCardTypeInstant}, true
	case parser.StaticDeclarationSpellTypeSorcery:
		return []StaticCardType{StaticCardTypeSorcery}, true
	case parser.StaticDeclarationSpellTypeInstantOrSorcery:
		return []StaticCardType{StaticCardTypeInstant, StaticCardTypeSorcery}, true
	default:
		return nil, false
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

func recognizeStaticControlGrantDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationControlGrant) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.Subject.Kind != parser.StaticDeclarationSubjectGroup ||
		node.Subject.Group.Kind != parser.EffectStaticSubjectAttachedObject {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationContinuous,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Group: StaticGroupReference{
			Span:   node.Subject.Span,
			Domain: StaticGroupAttachedObject,
		},
		Continuous: &StaticContinuousDeclaration{
			Layer:     StaticLayerControl,
			Operation: StaticContinuousChangeControl,
		},
	}, true
}

// recognizeStaticPlayerRuleDeclaration recognizes the controller-scoped static
// rule "You have no maximum hand size." emitted by the parser. The declaration
// requires an otherwise empty static ability shell.
func recognizeStaticPlayerRuleDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationPlayerRule) {
		return StaticDeclaration{}, false
	}
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if node.Subject.Kind != parser.StaticDeclarationSubjectController ||
		node.PlayerRule != parser.StaticDeclarationPlayerRuleNoMaximumHandSize {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationPlayerRule,
		Span:          node.Span,
		OperationSpan: node.OperationSpan,
		Player:        &StaticPlayerRuleDeclaration{Kind: StaticPlayerRuleNoMaximumHandSize},
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

func staticRuleQualifiersAre(qualifiers []parser.StaticRuleQualifier, kinds ...parser.StaticRuleQualifierKind) bool {
	actual := make([]parser.StaticRuleQualifierKind, len(qualifiers))
	for i := range qualifiers {
		actual[i] = qualifiers[i].Kind
	}
	return slices.Equal(actual, kinds)
}
