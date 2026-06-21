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
	AbilitySpellAlternativeCost
)

var abilityKindNames = [...]string{
	AbilityUnknown:              "unknown",
	AbilitySpell:                "spell",
	AbilityActivated:            "activated",
	AbilityLoyalty:              "loyalty",
	AbilityChapter:              "chapter",
	AbilityTriggered:            "triggered",
	AbilityReplacement:          "replacement",
	AbilityStatic:               "static",
	AbilityReminder:             "reminder",
	AbilitySpellAdditionalCost:  "spell additional cost",
	AbilitySpellAlternativeCost: "spell alternative cost",
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
	Kind     AbilityKind
	Optional bool
	// ExactSequence is a parser-recognized exact multi-instruction resolving
	// body. When set, the normal target/condition/effect content is empty and
	// lowering emits the fixed instruction template for the kind. It is declared
	// next to Optional so the byte packs into existing alignment padding.
	ExactSequence              ExactSequenceKind
	Span                       shared.Span
	Text                       string
	ActivationTiming           ActivationTimingKind
	ActivationTimingSpan       shared.Span
	ActivationZone             zone.Type
	AbilityWord                string
	Chapters                   []int
	ChapterSpan                shared.Span
	OptionalSpan               shared.Span
	Cost                       *CompiledCost
	SourceAbilityCostReduction *CompiledSourceAbilityCostReduction
	AlternativeCost            *CompiledAlternativeCost
	Trigger                    *CompiledTrigger
	Content                    AbilityContent
	Static                     *CompiledStaticSemantics
}

// CompiledSourceAbilityCostReduction describes a source-local activated-ability
// cost reduction derived from typed Oracle syntax.
type CompiledSourceAbilityCostReduction struct {
	Span           shared.Span
	Amount         int
	CountSelection CompiledSelector
}

// AlternativeCostCondition identifies a runtime condition on an alternative spell cost.
type AlternativeCostCondition uint8

// Supported alternative spell-cost conditions.
const (
	AlternativeCostConditionUnknown AlternativeCostCondition = iota
	AlternativeCostConditionControlsCommander
	// AlternativeCostConditionNotYourTurn gates a pitch alternative cost behind
	// "If it's not your turn,".
	AlternativeCostConditionNotYourTurn
)

// AlternativeCostKind identifies the semantic rules change attached to an
// alternative spell cost.
type AlternativeCostKind uint8

// Supported alternative spell-cost kinds.
const (
	AlternativeCostUnknown AlternativeCostKind = iota
	AlternativeCostCommander
	AlternativeCostOverload
	// AlternativeCostPitch is the Force of Will family: exile a colored card
	// from hand (optionally paying extra life) instead of paying mana.
	AlternativeCostPitch
	// AlternativeCostFlashback is the alternative-cost form of Flashback: cast
	// the spell from the graveyard by paying the cost carried on the ability's
	// CompiledCost, then exile it.
	AlternativeCostFlashback
)

// CompiledAlternativeCost is text-independent semantic data for an optional
// replacement of a spell's printed mana cost.
type CompiledAlternativeCost struct {
	Kind                  AlternativeCostKind
	Condition             AlternativeCostCondition
	WithoutPayingManaCost bool
	ManaCost              cost.Mana
	ReplaceTargetWithEach bool

	// PitchColor is the color of the card exiled from hand by a pitch cost.
	PitchColor color.Color
	// PitchColorKnown reports whether PitchColor was recognized.
	PitchColorKnown bool
	// PitchCount is the number of cards exiled from hand by a pitch cost.
	PitchCount int
	// PitchLife is additional life paid alongside a pitch cost, or zero.
	PitchLife int
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
	ActivationTimingDuringYourTurn
	ActivationTimingUnsupported
	// ActivationTimingInstant marks an explicit instant-speed restriction
	// ("Activate only as an instant"), which is the default timing for an
	// activated ability and lowers to no runtime restriction.
	ActivationTimingInstant
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

// ModeChoiceBonusCondition identifies a cast-time modal bonus condition.
type ModeChoiceBonusCondition uint8

const (
	// ModeChoiceBonusConditionNone marks content without a modal bonus.
	ModeChoiceBonusConditionNone ModeChoiceBonusCondition = iota
	// ModeChoiceBonusConditionControlsCommander requires controlling a commander.
	ModeChoiceBonusConditionControlsCommander
)

// CompiledModeChoiceBonus is a text-independent modal range expansion.
type CompiledModeChoiceBonus struct {
	Condition          ModeChoiceBonusCondition
	AdditionalMaxModes int
}

// CompiledModalSemantics holds a modal choice range and conditional bonus.
type CompiledModalSemantics struct {
	MinModes int
	MaxModes int
	Kind     CompiledModalChoiceKind
	Bonus    CompiledModeChoiceBonus
}

// CompiledModalChoiceKind identifies exact typed modal header vocabulary.
type CompiledModalChoiceKind uint8

// Compiled modal choice kinds.
const (
	CompiledModalChoiceUnknown CompiledModalChoiceKind = iota
	CompiledModalChoiceOneOrMore
)

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

// CompiledModeLabel identifies an exact typed mode label.
type CompiledModeLabel uint8

// Compiled mode labels.
const (
	CompiledModeLabelNone CompiledModeLabel = iota
	CompiledModeLabelSellContraband
	CompiledModeLabelBuyInformation
	CompiledModeLabelHireMercenary
)

// CompiledMode is one semantic option in a modal ability.
type CompiledMode struct {
	Span    shared.Span
	Text    string
	Label   CompiledModeLabel
	Content AbilityContent
	// Modal is populated only on the first mode of a modal ability.
	Modal *CompiledModalSemantics
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

	// DiscardWholeHand reports a "discard your hand" cost object, recognized by
	// the parser. The payer discards every card in their hand.
	DiscardWholeHand bool

	// ChoiceGroup tags this component as one alternative of a printed "<cost> or
	// <cost>" choice. Zero means a mandatory standalone cost; components sharing
	// a nonzero value are alternatives of which exactly one is paid.
	ChoiceGroup uint8

	// PayLifeAmountDynamic names a rules-derived amount for a "pay life equal
	// to ..." cost whose value is neither fixed nor X. DynamicAmountNone means
	// the life amount is a fixed value or X.
	PayLifeAmountDynamic DynamicAmountKind

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
	ConditionPredicateEventPlayerDoesNotPay
	ConditionPredicateObjectMatches
	ConditionPredicateObjectExists
	ConditionPredicateEventSubjectHadCounters
	// ConditionPredicateEventHistory is satisfied when the event history for the
	// current or previous turn contains at least one event matching
	// EventHistoryPattern. When Negated is true the condition is satisfied when
	// no matching event is found (e.g. "if no spells were cast last turn").
	ConditionPredicateEventHistory
	// ConditionPredicateAnyOpponentPoisonAtLeast is satisfied when at least one
	// opponent has at least Threshold poison counters.
	ConditionPredicateAnyOpponentPoisonAtLeast
	// ConditionPredicateControllerHandSizeExactly is satisfied when the
	// controller's hand holds exactly Threshold cards.
	ConditionPredicateControllerHandSizeExactly
	// ConditionPredicateControllerCreatedTokenThisTurn is satisfied when the
	// controller created at least one token during the current turn ("Activate
	// only if you created a token this turn").
	ConditionPredicateControllerCreatedTokenThisTurn
	// ConditionPredicateCounterPlacementOnControlledPermanent is satisfied when
	// one or more counters would be put on a permanent the controller controls,
	// as in Doubling Season's counter clause. Counter is optional: when set the
	// replacement is restricted to that counter kind, otherwise it applies to
	// every counter kind.
	ConditionPredicateCounterPlacementOnControlledPermanent
	// ConditionPredicateControllerWouldCreateNamedToken is satisfied when the
	// controller would create a token matching a named-token replacement set, as
	// in Academy Manufactor's "If you would create a Clue, Food, or Treasure
	// token, instead create one of each." The replaced token types come from the
	// owning create effect's selector.
	ConditionPredicateControllerWouldCreateNamedToken
	// ConditionPredicateControlComparison is satisfied when one player scope's
	// count of permanents matching Selection compares (greater/less) against
	// another scope's count ("if an opponent controls more lands than you").
	ConditionPredicateControlComparison
	// ConditionPredicateEventSubjectNameUnique is satisfied when the triggering
	// event permanent's name differs from every other creature its controller
	// controls and every creature card in their graveyard ("if it doesn't have
	// the same name as another creature you control or a creature card in your
	// graveyard", Guardian Project).
	ConditionPredicateEventSubjectNameUnique
	// ConditionPredicateTargetColor is satisfied when the effect's chosen target
	// has the recognized color ("if it's blue" on Pyroblast / Red Elemental
	// Blast). The color filter lives in Selection.ColorsAny; counter/destroy
	// lowering binds the predicate to the effect's target object.
	ConditionPredicateTargetColor
	// ConditionPredicateWouldDrawFromEmptyLibrary is satisfied when the
	// controller would draw a card while their library is empty ("if you would
	// draw a card while your library has no cards in it"). It gates the
	// draw-from-empty-library win replacement (Laboratory Maniac).
	ConditionPredicateWouldDrawFromEmptyLibrary
	// ConditionPredicateCastDuringControllerMainPhase is satisfied when the
	// resolving spell was cast during its controller's main phase ("Addendum —
	// If you cast this spell during your main phase, ...").
	ConditionPredicateCastDuringControllerMainPhase
	// ConditionPredicateWouldDrawCard is satisfied when the controller would draw
	// a card ("if you would draw a card"). It gates the draw-doubling replacement
	// (Thought Reflection).
	ConditionPredicateWouldDrawCard
	// ConditionPredicateWouldDrawCardExceptFirstInDrawStep is satisfied when the
	// controller would draw a card other than the first one they draw in each of
	// their draw steps ("if you would draw a card except the first one you draw
	// in each of your draw steps"). It gates the draw-doubling replacement whose
	// draw-step draw is exempt (Teferi's Ageless Insight).
	ConditionPredicateWouldDrawCardExceptFirstInDrawStep
	// ConditionPredicateCardWouldGoToGraveyard is satisfied when a card (or
	// permanent) would be put into a watched graveyard. It gates the continuous
	// graveyard-redirect replacement "If a card would be put into [a/your/an
	// opponent's] graveyard from anywhere, exile it instead." (Leyline of the
	// Void, Samurai of the Pale Curtain, Dryad Militant, Rest in Peace). The
	// watched scope, card-type filter, and battlefield-only restriction live in
	// the condition's Graveyard* fields.
	ConditionPredicateCardWouldGoToGraveyard
	// ConditionPredicateControllerLifeGain is satisfied when the controller would
	// gain life ("if you would gain life"). It gates the life-gain replacement
	// "you gain twice that much life instead." / "you gain that much life plus N
	// instead." (Boon Reflection, Angel of Vitality).
	ConditionPredicateControllerLifeGain
	// ConditionPredicateTokenCreationAnyController is satisfied when one or more
	// tokens would be created under any player's control ("If one or more tokens
	// would be created, twice that many of those tokens are created instead.",
	// Primal Vigor, Selesnya Loft Gardens). It is the any-player counterpart of
	// ConditionPredicateTokenCreationUnderController.
	ConditionPredicateTokenCreationAnyController
	// ConditionPredicateCounterPlacementOnAnyCreature is satisfied when one or
	// more counters would be put on any creature, regardless of its controller
	// ("If one or more +1/+1 counters would be put on a creature, twice that many
	// +1/+1 counters are put on that creature instead.", Primal Vigor). Counter
	// optionally restricts the replacement to a single counter kind.
	ConditionPredicateCounterPlacementOnAnyCreature
	// ConditionPredicateOpponentLifeLossDuringControllerTurn is satisfied when an
	// opponent of the controller would lose life during the controller's turn. It
	// gates the life-loss replacement "they lose twice that much life instead."
	// (Bloodletter of Aclazotz).
	ConditionPredicateOpponentLifeLossDuringControllerTurn
	// ConditionPredicateOpponentLifeLoss is satisfied when an opponent of the
	// controller would lose life at any time. It gates the untimed life-loss
	// doubling generalization.
	ConditionPredicateOpponentLifeLoss
	// ConditionPredicateAnyPlayerLifeLoss is satisfied when any player would lose
	// life. It gates the any-player life-loss doubling generalization.
	ConditionPredicateAnyPlayerLifeLoss
)

// GraveyardRedirectScope identifies whose graveyard a card-to-graveyard
// replacement watches.
type GraveyardRedirectScope uint8

// Graveyard redirect scopes recognized by the semantic compiler.
const (
	GraveyardRedirectScopeAny GraveyardRedirectScope = iota
	GraveyardRedirectScopeYou
	GraveyardRedirectScopeOpponent
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
	ConditionSupertypeLegendary
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

// ConditionCombatState is a closed semantic combat-involvement selection filter.
type ConditionCombatState uint8

// Condition combat-state values.
const (
	ConditionCombatStateAny ConditionCombatState = iota
	ConditionCombatStateAttacking
	ConditionCombatStateBlocking
	ConditionCombatStateAttackingOrBlocking
)

// ConditionComparisonScope selects which players' battlefields a control-count
// comparison counts.
type ConditionComparisonScope uint8

// Condition comparison scope values.
const (
	// ConditionComparisonScopeController counts the controller's permanents
	// ("you").
	ConditionComparisonScopeController ConditionComparisonScope = iota
	// ConditionComparisonScopeAnyOpponent quantifies existentially over
	// opponents ("an opponent").
	ConditionComparisonScopeAnyOpponent
	// ConditionComparisonScopeEachOpponent quantifies universally over opponents
	// ("each opponent").
	ConditionComparisonScopeEachOpponent
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
	CombatState       ConditionCombatState
	Keyword           parser.KeywordKind
	PowerAtLeast      int
	MatchPowerAtLeast bool
	// TotalPowerAtLeast is the collective-power threshold for a "have total
	// power <n> or greater" qualifier. MatchTotalPowerAtLeast marks it present.
	TotalPowerAtLeast      int
	MatchTotalPowerAtLeast bool
	// DistinctNamesAtLeast is the distinct-name threshold for a "with different
	// names" qualifier. MatchDistinctNamesAtLeast marks it present.
	DistinctNamesAtLeast      int
	MatchDistinctNamesAtLeast bool
	// DamageRecipientOpponent, DamageNoncombatOnly, and DamageSourceAnyController
	// qualify a damage-by-controlled-source clause: opponent-only recipient,
	// noncombat-only damage, and a source controlled by any player respectively.
	DamageRecipientOpponent   bool
	DamageNoncombatOnly       bool
	DamageSourceAnyController bool
}

// CompiledCondition is a closed, source-spanned semantic condition.
type CompiledCondition struct {
	Kind          ConditionKind
	Span          shared.Span
	Text          string
	Intervening   bool
	Resolving     bool
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
	// EventHistoryMinCount is the minimum number of matching events the window
	// must contain when Predicate is ConditionPredicateEventHistory. A zero
	// value means a single matching event suffices.
	EventHistoryMinCount int

	// Order is the condition's dense source-order rank. The compiler tests
	// whether a reference or payment falls within the condition by comparing
	// these ranks instead of inspecting byte offsets.
	Order shared.SourceOrder

	// ControlComparisonLeft, ControlComparisonRight, and ControlComparisonGreater
	// describe a ConditionPredicateControlComparison: the subject and reference
	// player scopes and whether the subject must control strictly more (true) or
	// fewer (false) permanents matching Selection.
	ControlComparisonLeft    ConditionComparisonScope
	ControlComparisonRight   ConditionComparisonScope
	ControlComparisonGreater bool

	// SourceInGraveyard marks a condition introduced by "this card is in your
	// graveyard and ...", reporting that the enclosing static ability functions
	// from the graveyard zone. The remaining predicate carries the accompanying
	// runtime condition; lowering reads this flag to set the ability's zone of
	// function rather than emitting a runtime predicate for it.
	SourceInGraveyard bool

	// GraveyardRedirectScope, GraveyardSubjectTypesAny, and
	// GraveyardFromBattlefieldOnly carry the parameters of a
	// ConditionPredicateCardWouldGoToGraveyard clause: whose graveyard is
	// watched, which card types the moving object may have (empty matches any
	// card), and whether the moving object can only leave the battlefield.
	GraveyardRedirectScope       GraveyardRedirectScope
	GraveyardSubjectTypesAny     []TriggerCardType
	GraveyardFromBattlefieldOnly bool
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
	ChoiceSpan  shared.Span
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
	SelectorCommander
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
	// NonToken records a "nontoken" selector qualifier; TokenOnly records a
	// "token" qualifier. They lower to Selection.NonToken / Selection.TokenOnly.
	NonToken  bool
	TokenOnly bool
	Keyword   parser.KeywordKind
	// ExcludedKeyword records a "without <keyword>" selector qualifier (e.g.
	// "each creature without flying"); it is mutually exclusive with Keyword.
	ExcludedKeyword parser.KeywordKind
	Zone            zone.Type
	ManaValue       compare.Int
	MatchManaValue  bool
	// ManaValueX records that the MatchManaValue bound is the spell's chosen {X}
	// ("with mana value X or less") rather than a fixed number; ManaValue then
	// holds only the operator. It lowers to SearchSpec.MaxManaValueFromX.
	ManaValueX     bool
	Power          compare.Int
	MatchPower     bool
	Toughness      compare.Int
	MatchToughness bool
	Colorless      bool
	Multicolored   bool
	BasicLandType  bool
	// MatchCounter records whether RequiredCounter is active ("creature you
	// control with a +1/+1 counter on it"); RequiredCounter names the counter
	// kind the matched permanent must carry.
	MatchCounter    bool
	RequiredCounter counter.Kind
	// PlayerOrPlaneswalker marks the combined "player or planeswalker" /
	// "opponent or planeswalker" combined damage target. Kind stays
	// SelectorPlayer or SelectorOpponent; this flag records the additional
	// planeswalker-permanent half the merged Kind cannot express.
	PlayerOrPlaneswalker bool
	// SubtypeFromEntryChoice requires each matched permanent to share the creature
	// subtype the source permanent chose as it entered ("creatures you control of
	// the chosen type"). It lowers to Selection.SubtypeFromSourceEntryChoice.
	SubtypeFromEntryChoice bool
	// ConjunctiveTypes records that a multi-member RequiredTypesAny names types a
	// permanent must carry all at once ("artifact creature") rather than any one
	// of ("artifact or creature"). It lowers the type set to the conjunctive
	// TargetPredicate.PermanentTypesAll filter instead of PermanentTypes.
	ConjunctiveTypes bool
	Alternatives     []CompiledSelector
	atoms            *CompiledSelectorAtoms
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
	ExcludedSubtypes   []types.Sub
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

// ExcludedSubtypes returns subtype filters excluded from this selector (a
// "non-<subtype>" filter such as "non-Human creatures").
func (s CompiledSelector) ExcludedSubtypes() []types.Sub {
	return selectorAtoms(s).ExcludedSubtypes
}

func appendSelectorExcludedSubtypes(selector *CompiledSelector, subtypes ...types.Sub) {
	atoms := mutableSelectorAtoms(selector)
	atoms.ExcludedSubtypes = append(atoms.ExcludedSubtypes, subtypes...)
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
	EffectManaSpendRider
	EffectModifyPT
	EffectMustAttack
	EffectMustBeBlocked
	EffectPut
	EffectProliferate
	EffectRegenerate
	EffectReorderLibraryTop
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
	EffectLifeTotalCantChange
	EffectProtectionFromEverything
	EffectPhaseOut
	EffectImpulseExile
	EffectAdditionalLandPlays
	EffectLoseGame
	EffectChooseNewTargets
	EffectCastAsThoughFlash
	EffectCantCastSpells
	EffectWinGame
	EffectPreventDamage
	EffectSpellsCantBeCountered
	EffectEnterAsCopy
	EffectPunisherLoseLife
	EffectMassReanimationExchange
	EffectRepeatProcess
	EffectMoveCounters
	EffectCopyStackObject
	EffectBecomeCopy
	EffectAmass
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
	// DurationUntilEndOfYourNextTurn matches "until the end of your next turn",
	// the bounded play window impulse-draw effects grant ("exile the top card …
	// until the end of your next turn, you may play that card"). The effect
	// expires at the end of the controller's next turn.
	DurationUntilEndOfYourNextTurn
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
	StaticSubjectControlledPermanents
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
	StaticSubjectControlledArtifactCreatures
	StaticSubjectOtherControlledArtifactCreatures
	StaticSubjectControlledNontokenCreatures
	StaticSubjectOtherControlledNontokenCreatures
	StaticSubjectAllLands
	StaticSubjectControlledCreaturesChosenType
	StaticSubjectOtherControlledCreaturesChosenType
	StaticSubjectOpponentControlledPermanents
	StaticSubjectOtherAttackingCreatures
)

// CompiledEffect is one recognized instruction verb and the sentence containing
// it. Multiple effects may refer to the same sentence when instructions are
// coordinated.
type CompiledEffect struct {
	Kind                 EffectKind
	Context              parser.EffectContextKind
	Connection           parser.EffectConnectionKind
	ConnectionSpan       shared.Span
	Span                 shared.Span
	ClauseSpan           shared.Span
	Text                 string
	VerbSpan             shared.Span
	Player               parser.EffectPlayerKind
	CardSource           parser.EffectCardSourceKind
	RequirePermanentCard bool
	References           []CompiledReference
	SubjectReferences    []CompiledReference
	Targets              []CompiledTarget
	SubjectTargets       []CompiledTarget
	Duration             DurationKind
	DelayedTiming        game.DelayedTriggerTiming
	Selector             CompiledSelector
	// DamageRecipientSelectors holds the compiled recipient groups of a
	// dual-recipient fixed group-damage effect ("deals N damage to each X and
	// each Y"). It is empty for single-recipient damage; when present it has
	// exactly two entries that lowering damages in Oracle order.
	DamageRecipientSelectors []CompiledSelector
	// DamageRecipientReference marks a damage recipient that is the controller or
	// owner of a referenced object (the prior removal target), as in "deals 2
	// damage to that land's controller". It is None for every other recipient.
	DamageRecipientReference parser.DamageRecipientReferenceKind
	// EachSourceDamageGroup is the source group of an "each <group> deals N
	// damage to its controller/owner" effect ("Each creature deals 1 damage to
	// its controller."), where every group member is the damage source dealing
	// to the player who controls (or owns) it. EachSourceDamageRecipient records
	// the per-source recipient role; it is None for every other effect.
	EachSourceDamageGroup     CompiledSelector
	EachSourceDamageRecipient parser.DamageRecipientReferenceKind
	// HasSelfDamageRider reports a "... and N damage to you" rider on a
	// single-target deal-damage clause ("deals A damage to any target and B
	// damage to you"). SelfDamageRiderValue holds the fixed self-damage amount B
	// dealt to the source's own controller; lowering emits a second Damage
	// instruction after the primary target damage.
	HasSelfDamageRider   bool
	SelfDamageRiderValue int
	// TargetControllerDamageRiderRecipient marks a "... and B damage to that
	// creature's controller/owner" rider on a single-target deal-damage clause
	// ("deals A damage to target creature and B damage to that creature's
	// controller"). TargetControllerDamageRiderValue holds the fixed rider
	// amount B; lowering emits a second Damage instruction to the primary
	// target's controller or owner after the primary target damage.
	TargetControllerDamageRiderRecipient parser.DamageRecipientReferenceKind
	TargetControllerDamageRiderValue     int
	// HasSecondTargetDamageRider reports a "... and B damage to <second target>"
	// rider on a single-target deal-damage clause ("deals A damage to target
	// creature and B damage to target player or planeswalker").
	// SecondTargetDamageRiderValue holds the fixed rider amount B dealt to the
	// clause's second target; lowering emits a second Damage instruction after
	// the primary target damage.
	HasSecondTargetDamageRider   bool
	SecondTargetDamageRiderValue int
	Amount                       CompiledAmount
	PowerDelta                   CompiledSignedAmount
	ToughnessDelta               CompiledSignedAmount
	TokenPower                   int
	TokenToughness               int
	TokenPTKnown                 bool
	// TokenName is a created creature token's explicit Oracle name ("named Koma's
	// Coil"), captured verbatim from source. It is empty when the token is named
	// only by its subtypes.
	TokenName         string
	TokenCopyOfTarget bool
	// AmassSubtype is the creature subtype named by an EffectAmass keyword action
	// ("Amass Orcs N" -> Orc, "Amass Zombies N" -> Zombie). The untyped "Amass N"
	// form defaults to Zombie. Lowering carries it onto game.Amass so the runtime
	// builds the Army token with this subtype when one must be created.
	AmassSubtype types.Sub
	// TokenCopyOfReference reports that the created token is a copy of the
	// effect's single explicit reference ("Create a token that's a copy of this
	// creature[ instead]."). The copy source is the lone reference in References,
	// not a grammatical target.
	TokenCopyOfReference bool
	// TokenCopyOfAttached reports that the created token is a copy of the
	// permanent the source is attached to ("a copy of equipped creature" /
	// "enchanted creature"). The copy source resolves at runtime to the attached
	// permanent.
	TokenCopyOfAttached bool
	// TokenCopyDropLegendary reports a copy-token "except <it/the token> isn't
	// legendary" modifier: the created token drops the Legendary supertype.
	TokenCopyDropLegendary bool
	// TokenCopyGrantKeywords lists keyword abilities the created copy token gains
	// from a folded "[That token/It] gains <keyword>." rider, in source order.
	TokenCopyGrantKeywords []parser.KeywordKind
	// TokenCopyGrantRiderSpan covers the folded gain-keyword rider sentence so
	// lowering credits its tokens toward source coverage.
	TokenCopyGrantRiderSpan shared.Span
	// TokenChoice reports a create-token effect offering a choice among two or
	// more complete named-token specs ("create a Food token or a Treasure token",
	// "create your choice of a Clue token, a Food token, or a Treasure token").
	// The alternatives are the selector's SubtypesAny entries in source order;
	// lowering emits a choose-one modal ability creating exactly one of them.
	TokenChoice       bool
	StaticSubject     StaticSubjectKind
	StaticSubjectSpan shared.Span
	Details           *CompiledEffectDetails
	CounterKind       counter.Kind
	CounterKindKnown  bool
	// CounterRecipientAttached reports that a counter-placement effect places its
	// counters on the permanent the source Aura is attached to ("... on enchanted
	// creature"). Lowering routes it to the runtime's source attached-permanent
	// reference; it is false for every other recipient.
	CounterRecipientAttached bool
	// MoveCountersAll carries the parser's kind-agnostic "move all counters"
	// form of an EffectMoveCounters effect through to lowering, which moves every
	// counter on the source regardless of kind. It is false for a specific-kind
	// move, whose kind is in CounterKind / CounterKindKnown.
	MoveCountersAll bool
	FromZone        zone.Type
	// GraveyardZoneExile carries the parser's recognized whole-graveyard exile
	// owner relation ("Exile target player's graveyard.") through to lowering,
	// which builds the target-player + graveyard-group MoveCard. It is
	// GraveyardZoneExileNone for every other effect.
	GraveyardZoneExile parser.GraveyardZoneExileKind
	ToZone             zone.Type
	Destination        parser.EffectDestinationPosition
	EntersTapped       bool
	EntersTappedSelf   bool
	// EntersTappedGroup mirrors the parser flag for a static enters-tapped
	// replacement that taps a group of OTHER permanents as they enter (Authority
	// of the Consuls). Lowering reads it to build a continuous controller- and
	// type-scoped replacement; it is false for the self enters-tapped form.
	EntersTappedGroup        bool
	EntersTappedGroupScope   parser.EntersTappedGroupControllerScope
	EntersTappedGroupTypes   []types.Card
	EntersColorChoice        bool
	EntersColorChoiceExclude mana.Color
	EntersTypeChoice         bool
	EntersWithCounters       bool
	// EntersAsCopy mirrors the parser's enters-as-copy replacement flag and its
	// riders. Lowering reads the effect's Selector for the copied-permanent
	// filter and these flags for the "you may" form and the copiable riders.
	EntersAsCopy             bool
	EntersAsCopyOptional     bool
	EntersAsCopyNotLegendary bool
	EntersAsCopyAddTypes     []types.Card
	// EntersAsCopyConditionalCounters mirrors the parser's conditional copiable
	// counter riders (Spark Double). Lowering builds one
	// game.ConditionalCounterPlacement per entry.
	EntersAsCopyConditionalCounters []parser.EntersAsCopyConditionalCounter
	// BecomeCopyUntilEndOfTurn, BecomeCopyRetainsThisAbility, and
	// BecomeCopyAddKeywords mirror the parser's EffectBecomeCopy duration and
	// copiable exception riders. Lowering reads them to build the runtime copy
	// effect's duration and granted-keyword/retained-ability riders.
	BecomeCopyUntilEndOfTurn     bool
	BecomeCopyRetainsThisAbility bool
	BecomeCopyAddKeywords        []parser.KeywordKind
	// EntersAsCopyUntilEndOfTurn mirrors the parser's temporary "become a copy
	// ... until end of turn" copy duration (Cursed Mirror).
	EntersAsCopyUntilEndOfTurn bool
	// EntersAsCopyAddKeywords mirrors the parser's "except it has <keyword>"
	// copiable keyword riders (Cursed Mirror's haste).
	EntersAsCopyAddKeywords []parser.KeywordKind
	UnderYourControl        bool
	CastAsAdventure         bool
	// CastWithoutPayingManaCost mirrors the parser's free-cast rider flag for a
	// cast effect ("... without paying its mana cost"). Lowering reads it to
	// route the cast-for-free primitive; it is false for every other effect.
	CastWithoutPayingManaCost bool
	Negated                   bool
	// FallbackOnInability mirrors the parser flag for a "who can't" relative
	// clause effect ("Each player who can't discards a card."): it applies only
	// to players who couldn't satisfy the immediately preceding required action.
	FallbackOnInability bool
	Optional            bool
	Divided             bool
	OptionalSpan        shared.Span
	Mana                CompiledEffectMana
	Replacement         parser.EffectReplacementSyntax
	Payment             CompiledEffectPayment
	Exact               bool
	// SourceSpellCostReduction and SourceSpellCostReductionAmount carry the typed
	// source-scoped cast cost reduction recognized by the parser ("This spell
	// costs {N} less to cast for each <countable battlefield object>"). Amount
	// holds the per-object battlefield count; SourceSpellCostReductionAmount is
	// the per-object generic reduction N. Lowering reads these typed values
	// instead of inspecting source text.
	SourceSpellCostReduction       bool
	SourceSpellCostReductionAmount int
	// SourceSpellCostReductionDynamic carries the typed source-scoped cast cost
	// reduction whose amount is this effect's own dynamic Amount ("This spell
	// costs {X} less to cast, where X is <dynamic amount>"). Lowering reads the
	// typed Amount instead of inspecting source text.
	SourceSpellCostReductionDynamic bool
	RequiresOrderedLowering         bool
	HasUnrecognizedSibling          bool
	UnsupportedDetail               string
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
	// CopyMayChooseNewTargets reports a copy-stack-object effect carrying the
	// optional "You may choose new targets for the copy[ies]." rider.
	// CopyChooseNewTargetsRiderSpan covers the rider sentence so lowering can
	// credit its tokens toward source coverage.
	CopyMayChooseNewTargets       bool
	CopyChooseNewTargetsRiderSpan shared.Span
	// Dig carries the impulse put clause's structured fields from the parser so
	// the combined dig lowerer can pair an EffectDig look with its EffectPut put.
	Dig parser.DigSyntax
	// HandLibraryPut carries the exact own-hand-to-library-top ordering clause
	// through the text-blind compiler boundary.
	HandLibraryPut parser.HandLibraryPutSyntax
	// HandDiscard carries the exact fixed-cardinality own-hand discard clause
	// through the text-blind compiler boundary.
	HandDiscard parser.HandDiscardSyntax
	// SearchSplit carries the split-destination put clause's structured fields
	// from the parser so the search lowerer can build a SearchSpec.SplitDestination
	// from typed slots rather than re-reading the put text.
	SearchSplit parser.SearchSplitSyntax
	// ManaSpendRider carries the typed mana-spend rider recognized by the parser
	// (an EffectManaSpendRider effect that rides on a preceding add-mana effect).
	// It is nil for every other effect. Lowering reads its typed condition and
	// effect rather than inspecting source text.
	ManaSpendRider *CompiledManaSpendRider
	// SearchSharedSubtype carries the "that share a land type" correlation rider
	// from the parser so the search lowerer can set SearchSpec.SharedSubtype
	// without re-reading the search text.
	SearchSharedSubtype bool
	// SearchDestination carries the parser-recognized ordered destination for a
	// found card that remains in the library.
	SearchDestination parser.EffectDestinationPosition
	// DiscardEntireHand carries the parser-recognized "discard their hand" clause
	// through the text-blind compiler boundary: the affected player discards
	// every card in hand rather than a fixed count.
	DiscardEntireHand bool
	// CounteredSpellExileReplacement carries the parser-recognized "If that
	// spell is countered this way, exile it instead of putting it into its
	// owner's graveyard." rider through the text-blind compiler boundary.
	CounteredSpellExileReplacement bool
	// CantCastSpellsAllPlayers mirrors the parser flag for an EffectCantCastSpells
	// clause that affects every player ("Players can't cast spells this turn.")
	// rather than only the controller's opponents. Lowering reads it to pick the
	// affected-player relation; it is false for the opponents-only form.
	CantCastSpellsAllPlayers bool
	// PreventDamageTo and PreventDamageBy mirror the parser flags for an
	// EffectPreventDamage clause, recording whether all combat damage dealt to
	// and/or dealt by the referenced permanent is prevented for the turn.
	PreventDamageTo bool
	PreventDamageBy bool
	// SpellsCantBeCounteredNextOnly mirrors the parser flag for an
	// EffectSpellsCantBeCountered clause that limits the buff to the single next
	// spell the controller casts rather than every spell cast this turn.
	SpellsCantBeCounteredNextOnly bool
	// DoublePower and DoubleToughness mirror the parser flags for an EffectDouble
	// whose object is "the power[ and toughness] of <group>" (Unnatural Growth).
	// Lowering reads them together with StaticSubject to emit a power/toughness
	// doubling continuous effect; both are false for every other double effect.
	DoublePower     bool
	DoubleToughness bool
	// UnderOwnersControl mirrors the parser flag for a battlefield-destination
	// effect carrying the "under their owners' control" rider (Open the Vaults,
	// Planar Birth), where each moved card enters under its owner's control. It
	// is false for the bare and "under your control" forms.
	UnderOwnersControl bool
	// TokenCopyOfForEach mirrors the parser flag for a per-each copy-token create
	// whose copy source is each member of a controlled battlefield group (Second
	// Harvest). The iterated group is carried in TokenCopyForEachGroup.
	TokenCopyOfForEach bool
	// TokenCopyForEachGroup carries the controlled battlefield group a
	// TokenCopyOfForEach create iterates, copying each member in turn.
	TokenCopyForEachGroup CompiledSelector
	// PunisherSacrifice and PunisherDiscard mirror the parser flags for an
	// EffectPunisherLoseLife effect ("... unless that player sacrifices a
	// permanent of their choice or discards a card."): they record which
	// alternatives the affected players may pay instead of losing life. Lowering
	// reads them with the effect's Selector for the sacrifice filter.
	PunisherSacrifice bool
	PunisherDiscard   bool
	// RepeatBody carries the sub-effect(s) of a "Repeat the following process X
	// times. <body>" loop (EffectRepeatProcess). Lowering lowers it to a nested
	// AbilityContent executed Amount times; it is nil for every other effect.
	RepeatBody []CompiledEffect
	// ReturnAsEnchantment mirrors the parser flag for a return-to-battlefield
	// effect carrying an "It's an enchantment." rider (the Enduring cycle): the
	// returned permanent enters as an Enchantment, losing its creature type.
	// ReturnAsEnchantmentRiderSpan covers the rider sentence so lowering credits
	// it toward source coverage.
	ReturnAsEnchantment          bool
	ReturnAsEnchantmentRiderSpan shared.Span
}

// CompiledManaSpendRider is the typed semantic form of a mana-spend rider.
type CompiledManaSpendRider struct {
	Condition  parser.ManaSpendConditionKind
	Effect     parser.ManaSpendRiderEffectKind
	Restricted bool
	ScryAmount int
}

// compileManaSpendRider maps the parser's mana-spend rider syntax to its typed
// semantic form. It mechanically copies the closed typed fields and never reads
// source text; nil maps to nil.
func compileManaSpendRider(syntax *parser.ManaSpendRiderSyntax) *CompiledManaSpendRider {
	if syntax == nil {
		return nil
	}
	return &CompiledManaSpendRider{
		Condition:  syntax.Condition,
		Effect:     syntax.Effect,
		Restricted: syntax.Restricted,
		ScryAmount: syntax.ScryAmount,
	}
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
	// ChosenColorDevotion mirrors the parser's "an amount of mana of that color
	// equal to your devotion to that color." body (Nykthos, Shrine to Nyx). The
	// produced mana is the color chosen as the ability resolves; its amount is the
	// controller's devotion to that chosen color. See
	// parser.EffectManaSyntax.ChosenColorDevotion.
	ChosenColorDevotion bool
	// ChosenColorDynamic mirrors the parser's "an amount of mana of that color
	// equal to <dynamic count>" body (Three Tree City). The produced mana is the
	// color chosen as the ability resolves; its amount is the battlefield count
	// carried by the effect's Amount. See parser.EffectManaSyntax.ChosenColorDynamic.
	ChosenColorDynamic bool
	CommanderIdentity  bool
	DynamicColorless   bool
	LegacyBodyExact    bool
	// FilterPair and FilterColors mirror the parser's filter-land output body
	// "{X}{X}, {X}{Y}, or {Y}{Y}." (FilterColors holds the pair's two distinct
	// basic colors {X, Y}). See parser.EffectManaSyntax.FilterPair.
	FilterPair   bool
	FilterColors []mana.Color
	// LandsProduce and LandsProduceScope mirror the parser's "one mana of any
	// color that a land <scope> could produce" body (Exotic Orchard, Reflecting
	// Pool, Fellwar Stone). See parser.EffectManaSyntax.LandsProduce.
	LandsProduce        bool
	LandsProduceScope   parser.ManaLandsProduceScope
	LandsProduceAnyType bool
	// LinkedExileColors mirrors the parser's "one mana of any of the exiled
	// card's colors" body (Chrome Mox). See parser.EffectManaSyntax.LinkedExileColors.
	LinkedExileColors bool
	// ColorsAmongControlled mirrors the parser's "one mana of any color among
	// <permanents> you control" body (Mox Amber, Plaza of Heroes). The choosable
	// colors are recomputed at resolution as the union of colors of the
	// controller's permanents matching ColorsAmongSelector. See
	// parser.EffectManaSyntax.ColorsAmongControlled.
	ColorsAmongControlled bool
	// ColorsAmongSelector carries the permanent filter of a ColorsAmongControlled
	// body. It is set together with ColorsAmongControlled.
	ColorsAmongSelector *CompiledSelector
	// EachColorAmongControlled mirrors the parser's "For each color among
	// <permanents> you control, add one mana of that color" body (Bloom Tender).
	// One mana of each color in the union of the controller's permanents matching
	// ColorsAmongSelector is produced. See
	// parser.EffectManaSyntax.EachColorAmongControlled.
	EachColorAmongControlled bool
	// AnyOneColorDynamic mirrors the parser's "X mana of any one color" (or "an
	// amount of mana of any one color") body whose quantity is a dynamic amount
	// carried by the effect's Amount (Kami of Whispered Hopes). The produced mana
	// is the single color chosen as the ability resolves; its amount is the
	// dynamic value. See parser.EffectManaSyntax.AnyOneColorDynamic.
	AnyOneColorDynamic bool
	// AnyColorCount mirrors the parser's "<N> mana of any one color" body
	// (Gilded Lotus: "Add three mana of any one color."), N >= 2. It is set
	// together with AnyColor; N mana of the single chosen color are produced. See
	// parser.EffectManaSyntax.AnyColorCount.
	AnyColorCount int
	// Instead mirrors the parser's trailing-"instead" flag on a conditional
	// alternative mana production ("Add {B}{B}{B}{B}{B} instead if ...", the
	// Threshold cycle). See parser.EffectManaSyntax.Instead.
	Instead bool
}

// CompiledEffectPayment is a typed resolution payment embedded in an effect.
type CompiledEffectPayment struct {
	Span              shared.Span
	Form              parser.EffectPaymentForm
	Payer             parser.EffectPaymentPayerKind
	ManaCost          cost.Mana
	GenericManaAmount CompiledAmount
	// AdditionalCost is a non-mana resolution payment cost (such as "sacrifice a
	// land"). It is nil for mana-only payments; ManaCost and AdditionalCost are
	// never both set.
	AdditionalCost         *CompiledCost
	SuccessConditionNodeID int
	FailureConditionNodeID int
	// Order is the payment's dense source-order rank, used to test condition
	// containment without byte offsets.
	Order shared.SourceOrder
}

// CompiledEffectDetails holds rarely-used effect details outside the hot effect
// value copied during instruction scans.
type CompiledEffectDetails struct {
	StaticSubjectType    *CompiledStaticSubjectType
	StaticSubjectColors  *CompiledStaticSubjectColors
	StaticSubjectKeyword *CompiledStaticSubjectKeyword
	Symbol               string
}

// CompiledStaticSubjectType preserves a static subject's printed subtype and its
// parser-resolved canonical subtype when known.
type CompiledStaticSubjectType struct {
	Text     string
	Sub      types.Sub
	Known    bool
	Excluded bool
}

// CompiledStaticSubjectColors preserves a static subject's optional color filter:
// the single colors matched disjunctively and the colorless/multicolored
// color-family qualifiers.
type CompiledStaticSubjectColors struct {
	ColorsAny    []parser.Color
	Colorless    bool
	Multicolored bool
}

// CompiledStaticSubjectKeyword preserves a static subject's optional single
// keyword filter ("Creatures with flying ...", "Creatures without flying ...").
// Excluded distinguishes the "without" exclusion from the "with" requirement.
type CompiledStaticSubjectKeyword struct {
	Keyword  parser.KeywordKind
	Excluded bool
}

func staticSubjectType(text string, sub types.Sub, known, excluded bool) *CompiledStaticSubjectType {
	if text == "" && !known {
		return nil
	}
	return &CompiledStaticSubjectType{Text: text, Sub: sub, Known: known, Excluded: excluded}
}

func staticSubjectColors(colors []parser.Color, colorless, multicolored bool) *CompiledStaticSubjectColors {
	if len(colors) == 0 && !colorless && !multicolored {
		return nil
	}
	return &CompiledStaticSubjectColors{ColorsAny: colors, Colorless: colorless, Multicolored: multicolored}
}

func staticSubjectKeyword(keyword, excludedKeyword parser.KeywordKind) *CompiledStaticSubjectKeyword {
	if keyword != parser.KeywordUnknown {
		return &CompiledStaticSubjectKeyword{Keyword: keyword}
	}
	if excludedKeyword != parser.KeywordUnknown {
		return &CompiledStaticSubjectKeyword{Keyword: excludedKeyword, Excluded: true}
	}
	return nil
}

func compiledEffectDetails(staticType *CompiledStaticSubjectType, staticColors *CompiledStaticSubjectColors, staticKeyword *CompiledStaticSubjectKeyword, symbol string) *CompiledEffectDetails {
	if staticType == nil && staticColors == nil && staticKeyword == nil && symbol == "" {
		return nil
	}
	return &CompiledEffectDetails{StaticSubjectType: staticType, StaticSubjectColors: staticColors, StaticSubjectKeyword: staticKeyword, Symbol: symbol}
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

// StaticSubjectSubExcluded reports whether the static subject subtype is a
// "non-<subtype>" exclusion ("Non-Human creatures you control get ...").
func (e *CompiledEffect) StaticSubjectSubExcluded() bool {
	return e.Details != nil && e.Details.StaticSubjectType != nil && e.Details.StaticSubjectType.Excluded
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

// StaticSubjectKeyword returns the static subject's optional single keyword
// filter, whether it is an exclusion, and whether any keyword filter is present.
func (e *CompiledEffect) StaticSubjectKeyword() (keyword parser.KeywordKind, excluded, present bool) {
	if e.Details == nil || e.Details.StaticSubjectKeyword == nil {
		return parser.KeywordUnknown, false, false
	}
	return e.Details.StaticSubjectKeyword.Keyword, e.Details.StaticSubjectKeyword.Excluded, true
}

// Symbol returns the first mana symbol recognized in this effect.
func (e *CompiledEffect) Symbol() string {
	if e.Details == nil {
		return ""
	}
	return e.Details.Symbol
}

// CounterKindPlacementSupported reports whether named placement of kind has
// complete runtime semantics in the executable backend. Stun counters are
// supported: the untap step removes one stun counter instead of untapping a
// permanent that carries any (CR 122.6f). Finality counters remain unsupported
// because their death-replacement semantics are not yet modeled.
func CounterKindPlacementSupported(kind counter.Kind) bool {
	switch kind {
	case counter.Finality:
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
	// DynamicAmountSourceToughness is a referenced object's toughness ("its
	// toughness"), the toughness sibling of DynamicAmountSourcePower. Added last
	// so existing kinds keep their wire values.
	DynamicAmountSourceToughness
	// DynamicAmountSourceManaValue is a referenced object's mana value ("its
	// mana value", "that permanent's mana value"). It backs the
	// destroy-then-life-rider staples (Feed the Swarm, Divine Offering) where a
	// trailing life gain or loss reads the mana value of the permanent an
	// earlier clause destroyed. Added last so existing kinds keep their wire
	// values.
	DynamicAmountSourceManaValue
	// DynamicAmountSourceCounterCount is the number of counters of CounterKind
	// on the referenced object.
	DynamicAmountSourceCounterCount
	// DynamicAmountGreatestPower is the greatest power among the selector's
	// battlefield group ("the greatest power among <group>"). It backs the
	// dynamic "draw cards equal to the greatest power among creatures you
	// control" family. Added last so existing kinds keep their wire values.
	DynamicAmountGreatestPower
	// DynamicAmountGreatestToughness is the greatest toughness among the
	// selector's battlefield group, the toughness sibling of
	// DynamicAmountGreatestPower.
	DynamicAmountGreatestToughness
	// DynamicAmountGreatestManaValue is the greatest mana value among the
	// selector's battlefield group, the mana-value sibling of
	// DynamicAmountGreatestPower.
	DynamicAmountGreatestManaValue
	// DynamicAmountCommanderColorCount is the number of colors in the
	// controller's commander's color identity ("the number of colors in your
	// commanders' color identity"). It backs War Room's "pay life equal to ..."
	// activation cost. Added last so existing kinds keep their wire values.
	DynamicAmountCommanderColorCount
	// DynamicAmountDevotion is the controller's devotion to the amount's Colors
	// ("your devotion to <color>", "your devotion to <color> and <color>"), the
	// number of mana symbols of those colors among the mana costs of permanents
	// the controller controls (CR 700.5). It backs the devotion family such as
	// Gray Merchant of Asphodel. Added last so existing kinds keep their wire
	// values.
	DynamicAmountDevotion
	// DynamicAmountGreatestDiscardedThisWay is the greatest number of cards
	// discarded by any one player during a preceding discard effect in the same
	// ability ("the greatest number of cards a player discarded this way"). It
	// backs the Windfall draw amount and is realized by a sequence lowerer that
	// reads the maximum per-player discard count published by the preceding
	// discard instruction. Added last so existing kinds keep their wire values.
	DynamicAmountGreatestDiscardedThisWay
	// DynamicAmountSpellsCastThisTurn is the number of spells the controller has
	// cast this turn ("for each spell you've cast this turn"). It backs the
	// storm-counter family such as Aetherflux Reservoir. Added last so existing
	// kinds keep their wire values.
	DynamicAmountSpellsCastThisTurn
	// DynamicAmountTriggeringLifeChange is the amount of life gained or lost by
	// the event that triggered the enclosing life-change trigger ("that much
	// life"). It backs the life-drain mirror family (Sanguine Bond, Exquisite
	// Blood). Added last so existing kinds keep their wire values.
	DynamicAmountTriggeringLifeChange
	// DynamicAmountTotalPower is the sum of power across the selector's
	// battlefield group ("the total power of <group>"). It backs the dynamic
	// "where X is the total power of creatures you control" cost reduction
	// (Ghalta, Primal Hunger) and the matching draw and damage amounts.
	// DynamicAmountTotalToughness is the toughness sibling. Added last so
	// existing kinds keep their wire values.
	DynamicAmountTotalPower
	DynamicAmountTotalToughness
	// DynamicAmountColorCount is the number of distinct colors among the
	// selector's battlefield group ("the number of colors among <group>"). It
	// backs the "+1/+1 for each color among permanents you control" self-buff
	// family (Faeburrow Elder). Added last so existing kinds keep their wire
	// values.
	DynamicAmountColorCount
	// DynamicAmountSacrificedPower is the power of the permanent sacrificed to
	// pay the enclosing activated ability's cost ("the sacrificed creature's
	// power"). DynamicAmountSacrificedToughness and
	// DynamicAmountSacrificedManaValue are the toughness and mana-value
	// siblings. They back Altar of Dementia's "Sacrifice a creature: Target
	// player mills cards equal to the sacrificed creature's power." Added last
	// so existing kinds keep their wire values.
	DynamicAmountSacrificedPower
	DynamicAmountSacrificedToughness
	DynamicAmountSacrificedManaValue
	// DynamicAmountSharedCreatureTypeCount is the number of other creatures in
	// the selector's battlefield group that share at least one creature type with
	// the affected permanent ("for each other creature on the battlefield that
	// shares a creature type with it"). It backs the shared-creature-type anthem
	// family (Coat of Arms), a per-affected-creature dynamic power/toughness
	// bonus. Added last so existing kinds keep their wire values.
	DynamicAmountSharedCreatureTypeCount
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
	RangeKnown    bool
	Minimum       int
	Maximum       int
	VariableX     bool
	DynamicKind   DynamicAmountKind
	DynamicForm   DynamicAmountForm
	Multiplier    int
	ReferenceSpan shared.Span
	CounterKind   counter.Kind
	Text          string
	// Colors carries the colors of a devotion amount; empty otherwise.
	Colors   []color.Color
	selector *CompiledSelector
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

// CompiledEnchantTarget is the runtime-typed object restriction following an
// Enchant keyword, mapped from the parser's EnchantPredicate. A permanent
// matches when it has any listed card type or any listed subtype (disjunctive).
// Player, Opponent, and Permanent select non-type objects. Known is false when
// the predicate is empty or names a non-permanent card type or a subtype that no
// permanent type defines, so an unsupported Enchant target fails closed.
type CompiledEnchantTarget struct {
	Known     bool
	Player    bool
	Opponent  bool
	Permanent bool
	CardTypes []types.Card
	Subtypes  []types.Sub
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
	EnchantTarget   CompiledEnchantTarget
	Protection      game.ProtectionKeyword
	ProtectionKnown bool
	// EquipRestriction is the typed quality restriction of a restricted Equip
	// ability, or nil for an unrestricted Equip. It is set only when the parser
	// recognized every restriction word, so an unsupported restriction fails
	// closed upstream rather than reaching here.
	EquipRestriction *CompiledEquipRestriction
}

// CompiledEquipRestriction is the runtime-typed quality restriction of a
// restricted Equip ability: the Equipment may attach only to a creature with
// every listed supertype and at least one of the listed subtypes.
type CompiledEquipRestriction struct {
	Supertypes []types.Super
	Subtypes   []types.Sub
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
	ReferenceChosenCards
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
