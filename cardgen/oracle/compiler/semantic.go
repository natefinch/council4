package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Context supplies card facts needed during semantic compilation.
type Context struct{}

// AbilityKind is the semantic category of a compiled ability.
type AbilityKind uint8

// Semantic ability kinds.
const (
	AbilityUnknown AbilityKind = iota
	AbilitySpell
	AbilityActivated
	AbilityLoyalty
	AbilityChapter
	AbilityTriggered
	AbilityReplacement
	AbilityStatic
	AbilityReminder
	AbilitySpellAdditionalCost
)

var abilityKindNames = [...]string{
	AbilityUnknown:             "unknown",
	AbilitySpell:               "spell",
	AbilityActivated:           "activated",
	AbilityLoyalty:             "loyalty",
	AbilityChapter:             "chapter",
	AbilityTriggered:           "triggered",
	AbilityReplacement:         "replacement",
	AbilityStatic:              "static",
	AbilityReminder:            "reminder",
	AbilitySpellAdditionalCost: "spell additional cost",
}

func (k AbilityKind) String() string {
	if int(k) >= len(abilityKindNames) {
		return "unknown"
	}
	return abilityKindNames[k]
}

// Compilation is the semantic result for one card face.
type Compilation struct {
	Syntax    parser.Document
	Abilities []CompiledAbility
}

// CompiledAbility is a source-spanned semantic ability. Shell semantics
// (cost, trigger, timing, chapter numbers) are fields on CompiledAbility;
// the ability's instruction content (targets, conditions, effects, keywords,
// references, modes) lives in Content.
type CompiledAbility struct {
	Kind                 AbilityKind
	Span                 shared.Span
	Text                 string
	ActivationTiming     ActivationTimingKind
	ActivationTimingSpan shared.Span
	ActivationZone       zone.Type
	AbilityWord          string
	Chapters             []int
	ChapterSpan          shared.Span
	Optional             bool
	OptionalSpan         shared.Span
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
	Span       shared.Span
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
	Span    shared.Span
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
	Span       shared.Span
	Text       string
	Components []CostComponent
	// Order is the cost phrase's dense source-order rank, used to test reference
	// containment without byte offsets.
	Order shared.SourceOrder
}

// CostComponent is one comma-separated cost operation.
type CostComponent struct {
	Kind   CostKind
	Span   shared.Span
	Text   string
	Symbol string
	Amount string
	Object string

	AmountValue     int
	AmountKnown     bool
	AmountFromX     bool
	ObjectKind      SelectorKind
	ObjectType      types.Card
	ObjectTypeKnown bool

	// ObjectTypeAlt is a second permanent type accepted by a two-type cost
	// union such as "sacrifice an artifact or creature." ObjectTypeAltKnown
	// reports its presence.
	ObjectTypeAlt      types.Card
	ObjectTypeAltKnown bool

	ObjectSupertype   types.Super
	SupertypeKnown    bool
	ObjectColor       color.Color
	ObjectColorKnown  bool
	ObjectController  ControllerKind
	ObjectNonToken    bool
	PermanentModifier bool
	RequireTapped     bool
	RequireUntapped   bool
	SourceZone        zone.Type
	ToZone            zone.Type
	SourceSelf        bool
	CounterKind       counter.Kind
	CounterKindKnown  bool
	SubtypesAny       []types.Sub

	// ExcludeSource reports that the cost object excludes the ability's own
	// source ("another"), recognized by the parser.
	ExcludeSource bool

	// Order is the component's dense source-order rank, used to test reference
	// containment without byte offsets.
	Order shared.SourceOrder
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
	Span shared.Span
	Text string
	// Event retains the exact event-clause text for diagnostics and source
	// consumption. Executable lowering consumes Pattern instead.
	Event     string
	Pattern   TriggerPattern
	Condition *CompiledCondition
	// MaxTriggersPerTurn caps how many times this ability may trigger each turn,
	// taken from a recognized "This ability triggers only once/twice each turn."
	// qualifier. Zero means the ability may trigger without a per-turn limit.
	MaxTriggersPerTurn int
	// MaxTriggersPerTurnSpan is the source span of the recognized per-turn cap
	// qualifier, consumed by lowering so the clause counts as accounted-for.
	MaxTriggersPerTurnSpan shared.Span
	// Order is the trigger clause's dense source-order rank, used to bind
	// references that fall within the trigger body without byte offsets.
	Order shared.SourceOrder
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
	// ConditionPredicatePriorInstructionAccepted is satisfied when the prior
	// optional instruction was performed ("if you do"). It is the affirmative
	// complement of ConditionPredicatePriorInstructionNotAccepted.
	ConditionPredicatePriorInstructionAccepted
	ConditionPredicateCounterPlacementOnControlledCreature
	ConditionPredicateControllerCounterPlacement
	ConditionPredicateDamageByControlledSource
	ConditionPredicateTokenCreationUnderController
	ConditionPredicateSourceWouldDie
	ConditionPredicateSourceWouldGoToGraveyard
	ConditionPredicateTargetControllerDoesNotPay
	ConditionPredicateObjectMatches
	ConditionPredicateObjectExists
	ConditionPredicateEventSubjectHadCounters
	// ConditionPredicateEventHistory is satisfied when the event history for the
	// current or previous turn contains at least one event matching
	// EventHistoryPattern. When Negated is true the condition is satisfied when
	// no matching event is found (e.g. "if no spells were cast last turn").
	ConditionPredicateEventHistory
)

// ConditionEventHistoryWindow identifies which turn's event log to search.
type ConditionEventHistoryWindow uint8

// Condition event history window values.
const (
	// ConditionEventHistoryWindowCurrentTurn checks events that occurred during
	// the current turn (e.g. "if you attacked this turn").
	ConditionEventHistoryWindowCurrentTurn ConditionEventHistoryWindow = iota
	// ConditionEventHistoryWindowPreviousTurn checks events from the immediately
	// preceding turn (e.g. "if an opponent lost life last turn").
	ConditionEventHistoryWindowPreviousTurn
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

// ConditionTriState is a closed semantic true/false selection filter.
type ConditionTriState uint8

// Condition tri-state values.
const (
	ConditionTriAny ConditionTriState = iota
	ConditionTriTrue
	ConditionTriFalse
)

// ConditionSelection is the source-independent Selection vocabulary used by
// semantic conditions. Subtype names are canonicalized during recognition.
type ConditionSelection struct {
	RequiredTypes     []ConditionCardType
	Supertypes        []ConditionSupertype
	SubtypesAny       []string
	ColorsAny         []ConditionColor
	Colorless         bool
	Multicolored      bool
	TokenOnly         bool
	ExcludeSource     bool
	Tapped            ConditionTriState
	PowerAtLeast      int
	MatchPowerAtLeast bool
	// TotalPowerAtLeast is the collective-power threshold for a "have total
	// power <n> or greater" qualifier. MatchTotalPowerAtLeast marks it present.
	TotalPowerAtLeast      int
	MatchTotalPowerAtLeast bool
}

// CompiledCondition is a closed, source-spanned semantic condition.
type CompiledCondition struct {
	Kind          ConditionKind
	Span          shared.Span
	Text          string
	Intervening   bool
	Predicate     ConditionPredicate
	Negated       bool
	Threshold     int
	Selection     ConditionSelection
	Counter       ConditionCounter
	ObjectBinding ReferenceBinding

	// NodeID is the parser-assigned identity of this condition's boundary. A
	// triggered ability's intervening condition shares a NodeID with its content
	// condition, so the compiler links them by identity instead of span equality.
	NodeID int
	// ClauseIndex and EventHistoryIndex are the parser-resolved indices of the
	// typed clause and event-history condition that fill this condition's span,
	// or -1 when none does. The compiler reads the matching clause by index
	// instead of scanning for an equal span.
	ClauseIndex       int
	EventHistoryIndex int

	// SubjectSpan is the source span of the subject noun phrase for the
	// source-death predicates (ConditionPredicateSourceWouldDie and
	// ConditionPredicateSourceWouldGoToGraveyard). Reference binding confirms a
	// typed source reference fills that span; the compiler never re-derives the
	// subject from condition text.
	SubjectSpan shared.Span

	// SubjectRefID is the parser-assigned NodeID of the reference that fills the
	// subject span for the source-death predicates, or -1 when none does. The
	// compiler confirms the subject binds the source by matching this identity
	// rather than comparing the reference span to the subject span.
	SubjectRefID int

	// ActivationKeywordSpan is the source span of an "Activate" keyword that
	// introduces an "Activate only if ..." restriction. It is the zero span when
	// absent. The parser recognizes the keyword and reports it on the condition
	// boundary; lowering consumes this span for exact source accounting without
	// inspecting token spelling.
	ActivationKeywordSpan shared.Span

	// EventHistoryPattern and EventHistoryWindow are set when Predicate is
	// ConditionPredicateEventHistory. EventHistoryPattern describes the event
	// kind and optional filters; EventHistoryWindow selects the turn to search.
	// EventHistoryPattern is a pointer to avoid bloating CompiledCondition.
	EventHistoryPattern *TriggerPattern
	EventHistoryWindow  ConditionEventHistoryWindow

	// Order is the condition's dense source-order rank. The compiler tests
	// whether a reference or payment falls within the condition by comparing
	// these ranks instead of inspecting byte offsets.
	Order shared.SourceOrder
}

// TargetCardinality is an inclusive target count range.
type TargetCardinality struct {
	Min int
	Max int
}

// CompiledTarget is one occurrence of the word "target" and its local noun
// phrase.
type CompiledTarget struct {
	Span        shared.Span
	Text        string
	Cardinality TargetCardinality
	Selector    CompiledSelector
	Exact       bool
	// Order is the target's dense source-order rank, used to bind a reference to
	// its closest preceding target without byte offsets.
	Order shared.SourceOrder
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
	SelectorTriggeredAbilityOrSpell
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
	All        bool
	Another    bool
	Other      bool
	Attacking  bool
	Blocking   bool
	Tapped     bool
	Untapped   bool
	Keyword    parser.KeywordKind
	// ExcludedKeyword records a "without <keyword>" selector qualifier (e.g.
	// "each creature without flying"); it is mutually exclusive with Keyword.
	ExcludedKeyword parser.KeywordKind
	Zone            zone.Type
	ManaValue       compare.Int
	MatchManaValue  bool
	Power           compare.Int
	MatchPower      bool
	Toughness       compare.Int
	MatchToughness  bool
	Colorless       bool
	Multicolored    bool
	// PlayerOrPlaneswalker marks the combined "player or planeswalker" /
	// "opponent or planeswalker" combined damage target. Kind stays
	// SelectorPlayer or SelectorOpponent; this flag records the additional
	// planeswalker-permanent half the merged Kind cannot express.
	PlayerOrPlaneswalker bool
	atoms                *CompiledSelectorAtoms
}

// CompiledSelectorAtoms holds parser-owned atom-derived selector filters that
// are commonly empty. Keeping them behind one pointer avoids copying several
// slices with every selector, effect, and amount value.
type CompiledSelectorAtoms struct {
	RequiredTypesAny   []types.Card
	ExcludedTypes      []types.Card
	Supertypes         []types.Super
	ExcludedSupertypes []types.Super
	ColorsAny          []color.Color
	ExcludedColors     []color.Color
	SubtypesAny        []types.Sub
	SourceTypes        []types.Card
}

// Supertypes returns supertype filters accepted by this selector.
func (s CompiledSelector) Supertypes() []types.Super {
	return selectorAtoms(s).Supertypes
}

func appendSelectorSupertype(selector *CompiledSelector, supertype types.Super) {
	atoms := mutableSelectorAtoms(selector)
	atoms.Supertypes = append(atoms.Supertypes, supertype)
}

// ExcludedSupertypes returns supertype filters excluded from this selector (a
// "nonbasic" / "nonlegendary" filter).
func (s CompiledSelector) ExcludedSupertypes() []types.Super {
	return selectorAtoms(s).ExcludedSupertypes
}

func appendSelectorExcludedSupertype(selector *CompiledSelector, supertype types.Super) {
	atoms := mutableSelectorAtoms(selector)
	atoms.ExcludedSupertypes = append(atoms.ExcludedSupertypes, supertype)
}

func selectorAtoms(s CompiledSelector) CompiledSelectorAtoms {
	if s.atoms == nil {
		return CompiledSelectorAtoms{}
	}
	return *s.atoms
}

func mutableSelectorAtoms(s *CompiledSelector) *CompiledSelectorAtoms {
	if s.atoms == nil {
		s.atoms = &CompiledSelectorAtoms{}
	}
	return s.atoms
}

// RequiredTypesAny returns the required card-type filters for this selector.
func (s CompiledSelector) RequiredTypesAny() []types.Card {
	return selectorAtoms(s).RequiredTypesAny
}

// SourceTypes returns the card types a stack-object target requires of the
// targeted object's source, modeling "from an artifact source" restrictions.
func (s CompiledSelector) SourceTypes() []types.Card {
	return selectorAtoms(s).SourceTypes
}

func appendSelectorSourceType(selector *CompiledSelector, cardType types.Card) {
	atoms := mutableSelectorAtoms(selector)
	atoms.SourceTypes = append(atoms.SourceTypes, cardType)
}

func setSelectorRequiredTypesAny(selector *CompiledSelector, typesAny []types.Card) {
	mutableSelectorAtoms(selector).RequiredTypesAny = typesAny
}

// ExcludedTypes returns card-type filters excluded from this selector.
func (s CompiledSelector) ExcludedTypes() []types.Card {
	return selectorAtoms(s).ExcludedTypes
}

func appendSelectorExcludedType(selector *CompiledSelector, cardType types.Card) {
	atoms := mutableSelectorAtoms(selector)
	atoms.ExcludedTypes = append(atoms.ExcludedTypes, cardType)
}

// ColorsAny returns color filters accepted by this selector.
func (s CompiledSelector) ColorsAny() []color.Color {
	return selectorAtoms(s).ColorsAny
}

func appendSelectorColorAny(selector *CompiledSelector, colorValue color.Color) {
	atoms := mutableSelectorAtoms(selector)
	atoms.ColorsAny = append(atoms.ColorsAny, colorValue)
}

// ExcludedColors returns color filters excluded from this selector.
func (s CompiledSelector) ExcludedColors() []color.Color {
	return selectorAtoms(s).ExcludedColors
}

func appendSelectorExcludedColor(selector *CompiledSelector, colorValue color.Color) {
	atoms := mutableSelectorAtoms(selector)
	atoms.ExcludedColors = append(atoms.ExcludedColors, colorValue)
}

// SubtypesAny returns subtype filters accepted by this selector.
func (s CompiledSelector) SubtypesAny() []types.Sub {
	return selectorAtoms(s).SubtypesAny
}

func appendSelectorSubtypesAny(selector *CompiledSelector, subtypes ...types.Sub) {
	atoms := mutableSelectorAtoms(selector)
	atoms.SubtypesAny = append(atoms.SubtypesAny, subtypes...)
}

// EffectKind identifies an instruction verb recognized in Oracle text.
type EffectKind uint8

// Recognized effect kinds.
const (
	EffectUnknown EffectKind = iota
	EffectAddMana
	EffectAttach
	EffectCast
	EffectCantAttack
	EffectCantBeBlocked
	EffectCantBeBlockedByCreaturesWith
	EffectCantBeBlockedByMoreThanOne
	EffectCantBeCountered
	EffectCantBlock
	EffectCantAttackOrBlock
	EffectDoesntUntap
	EffectCounter
	EffectCreate
	EffectDealDamage
	EffectDestroy
	EffectDig
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
	EffectMustBeBlocked
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
	// DurationForAsLongAsControlledCreatureEnchanted matches the
	// attachment-dependent wording "for as long as that creature is enchanted".
	// The effect expires when the affected creature is no longer enchanted or
	// leaves the battlefield.
	DurationForAsLongAsControlledCreatureEnchanted
)

// StaticSubjectKind identifies the group affected by a static continuous effect.
type StaticSubjectKind uint8

// Recognized static-effect subject kinds.
const (
	StaticSubjectNone StaticSubjectKind = iota
	StaticSubjectAttachedObject
	StaticSubjectAllCreatures
	StaticSubjectAllOtherCreatures
	StaticSubjectAttackingCreatures
	StaticSubjectBlockingCreatures
	StaticSubjectControlledCreatures
	StaticSubjectOtherControlledCreatures
	StaticSubjectControlledWalls
	StaticSubjectControlledArtifacts
	StaticSubjectControlledTokens
	StaticSubjectOpponentControlledCreatures
	StaticSubjectControlledCreatureSubtype
	StaticSubjectOtherControlledCreatureSubtype
	StaticSubjectAllCreatureSubtype
	StaticSubjectOtherCreatureSubtype
	StaticSubjectControlledAttackingCreatures
	StaticSubjectControlledCreatureTokens
	StaticSubjectBattlefieldCreatureTokens
	StaticSubjectControlledLegendaryCreatures
	StaticSubjectControlledUntappedCreatures
	StaticSubjectOtherControlledTappedCreatures
)

// CompiledEffect is one recognized instruction verb and the sentence containing
// it. Multiple effects may refer to the same sentence when instructions are
// coordinated.
type CompiledEffect struct {
	Kind              EffectKind
	Context           parser.EffectContextKind
	Connection        parser.EffectConnectionKind
	ConnectionSpan    shared.Span
	Span              shared.Span
	ClauseSpan        shared.Span
	Text              string
	VerbSpan          shared.Span
	References        []CompiledReference
	SubjectReferences []CompiledReference
	Targets           []CompiledTarget
	SubjectTargets    []CompiledTarget
	Duration          DurationKind
	DelayedTiming     game.DelayedTriggerTiming
	Selector          CompiledSelector
	// DamageRecipientSelectors holds the compiled recipient groups of a
	// dual-recipient fixed group-damage effect ("deals N damage to each X and
	// each Y"). It is empty for single-recipient damage; when present it has
	// exactly two entries that lowering damages in Oracle order.
	DamageRecipientSelectors []CompiledSelector
	// DamageRecipientReference marks a damage recipient that is the controller or
	// owner of a referenced object (the prior removal target), as in "deals 2
	// damage to that land's controller". It is None for every other recipient.
	DamageRecipientReference parser.DamageRecipientReferenceKind
	// HasSelfDamageRider reports a "... and N damage to you" rider on a
	// single-target deal-damage clause ("deals A damage to any target and B
	// damage to you"). SelfDamageRiderValue holds the fixed self-damage amount B
	// dealt to the source's own controller; lowering emits a second Damage
	// instruction after the primary target damage.
	HasSelfDamageRider   bool
	SelfDamageRiderValue int
	Amount               CompiledAmount
	PowerDelta           CompiledSignedAmount
	ToughnessDelta       CompiledSignedAmount
	TokenPower           int
	TokenToughness       int
	TokenPTKnown         bool
	TokenCopyOfTarget    bool
	StaticSubject        StaticSubjectKind
	StaticSubjectSpan    shared.Span
	Details              *CompiledEffectDetails
	CounterKind          counter.Kind
	CounterKindKnown     bool
	// CounterRecipientAttached reports that a counter-placement effect places its
	// counters on the permanent the source Aura is attached to ("... on enchanted
	// creature"). Lowering routes it to the runtime's source attached-permanent
	// reference; it is false for every other recipient.
	CounterRecipientAttached bool
	FromZone                 zone.Type
	ToZone                   zone.Type
	Destination              parser.EffectDestinationPosition
	EntersTapped             bool
	EntersTappedSelf         bool
	EntersColorChoice        bool
	EntersColorChoiceExclude mana.Color
	EntersTypeChoice         bool
	EntersWithCounters       bool
	UnderYourControl         bool
	CastAsAdventure          bool
	Negated                  bool
	Optional                 bool
	Divided                  bool
	OptionalSpan             shared.Span
	Mana                     CompiledEffectMana
	Replacement              parser.EffectReplacementSyntax
	Payment                  CompiledEffectPayment
	Exact                    bool
	RequiresOrderedLowering  bool
	HasUnrecognizedSibling   bool
	UnsupportedDetail        string
	// Order is the effect's dense source-order rank (of Span); VerbOrder is the
	// rank of VerbSpan. The compiler compares these ranks to order effects and
	// bind references relative to effect verbs without inspecting byte offsets.
	Order     shared.SourceOrder
	VerbOrder shared.SourceOrder
	// LifeObject reports that a gain/lose effect's object is the player's life
	// rather than a keyword or quoted ability. Consumers route only true life
	// changes to the life lowerer.
	LifeObject bool
	// PreventRegeneration reports a destroy effect carrying a "can't be
	// regenerated" rider. RegenerationRiderSpan covers the rider sentence so
	// lowering can credit its tokens toward source coverage.
	PreventRegeneration   bool
	RegenerationRiderSpan shared.Span
	// Dig carries the impulse put clause's structured fields from the parser so
	// the combined dig lowerer can pair an EffectDig look with its EffectPut put.
	Dig parser.DigSyntax
}

// CompiledEffectMana describes exact typed add-mana output.
type CompiledEffectMana struct {
	Span                  shared.Span
	Symbols               []string
	Colors                []mana.Color
	ColorsKnown           bool
	Choice                bool
	AnyColor              bool
	ChosenColor           bool
	ChosenColorFixed      mana.Color
	ChosenColorFixedKnown bool
	CommanderIdentity     bool
	LegacyBodyExact       bool
}

// CompiledEffectPayment is a typed resolution payment embedded in an effect.
type CompiledEffectPayment struct {
	Span     shared.Span
	Payer    parser.EffectPaymentPayerKind
	ManaCost cost.Mana
	// Order is the payment's dense source-order rank, used to test condition
	// containment without byte offsets.
	Order shared.SourceOrder
}

// CompiledEffectDetails holds rarely-used effect details outside the hot effect
// value copied during instruction scans.
type CompiledEffectDetails struct {
	StaticSubjectType   *CompiledStaticSubjectType
	StaticSubjectColors *CompiledStaticSubjectColors
	Symbol              string
}

// CompiledStaticSubjectType preserves a static subject's printed subtype and its
// parser-resolved canonical subtype when known.
type CompiledStaticSubjectType struct {
	Text  string
	Sub   types.Sub
	Known bool
}

// CompiledStaticSubjectColors preserves a static subject's optional color filter:
// the single colors matched disjunctively and the colorless/multicolored
// color-family qualifiers.
type CompiledStaticSubjectColors struct {
	ColorsAny    []parser.Color
	Colorless    bool
	Multicolored bool
}

func staticSubjectType(text string, sub types.Sub, known bool) *CompiledStaticSubjectType {
	if text == "" && !known {
		return nil
	}
	return &CompiledStaticSubjectType{Text: text, Sub: sub, Known: known}
}

func staticSubjectColors(colors []parser.Color, colorless, multicolored bool) *CompiledStaticSubjectColors {
	if len(colors) == 0 && !colorless && !multicolored {
		return nil
	}
	return &CompiledStaticSubjectColors{ColorsAny: colors, Colorless: colorless, Multicolored: multicolored}
}

func compiledEffectDetails(staticType *CompiledStaticSubjectType, staticColors *CompiledStaticSubjectColors, symbol string) *CompiledEffectDetails {
	if staticType == nil && staticColors == nil && symbol == "" {
		return nil
	}
	return &CompiledEffectDetails{StaticSubjectType: staticType, StaticSubjectColors: staticColors, Symbol: symbol}
}

// StaticSubjectSubtype returns the printed subtype text on a static subject.
func (e *CompiledEffect) StaticSubjectSubtype() string {
	if e.Details == nil || e.Details.StaticSubjectType == nil {
		return ""
	}
	return e.Details.StaticSubjectType.Text
}

// StaticSubjectSub returns the parser-resolved static subject subtype.
func (e *CompiledEffect) StaticSubjectSub() types.Sub {
	if e.Details == nil || e.Details.StaticSubjectType == nil {
		return ""
	}
	return e.Details.StaticSubjectType.Sub
}

// StaticSubjectSubKnown reports whether the static subject subtype was resolved.
func (e *CompiledEffect) StaticSubjectSubKnown() bool {
	return e.Details != nil && e.Details.StaticSubjectType != nil && e.Details.StaticSubjectType.Known
}

// StaticSubjectColorsAny returns the static subject's any-of color filter.
func (e *CompiledEffect) StaticSubjectColorsAny() []parser.Color {
	if e.Details == nil || e.Details.StaticSubjectColors == nil {
		return nil
	}
	return e.Details.StaticSubjectColors.ColorsAny
}

// StaticSubjectColorless reports whether the static subject requires colorless.
func (e *CompiledEffect) StaticSubjectColorless() bool {
	return e.Details != nil && e.Details.StaticSubjectColors != nil && e.Details.StaticSubjectColors.Colorless
}

// StaticSubjectMulticolored reports whether the static subject requires
// multicolored.
func (e *CompiledEffect) StaticSubjectMulticolored() bool {
	return e.Details != nil && e.Details.StaticSubjectColors != nil && e.Details.StaticSubjectColors.Multicolored
}

// StaticSubjectHasColorFilter reports whether the static subject carries any
// color constraint.
func (e *CompiledEffect) StaticSubjectHasColorFilter() bool {
	return e.Details != nil && e.Details.StaticSubjectColors != nil
}

// Symbol returns the first mana symbol recognized in this effect.
func (e *CompiledEffect) Symbol() string {
	if e.Details == nil {
		return ""
	}
	return e.Details.Symbol
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
	DynamicAmountEventCardCount
	DynamicAmountLifeLostThisWay
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
	VariableX     bool
	DynamicKind   DynamicAmountKind
	DynamicForm   DynamicAmountForm
	Multiplier    int
	ReferenceSpan shared.Span
	Text          string
	selector      *CompiledSelector
}

// Selector returns the amount's dynamic count subject selector, when present.
func (a CompiledAmount) Selector() CompiledSelector {
	if a.selector == nil {
		return CompiledSelector{}
	}
	return *a.selector
}

// CompiledSignedAmount is a fixed signed amount recognized in an effect.
type CompiledSignedAmount struct {
	Value    int
	Known    bool
	Negative bool
	// VariableX marks a power/toughness side written as the variable "X", whose
	// magnitude is supplied by the effect's dynamic amount.
	VariableX bool
}

// CompiledKeyword is a recognized keyword ability.
type CompiledKeyword struct {
	Kind            parser.KeywordKind
	Name            string
	Span            shared.Span
	Text            string
	Parameter       string
	ParameterKind   parser.KeywordParameterKind
	ManaCost        cost.Mana
	Integer         int
	EnchantTarget   parser.ObjectNoun
	Protection      game.ProtectionKeyword
	ProtectionKnown bool
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
	ReferenceThatPlayer
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
	ReferenceBindingEventCard
	ReferenceBindingPriorInstructionResult
	// ReferenceBindingEventPlayer binds player pronouns (they/their/them) in
	// trigger bodies where the triggering event has an authoritative player
	// subject. At runtime EventPlayerReference() resolves this to the player
	// identified by the event.
	ReferenceBindingEventPlayer
)

// CompiledReference records a source-spanned reference and its bound referent.
type CompiledReference struct {
	Kind             ReferenceKind
	Pronoun          ReferencePronounKind
	Span             shared.Span
	Text             string
	Binding          ReferenceBinding
	Occurrence       int
	PriorInstruction int
	// NodeID is the parser-assigned stable identity of this reference within its
	// ability or mode. Distinct copies of the same source reference share a
	// NodeID, so the compiler matches references by identity instead of span
	// equality.
	NodeID int
	// Order is the reference's dense source-order rank. The compiler compares
	// these ranks to order references against effects, targets, the trigger, and
	// one another, and to test cost/component/effect/condition containment,
	// instead of inspecting byte offsets.
	Order shared.SourceOrder
}

// ReferencePronounKind identifies the grammatical pronoun carried by a
// compiled reference.
type ReferencePronounKind uint8

// Compiled reference pronouns.
const (
	ReferencePronounUnknown ReferencePronounKind = iota
	ReferencePronounIt
	ReferencePronounIts
	ReferencePronounThey
	ReferencePronounTheir
	ReferencePronounThem
	ReferencePronounThose
)
