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

// CompiledAbility is a source-spanned semantic ability.
type CompiledAbility struct {
	Kind                 AbilityKind
	Span                 Span
	Text                 string
	ActivationTiming     ActivationTimingKind
	ActivationTimingSpan Span
	AbilityWord          string
	Chapters             []int
	ChapterSpan          Span
	Optional             bool
	OptionalSpan         Span
	Cost                 *CompiledCost
	Trigger              *CompiledTrigger
	Modes                []CompiledMode
	Targets              []CompiledTarget
	Conditions           []CompiledCondition
	Effects              []CompiledEffect
	Keywords             []CompiledKeyword
	References           []CompiledReference
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
)

// CompiledMode is one semantic option in a modal ability.
type CompiledMode struct {
	Span       Span
	Text       string
	Targets    []CompiledTarget
	Conditions []CompiledCondition
	Effects    []CompiledEffect
	Keywords   []CompiledKeyword
	References []CompiledReference
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
	Kind      TriggerKind
	Span      Span
	Text      string
	Event     string
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

// CompiledCondition preserves a condition and whether it is an intervening-if
// trigger condition.
type CompiledCondition struct {
	Kind        ConditionKind
	Span        Span
	Text        string
	Intervening bool
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

// ReferenceKind identifies references requiring semantic binding.
type ReferenceKind uint8

// Reference kinds.
const (
	ReferenceUnknown ReferenceKind = iota
	ReferenceSelfName
	ReferenceThisObject
	ReferencePronoun
	ReferenceThatObject
)

// CompiledReference records a self-reference or antecedent-dependent phrase.
type CompiledReference struct {
	Kind ReferenceKind
	Span Span
	Text string
}
