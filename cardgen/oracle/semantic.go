package oracle

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Compilation is the semantic result for one card face.
type Compilation struct {
	Syntax    Document
	Abilities []CompiledAbility
}

// CompiledAbility is a source-spanned semantic ability. Shell semantics
// (cost, trigger, timing, chapter numbers) are fields on CompiledAbility;
// the ability's instruction content (targets, conditions, effects, keywords,
// references, modes) lives in Content.
type CompiledAbility struct {
	Kind                 AbilityKind
	Span                 Span
	Text                 string
	ActivationTiming     ActivationTimingKind
	ActivationTimingSpan Span
	ActivationZone       zone.Type
	AbilityWord          string
	Chapters             []int
	ChapterSpan          Span
	Optional             bool
	OptionalSpan         Span
	Cost                 *CompiledCost
	Trigger              *CompiledTrigger
	Content              AbilityContent
	Static               *CompiledStaticSemantics
}

// ActivationTimingKind identifies an exact restriction on when an activated
// ability may be activated.
type ActivationTimingKind int

// Activation timing kinds recognized by the semantic compiler.
const (
	ActivationTimingNone ActivationTimingKind = iota
	ActivationTimingSorcery
	ActivationTimingOncePerTurn
	ActivationTimingSorceryOncePerTurn
	ActivationTimingDuringCombat
	ActivationTimingDuringUpkeep
	ActivationTimingUnsupported
)

// AbilityContent is the reusable semantic content of an ability, independent
// of its shell (spell, activated, triggered, loyalty, chapter, or modal
// option). It owns the ordered targets, conditions, effects, keywords,
// references, and nested modes that form an ability's instruction content.
type AbilityContent struct {
	Span       Span
	Modes      []CompiledMode
	Targets    []CompiledTarget
	Conditions []CompiledCondition
	Effects    []CompiledEffect
	Keywords   []CompiledKeyword
	References []CompiledReference
}

// Unconsumed reports whether any sidechannel content fields (targets,
// conditions, keywords, modes, or references) are non-empty. Effect
// consumption is checked separately by lowering code since effect count
// drives dispatch. Lowerers that deliberately consume references must clear
// them from the content before calling Unconsumed.
func (c AbilityContent) Unconsumed() bool {
	return len(c.Targets) != 0 ||
		len(c.Conditions) != 0 ||
		len(c.Keywords) != 0 ||
		len(c.Modes) != 0 ||
		len(c.References) != 0
}

// CompiledMode is one semantic option in a modal ability.
type CompiledMode struct {
	Span    Span
	Text    string
	Content AbilityContent
}

// CostKind identifies a component paid to activate an ability.
type CostKind uint8

// Recognized cost component kinds.
const (
	CostUnknown CostKind = iota
	CostMana
	CostTap
	CostUntap
	CostSacrifice
	CostDiscard
	CostPayLife
	CostExile
	CostRemoveCounter
	CostReveal
	CostTapPermanents
	CostEnergy
	CostReturn
	CostExert
	CostMill
	CostPutCounter
	CostCollectEvidence
	CostLoyalty
)

// CompiledCost is the ordered cost before an activated ability's colon.
type CompiledCost struct {
	Span       Span
	Text       string
	Components []CostComponent
}

// CostComponent is one comma-separated cost operation.
type CostComponent struct {
	Kind   CostKind
	Span   Span
	Text   string
	Symbol string
	Amount string
	Object string
}

// TriggerKind identifies the leading trigger word.
type TriggerKind uint8

// Trigger kinds.
const (
	TriggerUnknown TriggerKind = iota
	TriggerWhen
	TriggerWhenever
	TriggerAt
)

// CompiledTrigger is the event clause before a triggered ability's first
// top-level comma.
type CompiledTrigger struct {
	Kind TriggerKind
	Span Span
	Text string
	// Event retains the exact event-clause text for diagnostics and source
	// consumption. Executable lowering consumes Pattern instead.
	Event     string
	Pattern   TriggerPattern
	Condition *CompiledCondition
}

// ConditionKind identifies recognized conditional wording.
type ConditionKind uint8

// Condition kinds.
const (
	ConditionUnknown ConditionKind = iota
	ConditionIf
	ConditionUnless
	ConditionOnlyIf
	ConditionAsLongAs
)

// ConditionPredicate identifies the closed semantic predicate recognized in a
// condition.
type ConditionPredicate uint8

// Condition predicates recognized by the semantic compiler.
const (
	ConditionPredicateUnsupported ConditionPredicate = iota
	ConditionPredicateControllerLifeAtLeast
	ConditionPredicateControllerHandSizeAtLeast
	ConditionPredicateAnyPlayerLifeAtMost
	ConditionPredicateOpponentCountAtLeast
	ConditionPredicateControllerControls
	ConditionPredicateAnyOpponentControls
	ConditionPredicateOpponentsControl
	ConditionPredicateControllerHandEmpty
	ConditionPredicateControllerGraveyardCardCountAtLeast
	ConditionPredicateControllerGraveyardCardTypeCountAtLeast
	ConditionPredicateControllerCreaturePowerDiversityAtLeast
	ConditionPredicateEventSubjectWasKicked
	ConditionPredicateEventSubjectWasCast
	ConditionPredicateEventSubjectWasCastByController
	ConditionPredicateEventSubjectHadNoCounter
	ConditionPredicatePriorInstructionNotAccepted
	ConditionPredicateCounterPlacementOnControlledCreature
	ConditionPredicateControllerCounterPlacement
	ConditionPredicateDamageByControlledSource
	ConditionPredicateTokenCreationUnderController
	ConditionPredicateSourceWouldDie
	ConditionPredicateSourceWouldGoToGraveyard
	ConditionPredicateTargetControllerDoesNotPay
)

// ConditionCardType identifies a card type in a semantic condition Selection.
type ConditionCardType uint8

// Condition card types.
const (
	ConditionCardTypeUnknown ConditionCardType = iota
	ConditionCardTypeArtifact
	ConditionCardTypeBattle
	ConditionCardTypeCreature
	ConditionCardTypeEnchantment
	ConditionCardTypeLand
	ConditionCardTypePlaneswalker
)

// ConditionSupertype identifies a supertype in a semantic condition Selection.
type ConditionSupertype uint8

// Condition supertypes.
const (
	ConditionSupertypeUnknown ConditionSupertype = iota
	ConditionSupertypeBasic
	ConditionSupertypeSnow
)

// ConditionColor identifies a color in a semantic condition Selection.
type ConditionColor uint8

// Condition colors.
const (
	ConditionColorUnknown ConditionColor = iota
	ConditionColorWhite
	ConditionColorBlue
	ConditionColorBlack
	ConditionColorRed
	ConditionColorGreen
)

// ConditionCounter identifies a counter kind in an event-subject condition.
type ConditionCounter uint8

// Condition counter kinds.
const (
	ConditionCounterUnknown ConditionCounter = iota
	ConditionCounterPlusOnePlusOne
	ConditionCounterMinusOneMinusOne
)

// ConditionSelection is the source-independent Selection vocabulary used by
// semantic conditions. Subtype names are canonicalized during recognition.
type ConditionSelection struct {
	RequiredTypes     []ConditionCardType
	Supertypes        []ConditionSupertype
	SubtypesAny       []string
	ColorsAny         []ConditionColor
	Colorless         bool
	ExcludeSource     bool
	PowerAtLeast      int
	MatchPowerAtLeast bool
}

// CompiledCondition is a closed, source-spanned semantic condition.
type CompiledCondition struct {
	Kind        ConditionKind
	Span        Span
	Text        string
	Intervening bool
	Predicate   ConditionPredicate
	Negated     bool
	Threshold   int
	Selection   ConditionSelection
	Counter     ConditionCounter
}

// TargetCardinality is an inclusive target count range.
type TargetCardinality struct {
	Min int
	Max int
}

// CompiledTarget is one occurrence of the word "target" and its local noun
// phrase.
type CompiledTarget struct {
	Span        Span
	Text        string
	Cardinality TargetCardinality
	Selector    CompiledSelector
}

// SelectorKind identifies the broad object selected by a phrase.
type SelectorKind uint8

// Selector kinds.
const (
	SelectorUnknown SelectorKind = iota
	SelectorAny
	SelectorPlayer
	SelectorOpponent
	SelectorArtifact
	SelectorCreature
	SelectorEnchantment
	SelectorLand
	SelectorPermanent
	SelectorCard
	SelectorSpell
	SelectorActivatedAbility
	SelectorTriggeredAbility
	SelectorActivatedOrTriggeredAbility
	SelectorSpellActivatedOrTriggeredAbility
	SelectorPlaneswalker
	SelectorBattle
)

// ControllerKind constrains a selected object by controller.
type ControllerKind uint8

// Controller constraints.
const (
	ControllerAny ControllerKind = iota
	ControllerYou
	ControllerOpponent
	ControllerNotYou
)

// CompiledSelector is a conservative semantic summary of a noun phrase.
type CompiledSelector struct {
	Kind       SelectorKind
	Controller ControllerKind
	Another    bool
	Other      bool
	Attacking  bool
	Blocking   bool
	Tapped     bool
	Untapped   bool
	Keyword    string
	Zone       zone.Type
	Raw        string
}

// EffectKind identifies an instruction verb recognized in Oracle text.
type EffectKind uint8

// Recognized effect kinds.
const (
	EffectUnknown EffectKind = iota
	EffectAddMana
	EffectAttach
	EffectCast
	EffectCantBeBlocked
	EffectCantBeCountered
	EffectCantBlock
	EffectCounter
	EffectCreate
	EffectDealDamage
	EffectDestroy
	EffectDiscard
	EffectDiscover
	EffectDouble
	EffectDraw
	EffectEnterTapped
	EffectEnterPrepared
	EffectExile
	EffectFight
	EffectGain
	EffectGainControl // gain control of [target permanent]
	EffectGrantKeyword
	EffectInvestigate
	EffectExplore
	EffectLose
	EffectManifest
	EffectManifestDread
	EffectMill
	EffectModifyPT
	EffectMustAttack
	EffectPut
	EffectProliferate
	EffectRegenerate
	EffectReturn
	EffectReveal
	EffectSacrifice
	EffectScry
	EffectSurveil
	EffectSearch
	EffectShuffle
	EffectTap
	EffectUntap
	EffectTransform
)

// DurationKind identifies common continuous-effect durations.
type DurationKind uint8

// Recognized durations.
const (
	DurationNone DurationKind = iota
	DurationUntilEndOfTurn
	DurationUntilYourNextTurn
	DurationThisTurn
	DurationThisCombat
	// DurationForAsLongAsSourceOnBattlefield matches "as long as this [type]
	// remains on the battlefield" and "for as long as this [type] remains on
	// the battlefield".  The effect expires when the source permanent leaves
	// the battlefield.
	DurationForAsLongAsSourceOnBattlefield
	// DurationForAsLongAsYouControlSource matches "for as long as you control
	// [source name]" or "for as long as you control this [type]".  The effect
	// expires when the effect controller no longer controls the source, or
	// when the source leaves the battlefield.
	DurationForAsLongAsYouControlSource
)

// StaticSubjectKind identifies the group affected by a static continuous effect.
type StaticSubjectKind uint8

// Recognized static-effect subject kinds.
const (
	StaticSubjectNone StaticSubjectKind = iota
	StaticSubjectAttachedObject
	StaticSubjectControlledCreatures
	StaticSubjectOtherControlledCreatures
	StaticSubjectControlledWalls
	StaticSubjectControlledArtifacts
	StaticSubjectControlledTokens
	StaticSubjectOpponentControlledCreatures
	StaticSubjectControlledCreatureSubtype
	StaticSubjectOtherControlledCreatureSubtype
)

// CompiledEffect is one recognized instruction verb and the sentence containing
// it. Multiple effects may refer to the same sentence when instructions are
// coordinated.
type CompiledEffect struct {
	Kind              EffectKind
	Span              Span
	Text              string
	VerbSpan          Span
	Duration          DurationKind
	DelayedTiming     game.DelayedTriggerTiming
	Selector          CompiledSelector
	Amount            CompiledAmount
	PowerDelta        CompiledSignedAmount
	ToughnessDelta    CompiledSignedAmount
	StaticSubject     StaticSubjectKind
	StaticSubjectSpan Span
	// StaticSubjectSubtype preserves the printed plural creature subtype for
	// validation and canonicalization by the executable lowering stage.
	StaticSubjectSubtype string
	Symbol               string
	CounterKind          counter.Kind
	CounterKindKnown     bool
	FromZone             zone.Type
	ToZone               zone.Type
	Negated              bool
}

// CounterKindPlacementSupported reports whether named placement of kind has
// complete runtime semantics in the executable backend.
func CounterKindPlacementSupported(kind counter.Kind) bool {
	switch kind {
	case counter.Stun, counter.Finality:
		return false
	default:
		return kind.Valid()
	}
}

// DynamicAmountKind identifies a rules-derived effect amount.
type DynamicAmountKind uint8

// Dynamic amount kinds recognized by the semantic compiler.
const (
	DynamicAmountNone DynamicAmountKind = iota
	DynamicAmountCount
	DynamicAmountControllerLife
	DynamicAmountOpponentCount
	DynamicAmountSourcePower
	DynamicAmountBasicLandTypes
)

// DynamicAmountForm identifies the exact Oracle formula used for an amount.
type DynamicAmountForm uint8

// Dynamic amount forms recognized by the semantic compiler.
const (
	DynamicAmountFormNone DynamicAmountForm = iota
	DynamicAmountEqual
	DynamicAmountForEach
	DynamicAmountWhereX
)

// CompiledAmount is a fixed or rules-derived amount recognized in an effect.
type CompiledAmount struct {
	Value         int
	Known         bool
	DynamicKind   DynamicAmountKind
	DynamicForm   DynamicAmountForm
	Multiplier    int
	Selector      CompiledSelector
	ReferenceSpan Span
	Text          string
}

// CompiledSignedAmount is a fixed signed amount recognized in an effect.
type CompiledSignedAmount struct {
	Value    int
	Known    bool
	Negative bool
}

// CompiledKeyword is a recognized keyword ability.
type CompiledKeyword struct {
	Name      string
	Span      Span
	Text      string
	Parameter string
}

// ReferenceKind identifies the exact reference wording recognized before
// antecedent binding.
type ReferenceKind uint8

// Reference kinds.
const (
	ReferenceUnknown ReferenceKind = iota
	ReferenceSelfName
	ReferenceThisObject
	ReferencePronoun
	ReferenceThatObject
)

// ReferenceBinding identifies the intended referent of a reference occurrence.
type ReferenceBinding uint8

// Bound reference kinds.
const (
	ReferenceBindingUnsupported ReferenceBinding = iota
	ReferenceBindingAmbiguous
	ReferenceBindingSource
	ReferenceBindingTarget
	ReferenceBindingEventPermanent
	ReferenceBindingPriorInstructionResult
)

// CompiledReference records a source-spanned reference and its bound referent.
type CompiledReference struct {
	Kind             ReferenceKind
	Span             Span
	Text             string
	Binding          ReferenceBinding
	Occurrence       int
	PriorInstruction int
}
