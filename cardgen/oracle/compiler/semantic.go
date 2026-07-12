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
	"github.com/natefinch/council4/opt"
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
	AbilityLevelBand
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
	AbilityLevelBand:            "level band",
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
	Kind                                        AbilityKind
	Optional                                    bool
	CastOnlyDuringDeclareAttackersAfterAttacked bool
	GoadedOpponentCreaturesCantBlock            bool
	QuestForRenewalUntap                        bool
	SemblanceAnvilImprint                       bool
	SemblanceAnvilReduction                     bool
	CloudKeyChoice                              bool
	CloudKeyReduction                           bool
	AugurOfAutumnCoven                          bool
	EvolutionaryLeapRevealUntil                 bool
	FlameshadowConjuringCopy                    bool
	StingCombatFirstStrike                      bool
	YevaGreenCreatureFlash                      bool
	ProgenitorIconNextFlash                     bool
	// ExactSequence is a parser-recognized exact multi-instruction resolving
	// body. When set, the normal target/condition/effect content is empty and
	// lowering emits the fixed instruction template for the kind. It is declared
	// next to Optional so the byte packs into existing alignment padding.
	ExactSequence ExactSequenceKind
	// ExactSequenceBottom and ExactSequenceDrawOffset carry the typed parameters
	// of ExactSequenceBottomHandThenDraw: the library end the hand cards move to
	// and the fixed offset added to the "draw that many cards" count. They are
	// zero for all other exact sequences. Both are declared next to Optional and
	// ExactSequence so the bytes pack into existing alignment padding.
	ExactSequenceBottom     bool
	ExactSequenceDrawOffset uint8
	// ExactSequenceLookAtTopTypes carries the disjunctive card types of
	// ExactSequenceConditionalLookAtTopReveal: the reveal succeeds only when the
	// looked-at card matches one of these types. It is nil for all other exact
	// sequences. The parser owns the wording; this holds only the typed values.
	ExactSequenceLookAtTopTypes []types.Card
	// ExactSequenceLookAtTopEntersTapped, ExactSequenceLookAtTopElseHand, and
	// ExactSequenceLookAtTopElseBottom carry the typed parameters of
	// ExactSequenceConditionalLookAtTopBattlefield: the first records the "tapped"
	// battlefield entry rider, the second that the card moves into the
	// controller's hand when it is not put onto the battlefield, and the third
	// that the controller may instead put it on the bottom of their library (at
	// most one else flag is set; both false leaves the card on top of the
	// library). All are false for every other exact sequence.
	ExactSequenceLookAtTopEntersTapped bool
	ExactSequenceLookAtTopElseHand     bool
	ExactSequenceLookAtTopElseBottom   bool
	// ExactSequenceDrawCount and ExactSequenceDiscardCount carry the typed counts
	// of ExactSequenceDrawThenDiscardUnlessType: the cards drawn first, then the
	// cards discarded unless an exempt-type card (in ExactSequenceLookAtTopTypes)
	// is discarded instead. Both are zero for every other exact sequence.
	ExactSequenceDrawCount    uint8
	ExactSequenceDiscardCount uint8
	// ExactSequenceChooseCount and ExactSequencePayLife carry the typed counts of
	// ExactSequenceExtraDrawThenPayLifeOrTop: the cards chosen from among those
	// drawn this turn (M) and the life paid to keep each chosen card (L).
	// ExactSequenceDrawCount carries the additional cards drawn (N) for that
	// sequence. All three are zero for every other exact sequence.
	ExactSequenceChooseCount uint8
	ExactSequencePayLife     uint8
	Span                     shared.Span
	Text                     string
	ActivationTiming         ActivationTimingKind
	ActivationTimingSpan     shared.Span
	// MaxActivationsPerTurn caps activations per turn ("Activate no more than
	// twice each turn."). Zero means no cap. MaxActivationsPerTurnSpan covers the
	// recognized restriction sentence for source coverage.
	MaxActivationsPerTurn      int
	MaxActivationsPerTurnSpan  shared.Span
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
	// ClassLevelGain is the target level of a Class enchantment's level-up
	// activated ability ("{cost}: Level N"), or 0 when this ability is not a
	// level-up. The compiler copies the parser's typed level so lowering emits
	// the SetClassLevel ability without re-reading the body wording.
	ClassLevelGain int
	// LevelUpRecognized reports that the parser recognized this paragraph as a
	// leveler card's "Level up {cost}" activated ability (CR 711.2). LevelUpCost
	// carries its mana cost. Lowering emits a sorcery-speed ability that puts a
	// level counter on the source.
	LevelUpRecognized bool
	LevelUpCost       cost.Mana
	// LevelBand carries a leveler card's "LEVEL lo-hi" / "LEVEL lo+" band header
	// with its printed base power/toughness. It is nil for non-band abilities.
	LevelBand *CompiledLevelBand
	// Companion reports that the parser recognized this paragraph as a companion
	// keyword ability (CR 702.139). Its content is otherwise empty; lowering
	// emits the inert companion static keyword.
	Companion bool
	// PartnerWith reports that the parser recognized this paragraph as a
	// "Partner with <name>" keyword ability (CR 702.124e). Its content is
	// otherwise empty; lowering emits the inert partner-with static keyword.
	PartnerWith bool
	// ChooseABackground reports that the parser recognized this paragraph as a
	// "Choose a Background" keyword ability (CR 702.124f). Its content is
	// otherwise empty; lowering emits the inert choose-a-background static
	// keyword.
	ChooseABackground bool
	// Partner reports that the parser recognized this paragraph as a "Partner"
	// keyword ability or one of its "Partner—<quality>" restricted variants
	// (CR 702.124a, 702.124f). Its content is otherwise empty; lowering emits the
	// inert partner static keyword.
	Partner bool
	// KeywordShare carries the recognized team keyword-sharing construct (Odric,
	// Lunarch Marshal), or nil when this paragraph is not one. Its content is
	// otherwise empty; lowering emits one gated continuous group grant per shared
	// keyword under the paragraph's phase/step trigger.
	KeywordShare *CompiledKeywordShare
}

// CompiledKeywordShare is the runtime-typed team keyword-sharing construct: the
// ordered list of shared keyword kinds a phase/step trigger grants to all the
// controller's creatures, each gated on the controller controlling a creature
// that already has that keyword. The compiler carries only the typed keyword
// kinds; lowering maps each to its runtime keyword and gate.
type CompiledKeywordShare struct {
	Keywords []parser.KeywordKind
}

// CompiledLevelBand is a leveler card's "LEVEL lo-hi" / "LEVEL lo+" band
// (CR 711.4) with its printed base power/toughness. Low is the band's first
// level; High is its last level, or 0 for the open-ended final band. Power and
// Toughness hold the printed base P/T when HasPowerToughness is true.
type CompiledLevelBand struct {
	Low               int
	High              int
	Power             int
	Toughness         int
	HasPowerToughness bool
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
	// AlternativeCostConditionYourTurn gates a free alternative cost behind "If
	// it's your turn,".
	AlternativeCostConditionYourTurn
	// AlternativeCostConditionControlsSubtype gates a free alternative cost
	// behind "If you control a <subtype>," where the subtype rides on the
	// CompiledAlternativeCost's ConditionSubtype field.
	AlternativeCostConditionControlsSubtype
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
	// AlternativeCostEscape is the alternative-cost form of Escape: cast the
	// spell from the graveyard by paying the compound escape cost carried on the
	// ability's CompiledCost (its mana cost plus the graveyard-exile additional
	// cost). Unlike Flashback the spell is not exiled, so it can be escaped again.
	AlternativeCostEscape
	// AlternativeCostDiscard is the Foil/Outbreak family: discard one or more
	// cards (each an optional subtype filter) from hand rather than pay the
	// spell's printed mana cost. The discards are carried as typed cost
	// components on the ability's CompiledCost.
	AlternativeCostDiscard
	AlternativeCostBorderpost
	// AlternativeCostFree is the "free spell" family: cast the spell by paying a
	// single non-mana cost (carried on the ability's CompiledCost) rather than
	// its printed mana cost, optionally gated by a condition. Snuff Out ("If you
	// control a Swamp, you may pay 4 life ...") is the canonical member.
	AlternativeCostFree
)

// CompiledAlternativeCost is text-independent semantic data for an optional
// replacement of a spell's printed mana cost.
type CompiledAlternativeCost struct {
	Kind                  AlternativeCostKind
	Condition             AlternativeCostCondition
	ConditionSubtype      types.Sub
	WithoutPayingManaCost bool
	ManaCost              cost.Mana
	ReplaceTargetWithEach bool
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
	// ActivationTimingDuringYourTurnBeforeAttackers marks the combined "during
	// your turn, before attackers are declared" window (the Portal precombat
	// cycle). It lowers to a runtime timing restriction that permits activation
	// only on the controller's turn, before the declare-attackers step.
	ActivationTimingDuringYourTurnBeforeAttackers
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
	// ModesUniquePerTurn requires each mode to be chosen at most once per turn
	// for each source object and triggered ability.
	ModesUniquePerTurn bool
	// Spree marks a Spree modal whose options each carry an additional mana cost
	// (CR 702.171), recorded on each CompiledMode's SpreeCost.
	Spree bool
	// Escalate marks an Escalate modal (CR 702.121) whose controller pays
	// EscalateCost once for each mode chosen beyond the first.
	Escalate bool
	// EscalateCost is the shared per-extra-mode cost of an Escalate modal. It is
	// set only when Escalate is true.
	EscalateCost cost.Mana
}

// CompiledModalChoiceKind identifies exact typed modal header vocabulary.
type CompiledModalChoiceKind uint8

// Compiled modal choice kinds.
const (
	CompiledModalChoiceUnknown CompiledModalChoiceKind = iota
	CompiledModalChoiceOneOrMore
	CompiledModalChoiceOneAtRandom
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
	// SpreeCost is the additional mana cost paid to choose this option on a Spree
	// spell (CR 702.171). It is empty on non-Spree modes.
	SpreeCost cost.Mana
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

	AmountValue int
	AmountKnown bool
	AmountFromX bool

	// AmountOneOrMore reports a player-chosen "one or more" cost amount, where
	// the payer removes at least one of the named object. The chosen count is
	// announced as the ability's X (AmountFromX is also set); lowering carries
	// it onto an X-driven additional cost and the payment enumeration requires
	// at least one.
	AmountOneOrMore bool

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
	ObjectTokenOnly   bool
	PermanentModifier bool

	// ObjectExcludedType constrains a permanent cost object to permanents that
	// are not of the named card type ("nonland permanents"), recognized by the
	// parser. ObjectExcludedTypeKnown reports its presence.
	ObjectExcludedType      types.Card
	ObjectExcludedTypeKnown bool

	// ObjectExcludedSubtype constrains a permanent cost object to permanents
	// that lack the named subtype ("non-Lair land"), recognized by the parser.
	// ObjectExcludedSubtypeKnown reports its presence.
	ObjectExcludedSubtype      types.Sub
	ObjectExcludedSubtypeKnown bool

	RequireTapped    bool
	RequireUntapped  bool
	SourceZone       zone.Type
	ToZone           zone.Type
	SourceSelf       bool
	CounterKind      counter.Kind
	CounterKindKnown bool
	SubtypesAny      []types.Sub

	// RemoveCounterAmong reports a "remove N counters from among <permanents>
	// you control" cost, recognized by the parser. The removed counters are
	// spread across the chosen controlled permanents named by the object
	// selector fields rather than taken from the ability's own source.
	RemoveCounterAmong bool

	// ExcludeSource reports that the cost object excludes the ability's own
	// source ("another"), recognized by the parser.
	ExcludeSource bool

	// DiscardWholeHand reports a "discard your hand" cost object, recognized by
	// the parser. The payer discards every card in their hand.
	DiscardWholeHand bool

	// Random reports a "discard <count> card(s) at random" cost object,
	// recognized by the parser. The payer discards randomly chosen cards
	// rather than cards of their choice (CR 701.9a).
	Random bool

	// AnyNumber reports a variable-cardinality card cost object ("any number of
	// <cards>"), recognized by the parser. The payer chooses how many matching
	// cards to take, bounded by a constraint such as TotalManaValueAtLeast.
	AnyNumber bool

	// ObjectHistoric reports that the cost object is constrained to historic
	// cards (artifacts, legendaries, or Sagas), recognized by the parser.
	ObjectHistoric bool

	// TotalManaValueAtLeast, when positive, constrains a variable-cardinality
	// card cost to "<cards> with total mana value N or greater," recognized by
	// the parser. The payer takes enough matching cards to total at least N.
	TotalManaValueAtLeast int

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
	// TriggerState is a state trigger checked continuously; it fires whenever its
	// board-state condition holds while it is not already on the stack
	// (CR 603.8). Unlike event triggers it carries no event pattern, only a
	// StateCondition on its TriggerPattern.
	TriggerState
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
	ConditionPredicateControllerLifeAtMost
	ConditionPredicateControllerLifeAtLeastAboveStarting
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
	ConditionPredicateEventSubjectWasCastFromControllerHand
	ConditionPredicateEventSubjectEnteredOrCastFromGraveyard
	ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard
	ConditionPredicateEventSubjectHadNoCounter
	// ConditionPredicateEventSubjectHadCounter is satisfied when the event's
	// permanent had at least one counter of Counter's kind in its last-known
	// information ("if it had a +1/+1 counter on it"). It is the affirmative
	// counterpart of ConditionPredicateEventSubjectHadNoCounter.
	ConditionPredicateEventSubjectHadCounter
	ConditionPredicatePriorInstructionNotAccepted
	// ConditionPredicatePriorInstructionAccepted is satisfied when the prior
	// optional instruction was performed ("if you do"). It is the affirmative
	// complement of ConditionPredicatePriorInstructionNotAccepted.
	ConditionPredicatePriorInstructionAccepted
	// ConditionPredicateDestroyedThisWay is satisfied when a permanent matching
	// the named type was destroyed by the prior destroy effect ("if a creature is
	// destroyed this way"). The typed condition carries no selection, so it is the
	// resolving-success equivalent of "if you do" only when the named type matches
	// every object the prior clause could have destroyed; the lowering treats it
	// as an "if you do" gate solely for the optional-destroy shape and otherwise
	// fails closed.
	ConditionPredicateDestroyedThisWay
	// ConditionPredicateDiesThisWay is satisfied when the prior destroy effect's
	// target was actually put into a graveyard from the battlefield ("if that
	// creature dies this way"; Saw in Half). Unlike ConditionPredicateDestroyedThisWay
	// it back-references the destroy's single target rather than a named type, so
	// it gates the linked follow-up on that specific creature having died. The
	// dies-this-way copy sequence lowering treats it as the resolving-success gate
	// on the preceding destroy and otherwise fails closed.
	ConditionPredicateDiesThisWay
	ConditionPredicateCounterPlacementOnControlledCreature
	// ConditionPredicateCounterPlacementOnSelf is satisfied when one or more
	// counters would be put on the source permanent itself ("If one or more
	// +1/+1 counters would be put on Mowu, that many plus one +1/+1 counters are
	// put on it instead.", Mowu, Loyal Companion). Counter restricts the
	// replacement to a single counter kind.
	ConditionPredicateCounterPlacementOnSelf
	ConditionPredicateControllerCounterPlacement
	ConditionPredicateDamageByControlledSource
	ConditionPredicateDamageWouldBeDealtToPermanent
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
	// ConditionPredicateSourceTributeNotPaid is satisfied when the source
	// permanent's Tribute was not paid as it entered (CR 702.110). It gates the
	// paired "When this creature enters, if tribute wasn't paid, ..." ability.
	ConditionPredicateSourceTributeNotPaid
	// ConditionPredicateControllerControlsCommander is satisfied when the context
	// controller controls their commander on the battlefield ("if you control
	// your commander" / "as long as you control your commander"). It gates the
	// Lieutenant ability word's intervening and static conditions.
	ConditionPredicateControllerControlsCommander
	// ConditionPredicateSpellWasKicked is satisfied when the resolving spell was
	// kicked ("if this spell was kicked, ... instead"). It gates the kicked
	// effect variant against the spell's kicker-paid cast context.
	ConditionPredicateSpellWasKicked
	// ConditionPredicateGiftPromised is satisfied when the resolving spell's Gift
	// keyword action promised a gift to an opponent as it was cast ("if the gift
	// was promised, ..."; CR 702.171). It gates the promoted effect variant
	// against the spell's gift-promised cast context; its negation gates the "if
	// the gift wasn't promised" penalty clause.
	ConditionPredicateGiftPromised
	// ConditionPredicateSpellWasCastFromGraveyard is satisfied when the resolving
	// spell was cast from a graveyard ("if this spell was cast from a graveyard,
	// ..."). It gates a per-effect branch against the spell's source zone and is
	// false for copies. It backs Sevinne's Reclamation's conditional self-copy.
	ConditionPredicateSpellWasCastFromGraveyard
	// ConditionPredicateSourceSaddled is satisfied when the source Mount is
	// currently saddled (CR 702.166), as in "if this creature is saddled". It
	// gates a per-effect branch on the source's runtime saddled state.
	ConditionPredicateSourceSaddled
	// ConditionPredicateSourceNotSaddled is satisfied when the source Mount is
	// not currently saddled ("if this creature isn't saddled", Caustic Bronco).
	// It is the negated complement of ConditionPredicateSourceSaddled.
	ConditionPredicateSourceNotSaddled
	// ConditionPredicateAttackersAttackingControllerAtLeast is satisfied when at
	// least Threshold of the attackers declared by the triggering "an opponent
	// attacks with creatures" event are attacking the context controller
	// directly or one of the controller's planeswalkers ("if two or more of
	// those creatures are attacking you and/or planeswalkers you control";
	// Mangara, the Diplomat). It gates the trigger's intervening-if condition
	// against live combat state.
	ConditionPredicateAttackersAttackingControllerAtLeast
	// ConditionPredicateControllerLibrarySizeAtLeast is satisfied when the
	// controller's library holds at least Threshold cards ("if you have 200 or
	// more cards in your library", Battle of Wits).
	ConditionPredicateControllerLibrarySizeAtLeast
	// ConditionPredicateControllerLifeExactly is satisfied when the controller's
	// life total is exactly Threshold ("if you have exactly 1 life", Near-Death
	// Experience).
	ConditionPredicateControllerLifeExactly
	// ConditionPredicateControllerGainedLifeThisTurnAtLeast is satisfied when the
	// context controller has gained at least Threshold total life so far this
	// turn ("if you gained 3 or more life this turn"; Angelic Accord). It gates
	// an intervening-if trigger against the turn's accumulated life gain.
	ConditionPredicateControllerGainedLifeThisTurnAtLeast
	// ConditionPredicateSpellXAtLeast is satisfied when the resolving spell's
	// chosen value of {X} is at least Threshold ("if X is 10 or more", the
	// Finale cycle). It gates a per-effect branch against the resolving stack
	// object's captured X value.
	ConditionPredicateSpellXAtLeast
	// ConditionPredicateControllerGraveyardCardOfTypeCountAtLeast is satisfied
	// when the controller's graveyard holds at least Threshold cards of the
	// GraveyardCountCardType card type ("if twenty or more creature cards are in
	// your graveyard", Mortal Combat).
	ConditionPredicateControllerGraveyardCardOfTypeCountAtLeast
	// ConditionPredicateControllerGraveyardInstantOrSorceryCountAtLeast is
	// satisfied when the controller's graveyard holds at least Threshold cards
	// that are instants and/or sorceries ("Spell mastery — If there are two or
	// more instant and/or sorcery cards in your graveyard, ...", Fiery Impulse).
	ConditionPredicateControllerGraveyardInstantOrSorceryCountAtLeast
	// ConditionPredicateControllerDoesNotPay is satisfied when the ability's
	// controller does not pay the cost embedded in an "unless you pay {cost}"
	// effect ("sacrifice this creature unless you pay {U}"). It is derived from
	// the effect's controller-payer payment by applyEffectPaymentsToConditions
	// and gates the unless consequence on the payment being declined.
	ConditionPredicateControllerDoesNotPay
	// ConditionPredicateControllerControlsNamed is satisfied when the context
	// controller controls at least one permanent matching each card name listed
	// in ControlledNames ("If you control an Urza's Mine and an Urza's Tower,
	// ..."; the Urza tron lands). Names are compared case-insensitively with
	// hyphens and spaces treated alike.
	ConditionPredicateControllerControlsNamed
	// ConditionPredicateFirstCombatPhaseOfTurn is satisfied while the current
	// turn is still in its first combat phase ("if it's the first combat phase of
	// the turn"; Raiyuu, Storm's Edge, Karlach, Fury of Avernus). It gates an
	// extra-combat insertion against TurnState.CombatPhasesThisTurn so the loop
	// fires once per turn.
	ConditionPredicateFirstCombatPhaseOfTurn
	// ConditionPredicateControlsGreatestPowerCreature is satisfied when the
	// context controller controls a creature whose power is greater than or equal
	// to every other creature's power on the battlefield ("if you control the
	// creature with the greatest power or tied for the greatest power"; Summon:
	// Fenrir chapter III). It holds when the controller has the sole
	// highest-power creature or is tied for highest, and is false when no
	// creatures exist.
	ConditionPredicateControlsGreatestPowerCreature
	// ConditionPredicateControlsGreatestToughnessCreature is satisfied when the
	// context controller controls a creature whose toughness is greater than or
	// equal to every other creature's toughness on the battlefield ("if you
	// control the creature with the greatest toughness or tied for the greatest
	// toughness"; Abzan Beastmaster). It holds when the controller has the sole
	// highest-toughness creature or is tied for highest, and is false when no
	// creatures exist.
	ConditionPredicateControlsGreatestToughnessCreature
	// ConditionPredicateEventSubjectPowerGreatestOnBattlefield is satisfied when
	// the permanent named by the triggering zone-change event has power strictly
	// greater than every other creature's power on the battlefield ("if its power
	// is greater than each other creature's power"; Selvala, Heart of the Wilds).
	// It holds only when the event creature is the sole highest-power creature; a
	// tie or the absence of the event permanent fails closed.
	ConditionPredicateEventSubjectPowerGreatestOnBattlefield
	// ConditionPredicateSubjectSharesCreatureTypeWithSource is satisfied when the
	// condition's subject card (the just-looked-at top card of the controller's
	// library) shares at least one creature type with the source permanent ("if
	// it shares a creature type with this creature", the Kinship ability word).
	// It carries no selection parameters; the lowering binds it to the looked-at
	// card and the source permanent.
	ConditionPredicateSubjectSharesCreatureTypeWithSource
	// ConditionPredicateControllerIsMonarch is satisfied when the context
	// controller is the monarch (CR 720). It is a live single-player game-state
	// predicate with no clause parameters.
	ConditionPredicateControllerIsMonarch
	// ConditionPredicateControllerWasMonarchAtTurnStart is satisfied when the
	// context controller was the monarch (CR 720) as the current turn began, as in
	// "if you were the monarch as the turn began" (Knights of the Black Rose). It
	// reads the monarch snapshot taken when the turn advanced, not the live
	// designation.
	ConditionPredicateControllerWasMonarchAtTurnStart
	// ConditionPredicateAnOpponentIsMonarch is satisfied when any of the context
	// controller's opponents is the monarch (CR 720), as in "if an opponent is
	// the monarch" (Queen Marchesa). It is a live game-state predicate with no
	// clause parameters.
	ConditionPredicateAnOpponentIsMonarch
	// ConditionPredicateNoMonarch is satisfied when no player currently holds the
	// monarch designation ("if there is no monarch", Crown of Gondor, Archivist of
	// Gondor). It is a live game-state predicate with no clause parameters.
	ConditionPredicateNoMonarch
	// ConditionPredicateDefendingPlayerIsMonarch is satisfied when the defending
	// player of an attack currently holds the monarch (CR 720), as in "can't
	// attack unless defending player is the monarch" (Crown-Hunter Hireling). Like
	// ConditionPredicateDefendingPlayerControls, it is only meaningful as the guard
	// on a can't-attack static rule, where the defending player is resolved per
	// attack.
	ConditionPredicateDefendingPlayerIsMonarch
	// ConditionPredicateThatPlayerIsMonarch is satisfied when the player named by
	// a preceding "its controller" reference currently holds the monarch (CR 720),
	// as in "doesn't untap during its controller's untap step unless that player is
	// the monarch" (Fall from Favor). Like ConditionPredicateDefendingPlayerIsMonarch
	// it is only meaningful as the guard on a static rule, where the affected
	// permanent's controller is resolved per untap step.
	ConditionPredicateThatPlayerIsMonarch
	// ConditionPredicateControllerHasInitiative is satisfied when the context
	// controller has the initiative (CR 720). It is a live single-player
	// game-state predicate with no clause parameters.
	ConditionPredicateControllerHasInitiative
	// ConditionPredicateControllerHasCityBlessing is satisfied when the context
	// controller has the city's blessing (CR 702.131 ascend). It is a live
	// single-player game-state predicate with no clause parameters.
	ConditionPredicateControllerHasCityBlessing
	// ConditionPredicateControllerTurn is satisfied while it is the context
	// controller's turn, i.e. the controller is the active player ("During your
	// turn, this creature has first strike"; Fresh-Faced Recruit, Embereth
	// Skyblazer). It gates a continuous self-static so the granted keyword or
	// power/toughness bonus applies only on the controller's own turns.
	ConditionPredicateControllerTurn
	// ConditionPredicateColoredManaSpentToCastAtLeast is satisfied when at least
	// Threshold mana of ManaSpentColor was spent to cast the resolving spell ("if
	// at least three white mana was spent to cast this spell"; the Adamant
	// ability word on the Throne of Eldraine Paladin cycle). It gates a
	// resolving-spell enters-with-counters replacement against the per-color mana
	// spend captured on the stack object.
	ConditionPredicateColoredManaSpentToCastAtLeast
	// ConditionPredicateSameColorManaSpentToCastAtLeast is satisfied when at least
	// Threshold mana of a single color was spent to cast the resolving spell ("if
	// at least three mana of the same color was spent to cast this spell"; Henge
	// Walker). It compares the largest single-color tally of the captured
	// per-color mana spend against Threshold.
	ConditionPredicateSameColorManaSpentToCastAtLeast
	// ConditionPredicateControllerGraveyardPermanentCardCountAtLeast is satisfied
	// when at least Threshold permanent cards are in the context controller's
	// graveyard ("as long as there are four or more permanent cards in your
	// graveyard"; the Descend ability word on Basking Capybara, Frilled
	// Cave-Wurm, Didact Echo). It gates a continuous self-static against the
	// count of permanent-type cards in the controller's graveyard.
	ConditionPredicateControllerGraveyardPermanentCardCountAtLeast
	// ConditionPredicateControllerGraveyardManaValueCountAtLeast is satisfied when
	// there are at least Threshold distinct mana values among cards in the context
	// controller's graveyard ("as long as there are five or more mana values among
	// cards in your graveyard"; Syndicate Infiltrator, Aven Heartstabber). It
	// gates a continuous self-static against the number of distinct mana values in
	// the controller's graveyard.
	ConditionPredicateControllerGraveyardManaValueCountAtLeast
	// ConditionPredicateAnyOpponentGraveyardCardCountAtLeast is satisfied when at
	// least one opponent has at least Threshold cards in their graveyard ("as long
	// as an opponent has eight or more cards in their graveyard"; Nimana
	// Skitter-Sneak, Jace's Phantasm, Thieves' Guild Enforcer). It gates a
	// continuous self-static against the largest opponent graveyard size.
	ConditionPredicateAnyOpponentGraveyardCardCountAtLeast
	// ConditionPredicateEventSpellManaSpentToCastAtLeast is satisfied when at
	// least Threshold total mana was spent to cast a spell-cast trigger's
	// triggering spell ("if at least four mana was spent to cast it"; Blazing
	// Bomb, Sahagin, Prompto Argentum, Raggadragga). It reads the mana spent
	// recorded on the triggering event, so it resolves only in a spell-cast
	// trigger's intervening-if context.
	ConditionPredicateEventSpellManaSpentToCastAtLeast
	// ConditionPredicateEventSpellManaSpentToCastAtMost is satisfied when at most
	// Threshold total mana was spent to cast a spell-cast trigger's triggering
	// spell. With Threshold zero it backs "if no mana was spent to cast it"
	// (Boromir, Lavinia, Roiling Vortex), the free-spell punisher gate. It reads
	// the mana spent recorded on the triggering event.
	ConditionPredicateEventSpellManaSpentToCastAtMost
	// ConditionPredicateTriggeringPlayerHandSizeAtMost is satisfied when the
	// triggering player has at most Threshold cards in hand ("if that player has
	// two or fewer cards in hand"; with Threshold zero, "if that player has no
	// cards in hand"). It reads the hand of the player recorded on the triggering
	// step event, so it resolves only in a phase/step trigger's intervening-if
	// context (each opponent's or each player's upkeep).
	ConditionPredicateTriggeringPlayerHandSizeAtMost
	// ConditionPredicateTriggeringPlayerHandSizeAtLeast is satisfied when the
	// triggering player has at least Threshold cards in hand ("if that player has
	// five or more cards in hand"). It reads the hand of the player recorded on
	// the triggering step event and resolves only in a phase/step trigger's
	// intervening-if context.
	ConditionPredicateTriggeringPlayerHandSizeAtLeast
	// ConditionPredicateLandEnteredThisTurnOrControlsBasic is satisfied when the
	// source land entered the battlefield this turn or its controller controls a
	// basic land ("Activate only if this land entered this turn or if you control
	// a basic land."; the Mercadian Masques tap-for-two-colors land cycle). It
	// gates the dual-mana ability on either disjunct.
	ConditionPredicateLandEnteredThisTurnOrControlsBasic
	// ConditionPredicateDefendingPlayerControls is satisfied when the defending
	// player of an attack controls at least one permanent matching Selection
	// ("defending player controls an Island", Sea Monster). It is only supported
	// as the negated guard on a can't-attack static rule, where the defending
	// player is resolved per attack; the closed vocabulary carries no defending
	// player anywhere else, so every other use fails closed downstream.
	ConditionPredicateDefendingPlayerControls
	// ConditionPredicateDefendingPlayerDoesNotPay is the failure gate of an
	// attack-triggered defending-player optional-payment sequence ("defending
	// player may pay {N}. If that player doesn't, <consequence>."). It is the
	// defending-player counterpart of ConditionPredicateEventPlayerDoesNotPay,
	// satisfied when the offered payment is declined (Shrouded Serpent).
	ConditionPredicateDefendingPlayerDoesNotPay
	// ConditionPredicateNoLifeLostThisWay is satisfied when the immediately
	// preceding lose-life effect caused no life loss ("if no life is lost this
	// way"). It is the failure complement of the "if you do" resolving-success
	// gate: the gated effect runs only when the prior lose-life effect published
	// a failed result (CR 608.2c). It backs Blitzwing, Cruel Tormentor's
	// end-step convert. Added last so existing kinds keep their wire values.
	ConditionPredicateNoLifeLostThisWay
	// ConditionPredicateSourceAbilityResolutionOrdinalThisTurn is satisfied when
	// the resolving triggered ability has resolved exactly Threshold times this
	// turn, counting the current resolution ("if this is the second time this
	// ability has resolved this turn"; Prowl, Pursuit Vehicle). Added last so
	// existing kinds keep their wire values.
	ConditionPredicateSourceAbilityResolutionOrdinalThisTurn
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

// GraveyardRedirectControlScope identifies, for a "would die" card-to-graveyard
// replacement, who controls the dying permanent the replacement watches.
type GraveyardRedirectControlScope uint8

// Graveyard redirect control scopes recognized by the semantic compiler.
const (
	GraveyardRedirectControlScopeAny GraveyardRedirectControlScope = iota
	GraveyardRedirectControlScopeYou
	GraveyardRedirectControlScopeOpponent
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

// ConditionAttachment is a closed semantic attachment-state selection filter
// testing whether the matched permanent has an Aura ("enchanted") or Equipment
// ("equipped") attached to it.
type ConditionAttachment uint8

// Condition attachment-state values.
const (
	ConditionAttachmentNone ConditionAttachment = iota
	ConditionAttachmentEnchanted
	ConditionAttachmentEquipped
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
	// ConditionComparisonScopeTriggeringPlayer counts the permanents of the
	// player tied to the triggering event ("that player").
	ConditionComparisonScopeTriggeringPlayer
)

// ConditionSelection is the source-independent Selection vocabulary used by
// semantic conditions. Subtype names are canonicalized during recognition.
type ConditionSelection struct {
	RequiredTypes []types.Card
	Supertypes    []types.Super
	SubtypesAny   []string
	ColorsAny     []color.Color
	Colorless     bool
	Multicolored  bool
	TokenOnly     bool
	ExcludeSource bool
	Tapped        ConditionTriState
	CombatState   ConditionCombatState
	// Attachment tests whether the matched permanent has an Aura ("enchanted")
	// or Equipment ("equipped") attached to it. Its zero value imposes no
	// attachment requirement.
	Attachment        ConditionAttachment
	Keyword           parser.KeywordKind
	PowerAtLeast      int
	MatchPowerAtLeast bool
	// TotalPowerAtLeast is the collective-power threshold for a "have total
	// power <n> or greater" qualifier. MatchTotalPowerAtLeast marks it present.
	TotalPowerAtLeast      int
	MatchTotalPowerAtLeast bool
	// TotalPowerAtMost is the collective-power ceiling for a "have total power
	// <n> or less" qualifier, the upper-bound counterpart of TotalPowerAtLeast.
	// MatchTotalPowerAtMost marks it present.
	TotalPowerAtMost      int
	MatchTotalPowerAtMost bool
	// DistinctNamesAtLeast is the distinct-name threshold for a "with different
	// names" qualifier. MatchDistinctNamesAtLeast marks it present.
	DistinctNamesAtLeast      int
	MatchDistinctNamesAtLeast bool
	// DamageRecipientOpponent, DamageNoncombatOnly, and DamageSourceAnyController
	// qualify a damage-by-controlled-source clause: opponent-only recipient,
	// noncombat-only damage, and a source controlled by any player respectively.
	DamageRecipientOpponent           bool
	DamageRecipientOpponentPlayerOnly bool
	DamageNoncombatOnly               bool
	DamageSourceAnyController         bool
	// DamageRecipientController restricts a damage-by-source clause to damage
	// dealt to the source permanent's controller alone ("would deal damage to
	// you"). DamageSourceControllerOpponent restricts it to a source controlled
	// by an opponent ("a source an opponent controls"). They back the continuous
	// static damage-prevention statics.
	DamageRecipientController      bool
	DamageSourceControllerOpponent bool
	// DamageRecipientSelf and DamageRecipientAttached qualify a
	// ConditionPredicateDamageWouldBeDealtToPermanent clause: the damaged
	// permanent is the ability's own source or the permanent it is attached to.
	DamageRecipientSelf     bool
	DamageRecipientAttached bool
	// DamageRecipientMonarchGate marks a DamageWouldBeDealtToPermanent clause
	// gated by "... while you're the monarch" (Jared Carthalion).
	DamageRecipientMonarchGate bool
	// AnyCounter requires the matched permanent to carry at least one counter of
	// any kind ("if this permanent has counters on it").
	AnyCounter bool
	// CounterKind, CounterKindKnown, CounterCountAtLeast, and CounterCountLessThan
	// express a named-counter-count threshold the matched permanent must satisfy
	// ("has seven or more quest counters on it", "has fewer than three +1/+1
	// counters on it"). CounterKindKnown marks the kind present; CounterCountAtLeast
	// carries an inclusive minimum count (>=) and CounterCountLessThan an exclusive
	// maximum count (<). At most one bound is non-zero.
	CounterKind          counter.Kind
	CounterKindKnown     bool
	CounterCountAtLeast  int
	CounterCountLessThan int
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
	GraveyardSubjectTypesAny     []types.Card
	GraveyardFromBattlefieldOnly bool

	// GraveyardRedirectControlScope carries the control qualifier of a "would
	// die" ConditionPredicateCardWouldGoToGraveyard clause: who controls the
	// dying permanent the replacement watches. It is GraveyardRedirectControlScopeAny
	// for owner-scoped "would be put into a graveyard" forms.
	GraveyardRedirectControlScope GraveyardRedirectControlScope

	// CounterRecipientTypesAny carries the type-union recipient filter of a
	// ConditionPredicateCounterPlacementOnControlledPermanent clause ("an
	// artifact or creature you control", Ozolith, the Shattered Spire). It is
	// empty for the unrestricted "a permanent you control" form.
	CounterRecipientTypesAny []types.Card

	// CounterRecipientExcludesSource drops the source permanent from a
	// ConditionPredicateCounterPlacementOnControlledPermanent clause's recipient
	// match ("another creature you control", Benevolent Hydra). It is false for
	// recipient forms that include the source.
	CounterRecipientExcludesSource bool

	// GraveyardCountCardType carries the single card type counted by a
	// ConditionPredicateControllerGraveyardCardOfTypeCountAtLeast clause ("if
	// twenty or more creature cards are in your graveyard", Mortal Combat).
	// Threshold carries the minimum count. It is the empty card type for other
	// clauses.
	GraveyardCountCardType types.Card

	// ControlledNames carries the card names required by a
	// ConditionPredicateControllerControlsNamed clause ("If you control an
	// Urza's Mine and an Urza's Tower, ..."). The controller must control a
	// permanent matching each listed name.
	ControlledNames []string

	// ManaSpentColor carries the color required by a
	// ConditionPredicateColoredManaSpentToCastAtLeast clause ("if at least three
	// white mana was spent to cast this spell"; the Adamant ability word).
	// Threshold carries the minimum amount of that color of mana. It is the empty
	// color for the same-color form, which compares the largest single-color
	// tally instead of a named color.
	ManaSpentColor color.Color

	// Reflexive marks a prior-instruction-accepted gate that comes from a
	// reflexive "When you do," preamble (CR 603.11) rather than an immediate "If
	// you do," rider. Lowering reads this flag to route the gated consequence to
	// a reflexive triggered ability whose targets are chosen after the enabling
	// action, instead of resolving it inline with up-front targets.
	Reflexive bool
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
	// ControllerThatPlayer restricts a target to permanents controlled by the
	// triggering event's player ("target creature that player controls",
	// Garland, Royal Kidnapper). The lowering maps it to the runtime
	// Selection.ControlledByEventPlayer predicate.
	ControllerThatPlayer
	// ControllerDefendingPlayer restricts a target to permanents controlled by
	// the defending player of the triggering attack ("goad target creature
	// defending player controls", Coveted Peacock). The lowering maps it to the
	// runtime Selection.ControlledByDefendingPlayer predicate.
	ControllerDefendingPlayer
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
	// MatchTotalManaValue records whether TotalManaValue bounds the combined mana
	// value of the chosen set ("with total mana value N or less") rather than each
	// matched card's own mana value (MatchManaValue). It lowers to the runtime
	// ChooseFromZone Riders.MaxTotalManaValue cap.
	MatchTotalManaValue bool
	TotalManaValue      compare.Int
	// ManaValueX records that the MatchManaValue bound is the spell's chosen {X}
	// ("with mana value X or less") rather than a fixed number; ManaValue then
	// holds only the operator. It lowers to SearchSpec.MaxManaValueFromX.
	ManaValueX bool
	// ManaValueDynamic records a "with mana value less than or equal to the
	// amount of life you (lost|gained) this turn" bound (Betor, Ancestor's
	// Voice), whose upper bound is a turn-event life total rather than a fixed
	// number. It is independent of MatchManaValue/ManaValueX and lowers to the
	// runtime Selection.ManaValueDynamic predicate. DynamicAmountNone means no
	// dynamic bound.
	ManaValueDynamic DynamicAmountKind
	// ManaValueDynamicCount records a "with mana value less than or equal to the
	// number of <permanents> you control" bound (Beseech the Queen — number of
	// lands you control), whose upper bound is a controlled-permanent count
	// rather than a fixed number. It carries the counted-subject amount and is
	// mutually exclusive with ManaValueDynamic (life totals); it lowers to the
	// runtime Selection.ManaValueDynamic predicate with a DynamicAmountCountSelector
	// group. Nil means no count-based dynamic bound.
	ManaValueDynamicCount *CompiledAmount
	// ManaValueSacrificedCost records a "with mana value X or less, where X is N
	// plus the sacrificed creature's mana value" bound (Eldritch Evolution),
	// whose upper bound is the mana value of the creature sacrificed to pay the
	// spell's additional cost plus the fixed addend it carries. It lowers to
	// SearchSpec.MaxManaValueFromSacrificedCost, resolved as the search runs. Nil
	// means no sacrificed-cost bound.
	ManaValueSacrificedCost *int
	Power                   compare.Int
	MatchPower              bool
	Toughness               compare.Int
	MatchToughness          bool
	Colorless               bool
	Multicolored            bool
	// Colored records a "one or more colors" qualifier ("permanents ... that are
	// one or more colors", All Is Dust). It is the complement of Colorless and
	// lowers to Selection.Colored.
	Colored       bool
	BasicLandType bool
	// Historic records a "historic" card qualifier ("target historic card from
	// your graveyard"). A historic card is an artifact, a legendary, or a Saga
	// (CR 702.61b); the cross-category disjunction is kept as its own flag and
	// lowers to a Selection.AnyOf of those three alternatives.
	Historic bool
	// MatchCounter records whether RequiredCounter is active ("creature you
	// control with a +1/+1 counter on it"); RequiredCounter names the counter
	// kind the matched permanent must carry. MatchAnyCounter records the
	// kind-agnostic "with a counter on it" qualifier, matching a permanent
	// carrying a counter of any kind. MatchNoCounters records the kind-agnostic
	// "with no counters on it/them" qualifier, matching a permanent carrying no
	// counters of any kind.
	MatchCounter    bool
	RequiredCounter counter.Kind
	MatchAnyCounter bool
	MatchNoCounters bool
	// MatchExcludedCounter records the kind-specific negated "without a <kind>
	// counter on it/them" qualifier ("each creature without a +1/+1 counter on
	// it"); ExcludedCounter names the counter the matched permanent must not
	// carry. Unlike MatchNoCounters it is kind-specific rather than
	// kind-agnostic.
	MatchExcludedCounter bool
	ExcludedCounter      counter.Kind
	// PlayerOrPlaneswalker marks the combined "player or planeswalker" /
	// "opponent or planeswalker" combined damage target. Kind stays
	// SelectorPlayer or SelectorOpponent; this flag records the additional
	// planeswalker-permanent half the merged Kind cannot express.
	PlayerOrPlaneswalker bool
	// SubtypeFromEntryChoice requires each matched permanent to share the creature
	// subtype the source permanent chose as it entered ("creatures you control of
	// the chosen type"). It lowers to Selection.SubtypeFromSourceEntryChoice.
	SubtypeFromEntryChoice bool
	// ColorFromEntryChoice requires each matched object to share the color the
	// source permanent chose as it entered ("of the chosen color", Prism Ring,
	// Heraldic Banner). It lowers to Selection.ColorChoice = ColorChoiceSourceEntry.
	ColorFromEntryChoice bool
	// SubtypeFromChosenType requires each matched permanent to share the creature
	// subtype chosen earlier in the same resolution by a "Choose a creature type."
	// effect ("each permanent you control of that type"). It lowers to
	// Selection.SubtypeFromChosenType (which reads game.SpellChosenTypeChoiceKey).
	SubtypeFromChosenType bool
	// SubtypeFromChosenTypeExcluded requires each matched permanent to NOT share
	// the creature subtype chosen earlier in the same resolution by a "Choose a
	// creature type." effect ("all creatures that aren't of the chosen type",
	// Kindred Dominance). It lowers to game.SubtypeChoiceResolutionExcluded.
	SubtypeFromChosenTypeExcluded bool
	// ConjunctiveTypes records that a multi-member RequiredTypesAny names types a
	// permanent must carry all at once ("artifact creature") rather than any one
	// of ("artifact or creature"). It lowers the type set to the conjunctive
	// TargetPredicate.PermanentTypesAll filter instead of PermanentTypes.
	ConjunctiveTypes bool
	// RequiredName carries the verbatim card name of a "named <Name>" library
	// search filter ("a card named Trustworthy Scout"). It lowers to
	// SearchSpec.Name; the parser owns the wording, so the compiler only copies it.
	RequiredName string
	// EnteredThisTurn requires each matched permanent to have entered the
	// battlefield this turn ("each green creature that entered this turn"). It
	// lowers to Selection.EnteredThisTurn.
	EnteredThisTurn bool
	// DealtDamageThisTurn requires each matched permanent to have been dealt
	// damage this turn ("target creature that was dealt damage this turn", Fatal
	// Blow). It lowers to Selection.DealtDamageThisTurn.
	DealtDamageThisTurn bool
	// Modified requires each matched permanent to be modified (a counter, Aura,
	// or Equipment attached; CR 701.50) for "target modified creature you
	// control" (Silver Sable). Enchanted requires one or more Auras attached
	// ("target enchanted permanent", Cut the Earthly Bond); Equipped requires one
	// or more Equipment attached. They lower to Selection.MatchModified /
	// Selection.MatchEnchanted / Selection.MatchEquipped.
	Modified  bool
	Enchanted bool
	Equipped  bool
	// PowerLessThanSource requires each matched permanent's power to be strictly
	// less than the ability's source permanent's power ("target attacking
	// creature with lesser power", Mentor); PowerGreaterThanSource is the
	// "with greater power" sibling. They are source-relative, so unlike
	// Power/MatchPower they carry no fixed comparison and lower to
	// Selection.PowerLessThanSource / Selection.PowerGreaterThanSource.
	PowerLessThanSource    bool
	PowerGreaterThanSource bool
	// ManaValueLessThanEventPermanent requires each matched card's mana value to
	// be strictly less than the triggering event permanent's mana value ("return
	// target Cleric card with lesser mana value from your graveyard", Orah,
	// Skyclave Hierophant, where the bound is the creature that died). It is the
	// event-relative mana-value analogue of PowerLessThanSource: the bound reads
	// the triggering event's permanent, not the ability's source, and lowers to
	// Selection.ManaValueLessThanEventPermanent.
	ManaValueLessThanEventPermanent bool
	// NameUniqueAmongControlled requires the matched permanent's name to differ
	// from every other permanent its controller controls ("target enchantment
	// you control that doesn't have the same name as another permanent you
	// control", Yenna, Redtooth Regent). It lowers to
	// Selection.NameUniqueAmongControlled.
	NameUniqueAmongControlled bool
	// InclusiveOneOfEach records that the selection joined two or more singular
	// articled card nouns with "and/or" ("a Saga card and/or a land card"),
	// meaning up to one card of each named type may be chosen rather than a
	// single card matching any one of them. The merged RequiredTypesAny /
	// SubtypesAny carry the named types; the lowering realizes one independent
	// optional pick per named type for the put-from-among-onto-battlefield shape.
	InclusiveOneOfEach bool
	// SingleGraveyard records a "from a single graveyard" qualifier on a
	// graveyard-card target ("Exile up to three target cards from a single
	// graveyard"), requiring every chosen card to lie in one and the same
	// graveyard. It lowers to TargetSpec.SameGraveyard.
	SingleGraveyard bool
	// SameNameGroup records a "and all other <type> with the same name as that
	// <noun>" group attached to a single destroy target ("Destroy target nonland
	// permanent and all other permanents with the same name as that permanent",
	// Maelstrom Pulse). GroupTypes carries the printed group card types (empty
	// for "permanents", meaning no type restriction). It lowers to a
	// SameNamePermanentGroup on the destroy primitive.
	SameNameGroup *CompiledSameNameGroup
	// SpellTargetRestrictions records a "Counter target spell that targets <X>"
	// restriction, requiring the matched spell to have a chosen target satisfying
	// one of these alternatives. Each names either a permanent (by card types and
	// controller relation) or a player (by relation). It lowers to
	// TargetPredicate.SpellTargets.
	SpellTargetRestrictions []CompiledSpellTargetRestriction
	Alternatives            []CompiledSelector
	atoms                   *CompiledSelectorAtoms
}

// CompiledSpellTargetRestriction is one alternative of a counter spell target's
// "that targets <X>" restriction. A player alternative sets IsPlayer with a
// player relation in Controller; a permanent alternative leaves IsPlayer false
// and names optional required card types with a controller relation. Relations
// are relative to the player choosing the counter target.
type CompiledSpellTargetRestriction struct {
	IsPlayer       bool
	PermanentTypes []types.Card
	Controller     ControllerKind
}

// CompiledSameNameGroup is the compiled form of a destroy target's "and all
// other <type> with the same name as that <noun>" group. GroupTypes carries the
// printed group card types and is empty for the "permanents" wording, meaning no
// card-type restriction (the runtime relies on shared name, which implies a
// shared card, so the printed type is fidelity only).
type CompiledSameNameGroup struct {
	GroupTypes []types.Card
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

// WithAtoms returns a copy of the selector carrying the given parser-owned atom
// filters. Production compilation populates these through the package-internal
// append/set helpers; WithAtoms exposes the same wiring for callers and tests
// that need to build a CompiledSelector directly from its atom-derived filters.
func (s CompiledSelector) WithAtoms(atoms CompiledSelectorAtoms) CompiledSelector {
	clone := atoms
	s.atoms = &clone
	return s
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
	EffectGainControl       // gain control of [target permanent]
	EffectGainPlayerCounter // controller gains {E} energy counters
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
	EffectDirectedMustAttack
	EffectMustBeBlocked
	// EffectMustBeBlockedByAllAble is the true-lure requirement ("All creatures
	// able to block this creature do so.").
	EffectMustBeBlockedByAllAble
	// EffectAssignDamageAsUnblocked is the permission to assign combat damage as
	// though unblocked ("You may have this creature assign its combat damage as
	// though it weren't blocked.").
	EffectAssignDamageAsUnblocked
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
	EffectTapOrUntap
	EffectUntap
	EffectTransform
	EffectLifeTotalCantChange
	EffectProtectionFromEverything
	EffectPhaseOut
	EffectImpulseExile
	EffectCreateEmblem
	EffectAdditionalLandPlays
	EffectLoseGame
	EffectChooseNewTargets
	EffectCastAsThoughFlash
	EffectPlayFromLibraryTop
	EffectPlay
	EffectCantCastSpells
	EffectSpellCostModifier
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
	EffectDevour
	EffectRenown
	EffectTribute
	EffectChooseCreatureType
	EffectNoMaximumHandSize
	EffectAdditionalCombatPhase
	EffectLookAtHand
	// EffectRollDie rolls a single fair die with DieSides faces and publishes
	// the rolled value as its resolved amount ("roll a d20"). A later effect
	// consumes the value via a DynamicAmountDieRollResult amount. Added last so
	// existing kinds keep their wire values.
	EffectRollDie
	EffectRemoveFromCombat
	// EffectExileIfLeaveBattlefield models the leaves-the-battlefield exile
	// self-replacement "If it would leave the battlefield, exile it instead of
	// putting it anywhere else." (Whip of Erebos). It lowers to a
	// CreateReplacement bound to the affected object (a back-referenced object
	// for "it", or the source for "this <type>"). Added last so existing kinds
	// keep their wire values.
	EffectExileIfLeaveBattlefield
	EffectCantBlockAndCantBeBlocked
	// EffectBecomeType adds one or more card types to a targeted permanent for a
	// duration ("Target permanent becomes an artifact in addition to its other
	// types until end of turn.", Liquimetal Torque). It lowers to an
	// ApplyContinuous at LayerType. Added last so existing kinds keep their wire
	// values.
	EffectBecomeType
	// EffectRemoveCounter removes a fixed number of counters from a single
	// recognized target permanent ("Remove a counter from target permanent.",
	// Ferropede; "Remove a counter from target nonland permanent.", Thrull
	// Parasite). It lowers to a RemoveCounter primitive; the kind-unspecified "a
	// counter" form removes a counter of a kind the controller chooses. Added
	// last so existing kinds keep their wire values.
	EffectRemoveCounter
	EffectBecomeMonarch
	EffectCantBecomeMonarch
	// EffectDelayedTrigger creates an event-based delayed triggered ability that
	// fires on a matching game event within a bounded window ("Whenever you cast
	// a spell this turn, ...", Showdown of the Skalds; "When you next cast a
	// creature spell this turn, ...", Summon: Fenrir). It lowers to a
	// game.CreateDelayedTrigger carrying the nested ability's trigger pattern and
	// content. Added last so existing kinds keep their wire values.
	EffectDelayedTrigger
	// EffectRingTempts models the fixed designation effect "The Ring tempts
	// you." (CR 701.51): the resolving controller gets the Ring emblem, advances
	// it to its next level, and chooses a creature they control as their
	// Ring-bearer. It lowers to a game.RingTempts primitive scoped to the
	// controller. Added last so existing kinds keep their wire values.
	EffectRingTempts
	// EffectPolymorph sets a targeted creature's color, creature type, and base
	// power/toughness while removing all of its abilities until end of turn
	// ("Until end of turn, target creature loses all abilities and becomes a
	// <color> <subtype> with base power and toughness N/N.", Turn to Frog). It
	// lowers to an ApplyContinuous across the ability, color, type, and
	// power/toughness layers. Added last so existing kinds keep their wire
	// values.
	EffectPolymorph
	// EffectAttackTax is the resolving, duration-bounded attack-tax effect
	// "Until your next turn, creatures can't attack you unless their controller
	// pays {N} for each of those creatures." (Summon: Yojimbo chapters II/III).
	// AttackTaxGeneric carries the per-attacker generic mana N; lowering installs
	// a RuleEffectAttackTax for the recognized duration. Added last so existing
	// kinds keep their wire values.
	EffectAttackTax
	// EffectAdapt models the Adapt keyword action (CR 701.43) written out as an
	// activated ability effect ("Adapt N."): if the source creature has no
	// +1/+1 counters on it, it gets N +1/+1 counters. It lowers to a game.Adapt
	// primitive scoped to the source permanent; the runtime guard subsumes the
	// printed "if it has no +1/+1 counters" reminder. Added last so existing
	// kinds keep their wire values.
	EffectAdapt
	// EffectLookAtLibraryTop models the one-shot peek "look at the top card of
	// your library." (the Kinship ability word's leading instruction). The
	// controller privately sees the top card as the ability resolves, conveying
	// hidden information without moving it. It lowers to a game.LookAtLibraryTop
	// primitive. Added last so existing kinds keep their wire values.
	EffectLookAtLibraryTop
	// EffectConnive models the connive keyword action (CR 702.154): the
	// conniving permanent's controller draws N cards, then discards N cards, and
	// a +1/+1 counter is placed on that permanent for each nonland card discarded
	// this way. It lowers to a game.Connive primitive scoped to the source
	// permanent. Added last so existing kinds keep their wire values.
	EffectConnive
	// EffectSetBasePT is the one-shot continuous base power/toughness SET on a
	// group, single target, or the source, optionally adding every creature type,
	// until end of turn ("{X}: Until end of turn, creatures you control have base
	// power and toughness X/X and gain all creature types.", Mirror Entity;
	// "Target creature has base power and toughness 4/4 until end of turn.",
	// Square Up). It lowers to an ApplyContinuous at the power/toughness set layer
	// and, when riding the every-creature-type grant, the type layer. Added last
	// so existing kinds keep their wire values.
	EffectSetBasePT
	// EffectPayRepeatedlyAnimate is the kicker-on-resolution land-animation
	// trigger of Primal Adversary: the controller may pay a repeatable mana cost
	// any number of times, then puts that many +1/+1 counters on the source and
	// animates up to that many lands they control into creatures with a set
	// power/toughness, added subtype(s), and keyword(s) while they remain lands.
	// It lowers to a PayRepeatedly publishing the payment count, an AddCounter
	// sized by the count, and an ApplyContinuous that chooses up to that many
	// controlled lands. Added last so existing kinds keep their wire values.
	EffectPayRepeatedlyAnimate
	// EffectSwitchPT is the one-shot continuous "switch power and toughness until
	// end of turn" effect (CR 613.4e, layer 7e) on the source or a single target
	// creature ("Switch this creature's power and toughness until end of turn.",
	// Aeromoeba; "Switch target creature's power and toughness until end of
	// turn.", Twisted Image). It lowers to an ApplyContinuous at
	// LayerPowerToughnessSwitch until end of turn. Added last so existing kinds
	// keep their wire values.
	EffectSwitchPT
	// EffectExileIfWouldDieThisTurn models the damage-spell rider "If that
	// creature [or planeswalker] would die this turn, exile it instead." (Lava
	// Coil, Obliterating Bolt, Magma Spray, Flame-Blessed Bolt, ...). It rides on
	// a single-target damage or -X/-X spell and redirects the targeted
	// permanent's death (a battlefield-to-graveyard zone change) to exile for the
	// rest of the turn. "That creature"/"it" bind to the spell's single target,
	// so it lowers to a CreateReplacement bound to that target for the turn. Added
	// last so existing kinds keep their wire values.
	EffectExileIfWouldDieThisTurn
	// EffectCanBlockOnlyCreaturesWithFlying is the blocker-side permission
	// restriction "can block only creatures with flying" (Cloud Sprite,
	// Gloomwidow): the subject creature may block only attackers that have flying.
	// It lowers to the can-block-only runtime rule effect bounded by the flying
	// blocker restriction. Added last so existing kinds keep their wire values.
	EffectCanBlockOnlyCreaturesWithFlying
	// EffectCanBlockAdditional is the blocker-side capability "can block an
	// additional creature each combat" (Brave the Sands, Coastline Chimera): the
	// subject creature may block one more attacker than the usual single blocker
	// limit. It lowers to the can-block-additional runtime rule effect. Added last
	// so existing kinds keep their wire values.
	EffectCanBlockAdditional
	// EffectAnimateSelf is the one-shot continuous self-animation "This
	// <land|artifact|creature|permanent> becomes a N/N [<color>...] [artifact]
	// <subtype>... creature [with <keyword>...|all creature types] until end of
	// turn." (Faerie Conclave, the Keyrune mana rocks, Mutavault): the source
	// gains the creature card type (plus the artifact type when stated), the
	// named subtypes (or every creature type), the stated colors, the granted
	// keywords, and the literal base power/toughness until end of turn while
	// keeping its existing types. It lowers to a single ApplyContinuous over the
	// source for the turn. Added last so existing kinds keep their wire values.
	EffectAnimateSelf
	// EffectCantAttackAlone, EffectCantBlockAlone, and EffectCantAttackOrBlockAlone
	// are the static-only combat "alone" restrictions ("can't attack alone",
	// "can't block alone", "can't attack or block alone"); they never appear as a
	// resolving spell effect and pair only with a static rule declaration. Added
	// last so existing kinds keep their wire values.
	EffectCantAttackAlone
	EffectCantBlockAlone
	EffectCantAttackOrBlockAlone
	// EffectCanAttackAsThoughDefender is the self-permission "This creature can
	// attack this turn as though it didn't have defender." (Glade Watcher, Mirror
	// Wall, Walking Wall): the source creature may be declared as an attacker this
	// turn despite having defender. It lowers to a this-turn ApplyRule over the
	// source. Added last so existing kinds keep their wire values.
	EffectCanAttackAsThoughDefender
	EffectAnimateTarget
	// EffectCantBeBlockedExceptBy is the static combat-evasion restriction "can't
	// be blocked except by <quality>" (Dread Warlock, Silhana Ledgewalker, Noggle
	// Bandit): every blocker that does not match the named characteristic is
	// prohibited from blocking the source. It is the complement of
	// EffectCantBeBlockedByCreaturesWith and pairs only with a static rule
	// declaration carrying the allowed BlockerRestriction. Added last so existing
	// kinds keep their wire values.
	EffectCantBeBlockedExceptBy
	// EffectAssignsCombatDamageByToughness is the static combat-damage replacement
	// "<subject> assigns combat damage equal to its toughness rather than its
	// power." (Doran, the Siege Tower; Assault Formation; Belligerent Brontodon):
	// the affected creatures assign combat damage equal to their toughness instead
	// of their power. It pairs only with a static rule declaration. Added last so
	// existing kinds keep their wire values.
	EffectAssignsCombatDamageByToughness
	// EffectBecomeColor is the one-shot continuous color-set "<subject> becomes
	// <color>... until end of turn." (Cerulean Wisps, Niveous Wisps, Raging
	// Spirit): the named colors SET the subject's color set, or the "colorless"
	// form clears it, until end of turn. The subject is the source or a single
	// target. It lowers to an ApplyContinuous at LayerColor. Added last so
	// existing kinds keep their wire values.
	EffectBecomeColor
	// EffectCantAttackOrBlockAndCantActivate is the Arrest-family pinning
	// prohibition "Enchanted creature can't attack or block, and its activated
	// abilities can't be activated." (Arrest, Lawmage's Binding, Planar
	// Disruption): the affected permanent can't attack or block and none of its
	// activated abilities can be activated. It pairs only with a static rule
	// declaration and lowers to the can't-attack, can't-block, and permanent-
	// scoped can't-activate-abilities runtime rule effects. Added last so existing
	// kinds keep their wire values.
	EffectCantAttackOrBlockAndCantActivate
	// EffectCantAttackOrBlockAndCantActivateNonMana is the mana-exempt Arrest-
	// family variant "Enchanted permanent can't attack or block, and its activated
	// abilities can't be activated unless they're mana abilities." (Faith's
	// Fetters, Realmbreaker's Grasp): like EffectCantAttackOrBlockAndCantActivate,
	// except the permanent's mana abilities can still be activated. It pairs only
	// with a static rule declaration. Added last so existing kinds keep their wire
	// values.
	EffectCantAttackOrBlockAndCantActivateNonMana
	// EffectCloak is the Cloak keyword action (CR 701.56): put the top card of
	// your library onto the battlefield face down as a 2/2 creature with ward
	// {2}. Added last so existing kinds keep their wire values.
	EffectCloak
	// EffectGoad is the goad keyword action (CR 701.38): the goaded creature
	// attacks each combat if able and attacks a player other than its
	// controller if able, until that player's next turn. Added last so existing
	// kinds keep their wire values.
	EffectGoad
	// EffectMonstrosity is the monstrosity keyword action (CR 701.32) written out
	// as an activated ability effect ("Monstrosity N."): if the source creature
	// isn't monstrous, it gets N +1/+1 counters and becomes monstrous. It lowers
	// to a game.Monstrosity primitive scoped to the source permanent; the runtime
	// guard subsumes the printed "if this creature isn't monstrous" reminder.
	// Added last so existing kinds keep their wire values.
	EffectMonstrosity
	// EffectChooseExiledCard is the resolution-time choice "Choose an exiled card
	// an opponent owns with a <kind> counter on it." (Dauthi Voidwalker): the
	// resolving controller picks one card resting in exile that a scoped player
	// owns and that bears the named exile marker counter. It carries the source
	// zone (FromZone = Exile), the owner scope (ChooseExiledCardOwnerOpponent),
	// and the marker-counter filter (CounterKind/CounterKindKnown). It pairs at
	// lowering with a following EffectPlay back-reference into a single
	// PlayChosenExiledCard primitive and never lowers on its own. Added last so
	// existing kinds keep their wire values.
	EffectChooseExiledCard
	// EffectReturnExiledCardsWithCounter is the resolution-time mass return "Put
	// all exiled cards you own with <kind> counters on them into your hand."
	// (Flamewar, Brash Veteran): every card the resolving controller owns in
	// exile that bears the named marker counter returns to that controller's
	// hand. It carries the source zone (FromZone = Exile), the destination zone
	// (ToZone = Hand), and the marker-counter filter (CounterKind/
	// CounterKindKnown). It is the return companion to the exile-with-named-
	// counter substrate and lowers on its own to a single mass return. Added last
	// so existing kinds keep their wire values.
	EffectReturnExiledCardsWithCounter
	// EffectBolster is the bolster keyword action (CR 701.37) written out as an
	// ability effect ("Bolster N."): the controller chooses a creature with the
	// least toughness among creatures they control and puts N +1/+1 counters on
	// it. It lowers to a game.Bolster primitive; the chosen creature may be
	// published under a linked key so a later effect (such as "the chosen
	// creature gains trample" or a delayed trigger watching it) can resolve it.
	// Added last so existing kinds keep their wire values.
	EffectBolster
	// EffectCantBeSacrificed is the temporary sacrifice-protection resolving
	// effect "<referenced object> can't be sacrificed this turn." (Slicer, Hired
	// Muscle: "... it can't be sacrificed this turn."): the back-referenced
	// permanent can't be sacrificed until end of turn. It lowers to a this-turn
	// ApplyRule installing a RuleEffectCantBeSacrificed on the source permanent.
	// Added last so existing kinds keep their wire values.
	EffectCantBeSacrificed
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
	// DurationUntilYourNextEndStep matches "until your next end step", the
	// bounded play window Inti, Seneschal of the Sun grants its impulse-exiled
	// card. The effect expires at the controller's next end step.
	DurationUntilYourNextEndStep
	// DurationForAsLongAsThatPlayerIsMonarch matches "for as long as they're the
	// monarch" (Garland, Royal Kidnapper). The gain-control effect expires when
	// the triggering player who became the monarch is no longer the monarch.
	DurationForAsLongAsThatPlayerIsMonarch
	// DurationForAsLongAsExiled matches "for as long as that card remains
	// exiled" (Prowl, Stoic Strategist). The owner-scoped play permission
	// expires when the exiled card leaves exile.
	DurationForAsLongAsExiled
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
	StaticSubjectControlledSagas
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
	StaticSubjectControlledModifiedCreatures
	StaticSubjectOtherControlledTappedCreatures
	StaticSubjectControlledArtifactCreatures
	StaticSubjectOtherControlledArtifactCreatures
	StaticSubjectControlledNontokenCreatures
	StaticSubjectOtherControlledNontokenCreatures
	StaticSubjectAllLands
	StaticSubjectControlledCreaturesChosenType
	StaticSubjectOtherControlledCreaturesChosenType
	StaticSubjectAllCreaturesChosenType
	StaticSubjectOpponentControlledCreaturesChosenType
	StaticSubjectOpponentControlledPermanents
	StaticSubjectOtherAttackingCreatures
	StaticSubjectOtherControlledPermanents
	StaticSubjectControlledNonlegendaryCreatures
	StaticSubjectControlledLands
	StaticSubjectControlledCommanderCreatures
	StaticSubjectControlledCommanders
	StaticSubjectControlledPermanentSubtype
	StaticSubjectOtherControlledPermanentSubtype
	// StaticSubjectNonbasicLands, StaticSubjectNonlandPermanents,
	// StaticSubjectSnowPermanents, and StaticSubjectAllPermanentSubtype name
	// battlefield-wide land and permanent groups for the mass don't-untap
	// restriction ("Nonbasic lands ...", Back to Basics; "Nonland permanents ...",
	// Embargo; "Snow permanents ...", Freyalise's Radiance; "Islands ...", Choke).
	StaticSubjectNonbasicLands
	StaticSubjectNonlandPermanents
	StaticSubjectSnowPermanents
	StaticSubjectAllPermanentSubtype
	// StaticSubjectControlledCreatureSubtypeTokens and its "other" sibling name
	// the controlled creature tokens carrying a named creature subtype ("Zombie
	// tokens you control have <keyword>", the Amass Zombie cycle). The named
	// subtype rides the affected group's SubtypesAny slot and the token state is
	// required; the "other" form excludes the source permanent.
	StaticSubjectControlledCreatureSubtypeTokens
	StaticSubjectOtherControlledCreatureSubtypeTokens
	// StaticSubjectControlledAttackingCreatureSubtype names the attacking
	// creatures of a named subtype the controller controls ("Attacking Vampires
	// you control have ...", Crossway Troublemakers). The subtype rides the
	// affected group's SubtypesAny slot alongside the attacking combat state.
	StaticSubjectControlledAttackingCreatureSubtype
	// StaticSubjectControlledAttackingCreatureTokens names the attacking creature
	// tokens the controller controls ("Attacking tokens you control have ...",
	// Starry-Eyed Skyrider). The token state rides the affected group alongside
	// the attacking combat state.
	StaticSubjectControlledAttackingCreatureTokens
	// StaticSubjectControlledNotOwnedCreatures names the creatures the source's
	// controller controls but does not own ("Creatures you control but don't own
	// get +2/+2 and can't be sacrificed.", Garland, Royal Kidnapper). Its group is
	// the controller-permanents domain narrowed to creatures, and its selection
	// carries the owner-not-controller filter. Added last so existing subjects keep
	// their wire values.
	StaticSubjectControlledNotOwnedCreatures
	// StaticSubjectOtherControlledUntappedCreatures names the untapped creatures
	// the controller controls other than the source ("Other untapped creatures you
	// control have hexproof.", Saryth, the Viper's Fang). Its group is the
	// controller-permanents domain narrowed to creatures with the untapped tap
	// state and the source excluded; it is the untapped-state sibling of
	// StaticSubjectOtherControlledTappedCreatures. Added last so existing subjects
	// keep their wire values.
	StaticSubjectOtherControlledUntappedCreatures
)

// CompiledDamageRecipient bundles the primary-recipient descriptors of a
// deal-damage effect into one typed payload, mirroring the parser's
// DamageRecipientSyntax. Its zero value denotes an ordinary single-target
// recipient with no special routing.
type CompiledDamageRecipient struct {
	// GroupSelectors holds the compiled recipient groups of a dual-recipient
	// fixed group-damage effect ("deals N damage to each X and each Y"). It is
	// empty for single-recipient damage; when present it has exactly two entries
	// that lowering damages in Oracle order.
	GroupSelectors []CompiledSelector
	// Reference marks a damage recipient that is the controller or owner of a
	// referenced object (the prior removal target), as in "deals 2 damage to
	// that land's controller". It is None for every other recipient.
	Reference parser.DamageRecipientReferenceKind
	// EachSourceGroup is the source group of an "each <group> deals N damage to
	// its controller/owner" effect ("Each creature deals 1 damage to its
	// controller."), where every group member is the damage source dealing to
	// the player who controls (or owns) it. EachSourceRole records the per-source
	// recipient role; it is None for every other effect.
	EachSourceGroup CompiledSelector
	EachSourceRole  parser.DamageRecipientReferenceKind
}

// CompiledGroupEntryModification mirrors the parser's typed static group
// entry-modification payload (enters-tapped-group and enters-with-counters-group
// static replacements). Kind names the operation; the tapped form carries the
// controller scope and optional card-type restriction. The with-counters form
// reads its counter from the effect's Selector/CounterKind fields.
type CompiledGroupEntryModification struct {
	Kind            parser.GroupEntryModificationKind
	ControllerScope parser.EntersTappedGroupControllerScope
	Types           []types.Card
}

// CompiledEffect is one recognized instruction verb and the sentence containing
// it. Multiple effects may refer to the same sentence when instructions are
// coordinated.
type CompiledEffect struct {
	Kind           EffectKind
	Context        parser.EffectContextKind
	Connection     parser.EffectConnectionKind
	ConnectionSpan shared.Span
	Span           shared.Span
	ClauseSpan     shared.Span
	Text           string
	VerbSpan       shared.Span
	Player         parser.EffectPlayerKind
	CardSource     parser.EffectCardSourceKind
	// FaceDown mirrors EffectSyntax.FaceDown: a top-of-library exile card source
	// that exiles its cards face down. Lowering threads it onto the
	// ExileTopOfLibrary primitive; it is false for every face-up exile.
	FaceDown             bool
	RequirePermanentCard bool
	// ExileDieSubjectDamagedCreature marks an EffectExileIfWouldDieThisTurn rider
	// whose subject is "a creature dealt damage this way": the would-die exile is
	// scoped to the spell's single damaged target only when it is a creature.
	ExileDieSubjectDamagedCreature bool
	References                     []CompiledReference
	SubjectReferences              []CompiledReference
	Targets                        []CompiledTarget
	SubjectTargets                 []CompiledTarget
	Duration                       DurationKind
	DelayedTiming                  game.DelayedTriggerTiming
	Selector                       CompiledSelector
	// DamageRecipient bundles the primary-recipient descriptors of a deal-damage
	// effect (dual-recipient groups, each-source group, and referenced-player
	// reference) into one typed payload. Its zero value denotes a single-target
	// recipient with no special routing.
	DamageRecipient CompiledDamageRecipient
	// DamageRiders holds the ordered follow-on "... and N damage to <recipient>"
	// damage instructions of a deal-damage clause, in Oracle order: the self
	// rider, the target-controller/owner rider, then the second-target rider.
	// Lowering iterates the list and emits one Damage instruction per rider
	// after the primary damage. It is empty for clauses with no rider.
	DamageRiders   []parser.DamageRiderSyntax
	Amount         CompiledAmount
	PowerDelta     CompiledSignedAmount
	ToughnessDelta CompiledSignedAmount
	TokenPower     int
	TokenToughness int
	TokenPTKnown   bool
	// TokenPTVariableX reports a created token whose printed power and toughness
	// are both the variable "X" ("an X/X ... token"); lowering reads
	// TokenPTDynamic to size it at creation. It is false for fixed tokens.
	TokenPTVariableX bool
	// TokenPTDynamic names the rules-derived amount a variable-X token's power and
	// toughness each equal, bound from the ability's "where X is <dynamic>" clause.
	// It is set only when TokenPTVariableX is true. It is empty for fixed tokens.
	TokenPTDynamic parser.EffectDynamicAmountKind
	// TokenKeywords lists, in source order, every creature keyword a created token
	// enters with ("with flying and vigilance" -> [Flying, Vigilance]). Lowering
	// reads it for token-creation replacements whose substitute token carries more
	// than the single keyword the selector records (Divine Visitation's 4/4 Angel
	// with flying and vigilance). It is nil for tokens with no keyword rider.
	TokenKeywords []parser.KeywordKind
	// TokenToxic carries the integer rank of a created token's toxic keyword
	// ("with toxic 1" -> 1), the one parameterized creature keyword a created
	// token enters with. TokenKeywords records that toxic is present but drops
	// its rank; lowering reads this rank to grant the parameterized toxic keyword
	// ability. It is 0 for tokens with no toxic keyword.
	TokenToxic int
	// TokenGrantedAbility is the quoted ability a created token enters with ("...
	// token with \"When this token dies, you gain 1 life.\""), parsed once through
	// the pipeline. Lowering compiles its inner document and attaches the runtime
	// ability to the token's definition. It is nil for tokens with no such rider.
	TokenGrantedAbility *parser.StaticGrantedAbilitySyntax
	// TokenGrantedAbilityRiderSpan covers the trailing "It has \"...\"." / "They
	// have \"...\"." rider sentence that supplied TokenGrantedAbility, so lowering
	// credits its tokens toward source coverage. It is the zero span when the
	// granted ability came from an inline "token with \"...\"" clause instead.
	TokenGrantedAbilityRiderSpan shared.Span
	// GainGrantedAbility is the quoted ability a resolving ability grant confers
	// on its subject ("This creature gains \"Whenever this creature deals combat
	// damage to a player, that player loses the game.\""), parsed once through the
	// pipeline. Lowering compiles its inner document and applies the runtime
	// ability as a continuous grant. It is nil for gain effects with no such rider.
	GainGrantedAbility *parser.StaticGrantedAbilitySyntax
	// EmblemAbilities are the quoted abilities of an EffectCreateEmblem effect
	// ("You get an emblem with \"Creatures you control have base power and
	// toughness 9/9.\""), each parsed once through the pipeline. Lowering
	// compiles each inner document and emits a game.CreateEmblem carrying the
	// runtime abilities. It is nil for every effect that creates no emblem.
	EmblemAbilities []parser.StaticGrantedAbilitySyntax
	// DelayedTriggerAbility is the nested triggered ability of an
	// EffectDelayedTrigger effect, reparsed from the sentence with its "this
	// turn" window stripped. Lowering compiles its inner document and emits a
	// game.CreateDelayedTrigger carrying the inner trigger pattern and content.
	// It is nil for effects that are not event-based delayed triggers.
	DelayedTriggerAbility *parser.StaticGrantedAbilitySyntax
	// PayRepeatedlyAnimate is the typed payload of an EffectPayRepeatedlyAnimate
	// effect (Primal Adversary's enters trigger): the repeatable mana cost, the
	// +N/+N counter dimensions, the animated lands' base power/toughness, the
	// added creature subtype(s), and the granted keyword(s). It is nil for every
	// other effect.
	PayRepeatedlyAnimate *parser.PayRepeatedlyAnimateSyntax
	// AnimateSelf is the typed payload of an EffectAnimateSelf effect (Faerie
	// Conclave, the Keyrune mana rocks, Mutavault): the source's set base
	// power/toughness, the stated colors, the added artifact card type, the added
	// creature subtype(s) or every-creature-type rider, and the granted
	// keyword(s). It is nil for every other effect.
	AnimateSelf *parser.AnimateSelfSyntax
	// AnimateTarget is the typed payload of an EffectAnimateTarget effect (Animate
	// Land, Vivify, Hydroform): the targeted land's set base power/toughness, the
	// stated colors, the added creature subtype(s), and the granted keyword(s). It
	// reuses the AnimateSelfSyntax shape. It is nil for every other effect.
	AnimateTarget *parser.AnimateSelfSyntax
	// DelayedTriggerOneShot records that an EffectDelayedTrigger fires only on
	// the first matching event ("the next time you cast ..."). It is meaningful
	// only when Kind is EffectDelayedTrigger.
	DelayedTriggerOneShot bool
	// DelayedTriggerBindDamageSource records that an EffectDelayedTrigger's
	// combat-damage event source binds to the permanent an earlier clause in the
	// same resolution acted on ("... target creature ... Whenever that creature
	// deals combat damage to a player this turn, ..."). It is meaningful only
	// when Kind is EffectDelayedTrigger.
	DelayedTriggerBindDamageSource bool
	// DelayedTriggerBindAttacker records that an EffectDelayedTrigger's
	// attacker-declared event binds to the permanent an earlier clause in the
	// same resolution acted on ("... target creature ... Whenever that creature
	// attacks the monarch this turn, ..."). It is meaningful only when Kind is
	// EffectDelayedTrigger.
	DelayedTriggerBindAttacker bool
	// DelayedTriggerBindDyingObject records that an EffectDelayedTrigger's
	// permanent-died event binds to the permanent an earlier clause in the same
	// resolution acted on ("... target creature an opponent controls ... When the
	// creature an opponent controls dies this turn, ..."). It is meaningful only
	// when Kind is EffectDelayedTrigger.
	DelayedTriggerBindDyingObject bool
	// TokenName is a created creature token's explicit Oracle name ("named Koma's
	// Coil"), captured verbatim from source. It is empty when the token is named
	// only by its subtypes.
	TokenName string
	// TokenPredefinedName is a created predefined named token's name when that
	// name is a card name rather than a card subtype ("create a tapped Mutavault
	// token." -> "Mutavault"). Lowering maps the name to the token's fixed
	// definition. It is empty for tokens identified by their subtypes.
	TokenPredefinedName string
	TokenCopyOfTarget   bool
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
	// TokenCopyOfReferenceHalvedPT mirrors the parser flag for the linked
	// halved-copy create "its controller creates <N> tokens that are copies of
	// that creature, except their power/toughness is half that creature's ...
	// Round up each time." (Saw in Half). The dies-this-way copy sequence lowering
	// consumes it, copying the preceding destroy's target and halving each copy's
	// power and toughness (rounded up).
	TokenCopyOfReferenceHalvedPT bool
	// TokenCopyHalvePTRoundUp mirrors the parser flag recording that the halved
	// copy's power and toughness round up (the credited "Round up each time."
	// rider). Without it the halved-copy sequence fails closed.
	TokenCopyHalvePTRoundUp bool
	// TokenCopyHalveRoundUpRiderSpan mirrors the parser span covering the
	// "Round up each time." rider sentence, so lowering credits the rider tokens
	// as consumed. It is set together with TokenCopyHalvePTRoundUp.
	TokenCopyHalveRoundUpRiderSpan shared.Span
	// TokenCopyOfAttached reports that the created token is a copy of the
	// permanent the source is attached to ("a copy of equipped creature" /
	// "enchanted creature"). The copy source resolves at runtime to the attached
	// permanent.
	TokenCopyOfAttached bool
	// TokenCopyOfTriggeringSet reports that the created token is a copy of one of
	// the permanents that triggered this ability, chosen by the controller
	// ("create a token that's a copy of one of them." on a "Whenever one or more
	// ... enter" trigger). The copy source resolves at runtime to a
	// controller-chosen member of the triggering event batch.
	TokenCopyOfTriggeringSet bool
	// TokenCopyDropLegendary reports a copy-token "except <it/the token> isn't
	// legendary" modifier: the created token drops the Legendary supertype.
	TokenCopyDropLegendary bool
	// TokenCopyEntersTapped reports a copy-token "tapped" entry modifier: every
	// created copy enters the battlefield tapped.
	TokenCopyEntersTapped bool
	// TokenCopyGrantKeywords lists keyword abilities the created copy token gains
	// from a folded "[That token/It] gains <keyword>." rider, in source order.
	TokenCopyGrantKeywords []parser.KeywordKind
	// TokenCopyGrantRiderSpan covers the folded gain-keyword rider sentence so
	// lowering credits its tokens toward source coverage.
	TokenCopyGrantRiderSpan shared.Span
	// TokenCopyOverride and the TokenCopyOverride* fields carry a copy-token
	// characteristic-overriding "except" exception ("except it's a 1/1 green
	// Frog", "except it's an artifact in addition to its other types"). The
	// created token copies its source and then applies these power/toughness,
	// color, card-type, subtype, and keyword overrides. Colors and subtypes are
	// additive when TokenCopyOverrideAdditiveColors/AdditiveTypes is set and
	// replace the copied values otherwise; card types are always additive.
	TokenCopyOverride               bool
	TokenCopyOverridePTKnown        bool
	TokenCopyOverridePower          int
	TokenCopyOverrideToughness      int
	TokenCopyOverrideColors         []color.Color
	TokenCopyOverrideSubtypes       []types.Sub
	TokenCopyOverrideTypes          []types.Card
	TokenCopyOverrideKeywords       []parser.KeywordKind
	TokenCopyOverrideAdditiveTypes  bool
	TokenCopyOverrideAdditiveColors bool
	// TokenChoice reports a create-token effect offering a choice among two or
	// more complete named-token specs ("create a Food token or a Treasure token",
	// "create your choice of a Clue token, a Food token, or a Treasure token").
	// The alternatives are the selector's SubtypesAny entries in source order;
	// lowering emits a choose-one modal ability creating exactly one of them.
	TokenChoice bool
	// AdditionalTokens carries the second and later token specs of a multi-token
	// create effect ("Create a 1/1 green Snake creature token, a 2/2 green Wolf
	// creature token, and a 3/3 green Elephant creature token."). The effect's own
	// token fields describe the first token; each entry here is a compiled
	// creature-token spec for one of the remaining tokens, in source order.
	// Lowering emits one CreateToken instruction per token. It is empty for every
	// single-token create.
	AdditionalTokens  []CompiledEffect
	StaticSubject     StaticSubjectKind
	StaticSubjectSpan shared.Span
	Details           *CompiledEffectDetails
	CounterKind       counter.Kind
	CounterKindKnown  bool
	// ChooseExiledCardOwnerOpponent scopes an EffectChooseExiledCard choice to
	// cards owned by an opponent of the resolving controller ("Choose an exiled
	// card an opponent owns ...", Dauthi Voidwalker). Lowering maps it to the
	// PlayChosenExiledCard primitive's PlayerOpponent owner scope; it is false for
	// every other effect, and any unrecognized owner wording leaves it false so
	// the choice fails closed rather than silently widening the scope.
	ChooseExiledCardOwnerOpponent bool
	// CounterKindChoices lists the counter kinds a placement effect lets the
	// resolving controller choose between ("a +1/+1 counter or a loyalty counter
	// on it.", Elspeth Conquers Death chapter III). It holds two or more distinct
	// kinds and is set only when CounterKindKnown is false. Lowering emits an
	// AddCounter that prompts the controller for one of these kinds.
	CounterKindChoices []counter.Kind
	// CounterRecipientAttached reports that a counter-placement effect places its
	// counters on the permanent the source Aura is attached to ("... on enchanted
	// creature"). Lowering routes it to the runtime's source attached-permanent
	// reference; it is false for every other recipient.
	CounterRecipientAttached bool
	// FightSubjectAttached reports that a fight effect's fighter is the permanent
	// the source Aura or Equipment is attached to ("enchanted creature fights up
	// to one target creature"). Lowering routes the fighting object to the
	// runtime's source attached-permanent reference; it is false for every other
	// fight subject.
	FightSubjectAttached bool
	// CounterRecipientSingleChoice reports that a non-target counter-placement
	// effect places its counters on a single permanent the controller chooses
	// from a battlefield group ("put a vigilance counter on a creature you
	// control"), rather than on every member of an "each <group>" recipient.
	// Lowering emits a single-choice placement; it is false for the distributive
	// group form, which compiles to an identical selector.
	CounterRecipientSingleChoice bool
	// enchanted creature." / "Regenerate equipped creature."). Lowering routes it
	// to the runtime's source attached-permanent reference; it is false for every
	// other regenerate recipient.
	RegenerateAttached bool
	// ExileAttached reports the attached-recipient exile form ("Exile enchanted
	// creature." / "Exile equipped creature."). Lowering routes the exiled object
	// to the runtime's source attached-permanent reference; it is false for every
	// other exile recipient.
	ExileAttached bool
	// TapAttached reports the attached-recipient tap form ("Tap enchanted
	// creature." / "Tap enchanted permanent." / "Tap equipped creature.").
	// Lowering routes the tapped object to the runtime's source attached-permanent
	// reference; it is false for every other tap recipient.
	TapAttached bool
	// UntapAttached reports the attached-recipient untap form ("Untap enchanted
	// creature." / "Untap enchanted permanent." / "Untap equipped creature.").
	// Lowering routes the untapped object to the runtime's source attached-
	// permanent reference; it is false for every other untap recipient.
	UntapAttached bool
	// MoveCountersAll carries the parser's kind-agnostic "move all counters"
	// form of an EffectMoveCounters effect through to lowering, which moves every
	// counter on the source regardless of kind. It is false for a specific-kind
	// move, whose kind is in CounterKind / CounterKindKnown.
	MoveCountersAll bool
	// MoveCountersAllOfKind carries the parser's kind-specific mass form "move all
	// <kind> counters from <source> onto <target>" through to lowering, which
	// moves every counter of the single named kind (CounterKind / CounterKindKnown)
	// while leaving other kinds behind (CR 702.44 Modular). It is false for the
	// kind-agnostic "all counters" form (MoveCountersAll) and the fixed-count move.
	MoveCountersAllOfKind bool
	// RemoveCountersAll carries the parser's kind-agnostic "remove all counters"
	// form of an EffectRemoveCounter effect through to lowering, which removes
	// every counter on the object regardless of kind. It is false for a fixed or
	// kind-specific removal, whose count is in Amount and kind in CounterKind.
	RemoveCountersAll bool
	// MoveCountersDistribute carries the parser's "move any number of <kind>
	// counters from <source> onto other creatures" form through to lowering,
	// which distributes the source's counters among a group of other creatures
	// rather than moving them onto a single target. It is false for the
	// single-target move forms.
	MoveCountersDistribute bool
	// MoveThoseCounters carries the parser's counter-salvage form "put those
	// counters on <destination>" through to lowering, which reads the counters
	// the triggering event permanent had (its last-known information) and places
	// them on the destination. It is set only on EffectPut effects.
	MoveThoseCounters bool
	// MoveCountersFromTarget carries the parser's two-target counter-move form
	// (counters read from a first chosen target permanent and placed onto a
	// second chosen target permanent) through to lowering, which emits a
	// MoveCounters reading the source target. It is false for the self-source
	// single-target move and the distributed group form.
	MoveCountersFromTarget bool
	// MoveCountersAnyKind carries the parser's kind-unspecified single counter
	// move ("Move a counter ..."), where the controller moves one counter of any
	// kind present on the source. It is false for a named-kind move and the
	// kind-agnostic "all counters" move.
	MoveCountersAnyKind bool
	FromZone            zone.Type
	// GraveyardZoneExile carries the parser's recognized whole-graveyard exile
	// owner relation ("Exile target player's graveyard.") through to lowering,
	// which builds the target-player + graveyard-group MoveCard. It is
	// GraveyardZoneExileNone for every other effect.
	GraveyardZoneExile parser.GraveyardZoneExileKind
	ToZone             zone.Type
	Destination        parser.EffectDestinationPosition
	EntersTapped       bool
	EntersTappedSelf   bool
	// EntersAttacking mirrors the parser's "attacking" put rider ("put ... onto
	// the battlefield tapped and attacking"): the entered permanent is put onto
	// the battlefield attacking (CR 508.4).
	EntersAttacking bool
	// EntersTransformed mirrors the parser's Transformers "converted" return
	// rider ("return it to the battlefield converted", CR 712): the returned
	// transforming double-faced card enters as its back face. It is false for the
	// plain untransformed return.
	EntersTransformed bool
	// GroupEntryModification mirrors the parser's typed static group
	// entry-modification payload: the enters-tapped-group form (Authority of the
	// Consuls) and the enters-with-counters-group form (Tayam, Luminous Enigma).
	// Lowering reads it through the EntersTappedGroup and EntersWithCountersGroup
	// accessors; the tapped form's ControllerScope/Types build the continuous
	// controller- and type-scoped replacement, while the with-counters form reads
	// the counter from Selector/CounterKind. It is the zero value for the self
	// entry forms.
	GroupEntryModification   CompiledGroupEntryModification
	EntersColorChoice        bool
	EntersColorChoiceExclude mana.Color
	EntersTypeChoice         bool
	EntersWithCounters       bool
	// EntersWithCountersKeywordRider mirrors the parser-owned marker for the
	// unsupported combined "enters with counters ... and with <keyword>" form.
	EntersWithCountersKeywordRider bool
	// EntersDevour mirrors the parser's Devour as-enters replacement flag and
	// EntersDevourMultiplier its per-sacrificed-permanent +1/+1 counter count.
	// EntersDevourType and EntersDevourSubtype carry the typed-variant sacrifice
	// filter (artifact, land, Food); both are zero for the creature form.
	// Lowering reads them to build the runtime Devour replacement (CR 702.81).
	EntersDevour           bool
	EntersDevourMultiplier int
	EntersDevourType       types.Card
	EntersDevourSubtype    types.Sub
	// EntersTribute mirrors the parser's Tribute as-enters replacement flag and
	// EntersTributeCount its +1/+1 counter count N.
	EntersTribute      bool
	EntersTributeCount int
	// EntersAsCopy mirrors the parser's enters-as-copy replacement flag and its
	// riders. Lowering reads the effect's Selector for the copied-permanent
	// filter and these flags for the "you may" form and the copiable riders.
	EntersAsCopy             bool
	EntersAsCopyOptional     bool
	EntersAsCopyNotLegendary bool
	EntersAsCopyAddTypes     []types.Card
	// EntersAsCopyAddSubtypes mirrors the parser's "except it's a <subtype> in
	// addition to its other types" copiable subtype riders (Mockingbird's Bird,
	// Synth Infiltrator's Synth).
	EntersAsCopyAddSubtypes []types.Sub
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
	// BecomeTypeAddTypes and BecomeTypeUntilEndOfTurn mirror the parser's
	// EffectBecomeType targeted type-change ("Target permanent becomes an
	// artifact in addition to its other types until end of turn."). Lowering
	// reads them to build the ApplyContinuous LayerType type addition and its
	// duration. BecomeTypeAddColors carries the colors added by the additive
	// color-and-type form ("becomes a blue artifact in addition to its other
	// colors and types"); it is empty for the color-free form and drives a
	// LayerColor color addition during lowering.
	BecomeTypeAddTypes       []types.Card
	BecomeTypeAddColors      []color.Color
	BecomeTypeAddSubtypes    []types.Sub
	BecomeTypeUntilEndOfTurn bool
	// BecomeColorColors, BecomeColorColorless, BecomeColorSource, and
	// BecomeColorUntilEndOfTurn mirror the parser's EffectBecomeColor color-set
	// payload ("<subject> becomes <color>... until end of turn."). Lowering reads
	// them to build the ApplyContinuous LayerColor SET (or colorless clear) over
	// the source or single target for the duration.
	BecomeColorColors         []color.Color
	BecomeColorColorless      bool
	BecomeColorSource         bool
	BecomeColorUntilEndOfTurn bool
	// Polymorph* mirror the parser's EffectPolymorph payload ("Until end of turn,
	// target creature loses all abilities and becomes a <color> <subtype> with
	// base power and toughness N/N."). Lowering reads them to build the
	// ApplyContinuous that removes all abilities and SETS the creature's color,
	// type, and base power/toughness until end of turn.
	PolymorphColors        []color.Color
	PolymorphColorless     bool
	PolymorphSubtypes      []types.Sub
	PolymorphBasePower     int
	PolymorphBaseToughness int
	// PolymorphName, PolymorphSupertypes, and PolymorphPermanent mirror the
	// parser's permanent named-become polymorph payload ("Target nontoken
	// creature becomes a 6/6 legendary Horror creature named Fenric and loses all
	// abilities."). Lowering reads them to add a LayerText name change, the added
	// supertypes, and a permanent duration on top of the shared polymorph
	// continuous effects. They are zero for the until-end-of-turn polymorph forms.
	PolymorphName       string
	PolymorphSupertypes []types.Super
	PolymorphPermanent  bool
	// SetBasePower, SetBaseToughness, SetBasePTVariableX,
	// SetBasePTEveryCreatureType, and SetBasePTSource mirror the parser's
	// EffectSetBasePT payload ("{X}: Until end of turn, creatures you control have
	// base power and toughness X/X and gain all creature types.", Mirror Entity).
	// Lowering reads them to build the ApplyContinuous that SETS base
	// power/toughness at LayerPowerToughnessSet (a fixed value or the cost's X)
	// and, when SetBasePTEveryCreatureType is set, adds every creature type at
	// LayerType. The affected group is carried in the effect's Selector via
	// StaticSubject; SetBasePTSource marks the source-affecting form.
	SetBasePower               int
	SetBaseToughness           int
	SetBasePTVariableX         bool
	SetBasePTEveryCreatureType bool
	SetBasePTSource            bool
	// SetBasePTLosesAllAbilities mirrors the parser flag: the affected object also
	// loses all abilities for the duration ("<subject> loses all abilities and has
	// base power and toughness N/N").
	SetBasePTLosesAllAbilities bool
	// LoseAllAbilities mirrors the parser flag: a resolving "<subject> loses all
	// abilities" effect that removes every ability for the duration, lowering to a
	// LayerAbility RemoveAllAbilities continuous effect.
	LoseAllAbilities bool
	// SwitchPTSource mirrors the parser's EffectSwitchPT source-affecting form
	// ("Switch this creature's power and toughness until end of turn."). When
	// false and the effect carries a target, the single-target switch applies.
	SwitchPTSource bool
	// EntersAsCopyUntilEndOfTurn mirrors the parser's temporary "become a copy
	// ... until end of turn" copy duration (Cursed Mirror).
	EntersAsCopyUntilEndOfTurn bool
	// EntersAsCopyAddKeywords mirrors the parser's "except it has <keyword>"
	// copiable keyword riders (Cursed Mirror's haste).
	EntersAsCopyAddKeywords []parser.KeywordKind
	// EntersAsCopyTapped mirrors the parser's "enter tapped as a copy" form
	// (Vesuva), where the permanent enters tapped as its chosen copy.
	EntersAsCopyTapped bool
	// EntersAsCopyBasePower and EntersAsCopyBaseToughness mirror the parser's
	// "except it's N/N" copiable P/T-override rider (Quicksilver Gargantuan).
	EntersAsCopyBasePower     opt.V[int]
	EntersAsCopyBaseToughness opt.V[int]
	// EntersAsCopyMaxManaValueFromManaSpent mirrors the parser's "with mana value
	// less than or equal to the amount of mana spent to cast this creature"
	// copiable filter (Mockingbird).
	EntersAsCopyMaxManaValueFromManaSpent bool
	UnderYourControl                      bool
	CastAsAdventure                       bool
	// CastWithoutPayingManaCost mirrors the parser's free-cast rider flag for a
	// cast effect ("... without paying its mana cost"). Lowering reads it to
	// route the cast-for-free primitive; it is false for every other effect.
	CastWithoutPayingManaCost bool
	// PlayHideawayExiledCard mirrors the parser flag for the Hideaway activated
	// ability effect "(you may) play the exiled card without paying its mana
	// cost" (CR 702.75c). Lowering reads it to emit the play-hideaway-card
	// primitive gated by the ability's activation condition; it is false for
	// every other effect.
	PlayHideawayExiledCard bool
	// ImpulseCast mirrors the parser flag marking an impulse-exile play
	// permission that grants casting the exiled card ("you may cast that card")
	// rather than playing it. Lowering reads it to grant cast-only permission; it
	// is false for every other effect.
	ImpulseCast bool
	// ImpulseSpendAnyColor mirrors the parser flag marking an impulse-exile play
	// permission carrying the any-color rider "and you may spend mana as though
	// it were mana of any color to cast that spell." (Grenzo, Havoc Raiser).
	// Lowering reads it to set the exiled cast's spend-any-mana permission; it is
	// false for every other effect.
	ImpulseSpendAnyColor bool
	Negated              bool
	// FallbackOnInability mirrors the parser flag for a "who can't" relative
	// clause effect ("Each player who can't discards a card."): it applies only
	// to players who couldn't satisfy the immediately preceding required action.
	FallbackOnInability bool
	Optional            bool
	Divided             bool
	// DistributeCounters mirrors the parser flag for a "Distribute N <kind>
	// counters among <cardinality> target creatures" effect: a fixed (or X)
	// total of counters split among the chosen targets, at least one each.
	DistributeCounters bool
	OptionalSpan       shared.Span
	Mana               CompiledEffectMana
	Replacement        parser.EffectReplacementSyntax
	Payment            CompiledEffectPayment
	Exact              bool
	// TapUntapReferenceObjectClean mirrors the parser flag: the tap or untap
	// effect's clause is exactly "<verb> <object>." for the source or a singular
	// back-reference, tolerating "you may" optionality and sibling-clause
	// references. Lowering reads it to admit the self/back-reference tap-down
	// family while a tap/untap whose clause would drop a trailing unrecognized
	// conjunct stays unsupported.
	TapUntapReferenceObjectClean bool
	// disjunctive runtime choice ("gains banding, first strike, or trample")
	// rather than a conjunctive grant of every listed keyword. Lowering keys on
	// it to emit a choose-one keyword grant instead of granting all keywords.
	KeywordGrantChoice bool
	// KeywordGrantChoiceAtRandom marks a KeywordGrantChoice grant whose one
	// chosen keyword is selected at random rather than by the controller (the
	// two-sentence "choose <keyword> or <keyword> at random. <source> gains that
	// ability until end of turn." construction). Lowering keys on it to emit an
	// at-random modal keyword grant instead of a controller-chosen one.
	KeywordGrantChoiceAtRandom bool
	// KeywordChoiceAtRandomPreludeSpan covers the folded "choose <keyword> or
	// <keyword> at random." prelude sentence so lowering widens the trigger body
	// span to cover the listed keywords. It is the zero span for every effect
	// that is not an at-random keyword-choice grant.
	KeywordChoiceAtRandomPreludeSpan shared.Span
	// RevealUntilThenPut carries the parser's typed marker for the closed
	// "reveal from the top of a library until a <type> card, then put those
	// cards into <zone>" sequence. It is set on each of the three effects of
	// the recognized shape; lowering keys on it (with the match-Reveal's
	// Selector and the Put's ToZone) to emit a single RevealUntil primitive.
	RevealUntilThenPut bool
	// RevealTopPartition carries the parser's typed marker for the closed
	// "Reveal the top N cards of your library. Put all <type> cards revealed this
	// way into your hand and the rest <remainder>." sequence. It is set on the
	// Reveal and Put effects of the recognized shape; lowering keys on it (with
	// the Reveal's Amount, the Put's Selector, and RevealPartitionRemainder) to
	// emit a single RevealTopPartition primitive.
	RevealTopPartition bool
	// RevealPartitionRemainder records where the un-taken revealed cards go in a
	// RevealTopPartition sequence. It is set only on the Put effect; the zero
	// value is the graveyard remainder.
	RevealPartitionRemainder parser.DigRemainderKind
	// PileSplitSequence, with the role/destination/amount/middle-span fields
	// below, carries the parser's typed marker for the closed pile-split
	// sequence (Fact or Fiction, Steam Augury, Sphinx of Uthuun). It is set on
	// the Reveal and Put effects of the recognized shape; lowering keys on it to
	// emit a single PileSplit primitive.
	PileSplitSequence bool
	// PileSplitSeparatorOpponent and PileSplitChooserOpponent report whether an
	// opponent (rather than the controller) separates the revealed cards into two
	// piles, and whether an opponent chooses which pile the controller keeps. They
	// are set only on the Put effect of a recognized pile-split sequence.
	PileSplitSeparatorOpponent bool
	PileSplitChooserOpponent   bool
	// PileSplitOtherZone is the destination of the pile the controller does not
	// keep (the kept pile always goes to hand); PileSplitAmount is the number of
	// cards revealed. They are set only on the Put effect of a pile-split sequence.
	PileSplitOtherZone zone.Type
	PileSplitAmount    int
	// PileSplitMiddleSpan covers the zero-effect middle sentence so lowering can
	// credit its tokens toward source coverage. It is set only on the Put effect
	// of a recognized pile-split sequence.
	PileSplitMiddleSpan shared.Span
	// ExiledCardSplitOpponentChooses reports that a recognized "An opponent
	// chooses one of the exiled cards." antecedent names an opponent as the
	// chooser of the cost-exiled card this put effect disposes of (Coin of Fate).
	// It is set only on the put effect of a recognized exiled-card opponent-choice
	// disposal; lowering reads it to synthesize the opponent's choice.
	ExiledCardSplitOpponentChooses bool
	// ExiledCardChoiceRiderSpan covers the zero-effect antecedent "An opponent
	// chooses one of the exiled cards." so lowering can credit its tokens toward
	// source coverage. It is set only when ExiledCardSplitOpponentChooses is true.
	ExiledCardChoiceRiderSpan shared.Span
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
	// SourceSpellCostReductionConditional carries the typed source-scoped flat
	// cast cost reduction gated by the ability's condition clause ("This spell
	// costs {N} less to cast if <condition>"). SourceSpellCostReductionAmount is
	// the flat generic reduction N; lowering gates it on the ability's single
	// typed condition.
	SourceSpellCostReductionConditional           bool
	SourceSpellCostReductionTargetsTappedCreature bool
	RequiresOrderedLowering                       bool
	HasUnrecognizedSibling                        bool
	UnsupportedDetail                             string
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
	// PersistUntilEndOfTurnRiderSpan covers the cross-sentence "Until end of
	// turn, you don't lose this mana as steps and phases end" rider folded onto
	// an add-mana effect (Grand Warlord Radha) so lowering can credit its tokens
	// toward source coverage. It is set only when Mana.PersistUntilEndOfTurn is.
	PersistUntilEndOfTurnRiderSpan shared.Span
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
	// RandomDiscard carries the "discard N card(s) at random" flag for a
	// non-controller subject through the text-blind compiler boundary so lowering
	// selects the random discard primitive variant without inspecting Oracle
	// words.
	RandomDiscard bool
	// RevealChooseDiscard marks the reveal and discard halves of a recognized
	// "Target player reveals their hand. You choose a [filter] card from it. That
	// player discards that card." sequence so the text-blind lowering can pair
	// them into a single ChooseDiscardFromHand primitive.
	RevealChooseDiscard bool
	// HandChoiceDiscard carries the filter and coverage span of that sequence's
	// middle "You choose..." sentence. It is set only on the discard half
	// (HandChoiceDiscard.Present true).
	HandChoiceDiscard parser.HandChoiceDiscardSyntax
	// DiscardThenDraw marks a discard clause fused with a following "then draw
	// that many cards[ plus K]" clause into a single variable looter. The
	// controller discards a chosen number of cards (at most DiscardThenDrawMax,
	// or any number when zero), then draws that many plus DiscardThenDrawOffset.
	DiscardThenDraw       bool
	DiscardThenDrawMax    int
	DiscardThenDrawOffset int
	// SacrificeThenCount marks a sacrifice clause whose count feeds an
	// immediately following "then <create|draw|add> that many/much" reward, so
	// lowering publishes the number sacrificed and scales the reward by it.
	// SacrificeAnyNumber records that the sacrifice form is "any number of"
	// (player-chosen count) rather than "all". They are false for every other
	// effect.
	SacrificeThenCount bool
	SacrificeAnyNumber bool
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
	// SearchLandElseHand mirrors the parser-owned exact Archdruid's Charm search
	// mode.
	SearchLandElseHand bool
	// SearchDifferentNames carries the "with different names" correlation rider
	// from the parser so the search lowerer can set SearchSpec.DifferentNames
	// without re-reading the search text.
	SearchDifferentNames bool
	// SearchDestination carries the parser-recognized ordered destination for a
	// found card that remains in the library.
	SearchDestination parser.EffectDestinationPosition
	// SearchControl carries the parser-recognized "under target player's control"
	// controller rider on a search-and-put-onto-the-battlefield clause so the
	// search lowerer can route the found permanent under the named target
	// player's control without re-reading the search text.
	SearchControl parser.SearchControlRider
	// SearchSlots carries the per-slot subtypes of a heterogeneous multi-slot
	// library search ("a Forest card and a Plains card", Krosan Verge) from the
	// parser so the search lowerer can build a SearchSpec.SlotFilters from typed
	// subtypes rather than re-reading the search text. It is empty for every
	// ordinary single-filter search.
	SearchSlots []types.Sub
	// DiscardEntireHand carries the parser-recognized "discard their hand" clause
	// through the text-blind compiler boundary: the affected player discards
	// every card in hand rather than a fixed count.
	DiscardEntireHand bool
	// CounteredSpellExileReplacement carries the parser-recognized "If that
	// spell is countered this way, exile it instead of putting it into its
	// owner's graveyard." rider through the text-blind compiler boundary.
	CounteredSpellExileReplacement bool
	// CounteredSpellDestinationReplacement carries the parser-recognized "If that
	// spell is countered this way, put it [on top of its owner's library | into
	// its owner's hand] instead of into that player's graveyard." rider through
	// the text-blind compiler boundary. The redirect zone is read from ToZone and
	// Destination by the counter lowerer.
	CounteredSpellDestinationReplacement bool
	// CounterTriggeringStackObject mirrors the parser-owned exact
	// "Counter that spell or ability." reference form.
	CounterTriggeringStackObject bool
	// ShuffleEachPlayerGraveyardIntoLibrary mirrors the parser-owned exact
	// symmetric graveyard shuffle.
	ShuffleEachPlayerGraveyardIntoLibrary bool
	// ExileUntilSourceLeaves carries the parser-recognized O-Ring exile clause
	// "exile <target> until <this permanent> leaves the battlefield." through the
	// text-blind compiler boundary so lowering links the exile to the source.
	ExileUntilSourceLeaves bool
	// ExileUntilOpponentBecomesMonarch carries the parser-recognized monarch exile
	// clause "exile <target> until an opponent becomes the monarch." (Palace
	// Jailer) through the text-blind compiler boundary so lowering links the exile
	// to the source and synthesizes the become-monarch return trigger.
	ExileUntilOpponentBecomesMonarch bool
	// ReturnExiledCard carries the parser-recognized explicit O-Ring return
	// clause "return the exiled card to the battlefield under its owner's
	// control." through the text-blind compiler boundary so lowering emits the
	// linked battlefield return paired with the sibling exile.
	ReturnExiledCard bool
	// ExileEntireHand carries the parser-recognized whole-hand exile clause
	// "Exile all cards from your hand." (Wormfang Behemoth) through the
	// text-blind compiler boundary so lowering emits the linked entire-hand
	// exile paired with the sibling return-to-hand trigger.
	ExileEntireHand bool
	// ReturnExiledCardsToHand carries the parser-recognized return clause
	// "Return the exiled cards to their owner's hand." (Wormfang Behemoth)
	// through the text-blind compiler boundary so lowering emits the linked
	// return to hand of the set the sibling entire-hand exile removed.
	ReturnExiledCardsToHand bool
	// BottomLinkedExiledCards carries the parser-recognized linked disposal
	// clause "The owner of each card exiled with <this permanent> puts that card
	// on the bottom of their library." (Trial of a Time Lord) through the
	// text-blind compiler boundary so lowering emits the linked library-bottom
	// disposal paired with the sibling exile.
	BottomLinkedExiledCards bool
	// ExileForEachPlayerUntilSourceLeaves carries the parser-recognized
	// distributive Saga exile clause "For each player, exile up to one [other]
	// target <permanent> that player controls until <this Saga> leaves the
	// battlefield." (Vault 13: Dweller's Journey) through the text-blind compiler
	// boundary so lowering links each player's chosen permanent to the source.
	ExileForEachPlayerUntilSourceLeaves bool
	// ReturnLinkedExiledToBattlefieldPartial carries the parser-recognized
	// partial payoff clause "Return <count> cards exiled with <this Saga> to the
	// battlefield under their owners' control." (Vault 13: Dweller's Journey)
	// through the text-blind compiler boundary so lowering returns a fixed-size
	// subset of the linked exiled set.
	ReturnLinkedExiledToBattlefieldPartial bool
	// PutLinkedExiledRestOnLibraryBottom carries the parser-recognized remainder
	// disposal clause "put the rest on the bottom of their owners' libraries."
	// (Vault 13: Dweller's Journey) through the text-blind compiler boundary so
	// lowering routes the unreturned remainder of the linked exiled set to the
	// bottom of their owners' libraries.
	PutLinkedExiledRestOnLibraryBottom bool
	// DestroyForEachPlayer carries the parser-recognized distributive Saga destroy
	// clause "For each player, destroy up to one target creature that player
	// controls." (The Curse of Fenric, chapter I) through the text-blind compiler
	// boundary so lowering destroys up to one creature each player controls and
	// links each destroyed creature for the paired token payoff.
	DestroyForEachPlayer bool
	// EachPlayerChooseDestroy carries the parser-recognized "Starting with you,
	// each player may choose <permanent>. Destroy each permanent chosen this
	// way." construct (Druid of Purification) through the text-blind compiler
	// boundary. When set, Selector is the shared candidate pool and lowering
	// emits a single EachPlayerChooseDestroy over it.
	EachPlayerChooseDestroy bool
	// EachPlayerChooseDestroyOptional records the "may" of an
	// EachPlayerChooseDestroy construct, so each chooser may decline.
	EachPlayerChooseDestroyOptional bool
	// CreateTokenForEachDestroyedThisWay carries the parser-recognized per-
	// controller payoff "For each creature destroyed this way, its controller
	// creates a <token>." (The Curse of Fenric, chapter I) through the text-blind
	// compiler boundary so lowering creates one token for each creature a sibling
	// DestroyForEachPlayer destroyed, controlled by that creature's controller.
	CreateTokenForEachDestroyedThisWay bool
	// CreateTokenForEachExiledThisWay carries the parser-recognized per-controller
	// payoff "For each creature exiled this way, its controller creates a
	// <token>." (Curse of the Swine) through the text-blind compiler boundary so
	// lowering creates one token for each creature a sibling variable-target exile
	// removed, controlled by that creature's controller.
	CreateTokenForEachExiledThisWay bool
	// ExileForEachOpponent carries the parser-recognized distributive enters
	// exile clause "for each opponent, exile up to one target permanent that
	// player controls with mana value 3 or greater." (King Solomon's Frogs)
	// through the text-blind compiler boundary so lowering exiles up to one
	// permanent each opponent controls and links each exiled permanent for the
	// paired draw payoff.
	ExileForEachOpponent bool
	// DrawForEachExiledThisWay carries the parser-recognized per-controller
	// payoff "For each permanent exiled this way, its controller draws a card."
	// (King Solomon's Frogs) through the text-blind compiler boundary so lowering
	// draws one card for each permanent a sibling ExileForEachOpponent exiled,
	// for that permanent's last-known controller.
	DrawForEachExiledThisWay bool
	// to the mana value of the exiled card." (The Aesir Escape Valhalla) through
	// the text-blind compiler boundary so lowering scales the placement by the
	// linked exiled card's mana value.
	CounterExiledCardManaValue bool
	// ReturnSourceAndExiledCardToHand carries the parser-recognized chapter III
	// clause "Return this Saga and the exiled card to their owner's hand." (The
	// Aesir Escape Valhalla) through the text-blind compiler boundary so lowering
	// emits a source bounce paired with a linked return to hand.
	ReturnSourceAndExiledCardToHand bool
	// clause that affects every player ("Players can't cast spells this turn.")
	// rather than only the controller's opponents. Lowering reads it to pick the
	// affected-player relation; it is false for the opponents-only form.
	CantCastSpellsAllPlayers bool
	// CantCastSpellsRequiredTypes and CantCastSpellsExcludedTypes mirror the
	// parser's optional card-type filter on an EffectCantCastSpells clause:
	// RequiredTypes restricts the prohibition to spells of those types ("creature
	// spells"), ExcludedTypes exempts spells of those types ("noncreature
	// spells"). Lowering copies them onto the rule effect's SpellTypes and
	// ExcludedSpellTypes filters; both are empty for the unfiltered form.
	CantCastSpellsRequiredTypes []types.Card
	CantCastSpellsExcludedTypes []types.Card
	// SpellCostModifierCaster, SpellCostModifierAmount, SpellCostModifierIncrease,
	// SpellCostModifierRequiredTypes, and SpellCostModifierExcludedTypes mirror
	// the parser fields for an EffectSpellCostModifier clause: which player's
	// spells are affected, the generic amount they cost more or less, whether the
	// modifier is an increase, and the optional single-card-type required/excluded
	// filter. Lowering reads them to build the duration-bounded
	// RuleEffectCostModifier rule effect.
	SpellCostModifierCaster        parser.SpellCostCasterKind
	SpellCostModifierAmount        int
	SpellCostModifierIncrease      bool
	SpellCostModifierRequiredTypes []types.Card
	SpellCostModifierExcludedTypes []types.Card
	// AttackTaxGeneric mirrors the parser field for an EffectAttackTax clause: the
	// per-attacker generic mana the "... pays {N} for each of those creatures."
	// resolving attack tax charges. Lowering reads it to build the
	// duration-bounded RuleEffectAttackTax rule effect.
	AttackTaxGeneric int
	// PreventDamageTo and PreventDamageBy mirror the parser flags for an
	// EffectPreventDamage clause, recording whether all combat damage dealt to
	// and/or dealt by the referenced permanent is prevented for the turn.
	PreventDamageTo bool
	PreventDamageBy bool
	// PreventDamageAllTypes mirrors the parser flag: a PreventDamageTo/By clause
	// that prevents all damage of any kind rather than only combat damage.
	PreventDamageAllTypes bool
	// PreventDamageGlobal mirrors the parser flag for an EffectPreventDamage
	// clause that prevents every combat damage event this turn with no recipient
	// or source object.
	PreventDamageGlobal bool
	// PreventDamageToController mirrors the parser flag for an EffectPreventDamage
	// clause that prevents every combat damage event dealt to the controller this
	// turn ("Prevent all combat damage that would be dealt to you this turn." —
	// Inkshield). It is the controller-recipient sibling of PreventDamageGlobal.
	PreventDamageToController bool
	// PreventDamageNextRecipient mirrors the parser kind for the amount-based
	// "Prevent the next N damage that would be dealt to <recipient> this turn."
	// shield, naming the shielded recipient. The prevented amount N rides on the
	// effect's Amount. It is None for the combat prevention forms.
	PreventDamageNextRecipient parser.PreventDamageRecipientKind
	// PreventDamageThatAmount mirrors the parser amount for the continuous static
	// "prevent N of that damage" replacement (Sphere of Law). It is zero for
	// every one-shot or combat prevention form.
	PreventDamageThatAmount int
	// PreventDamageNextFromSource mirrors the parser flag for the one-shot "The
	// next time a [color] source of your choice would deal damage to you this
	// turn, prevent that damage." shield (Circle of Protection, Rune of
	// Protection). PreventDamageSourceColors carries its optional source color
	// filter; an empty slice matches a source of any color.
	PreventDamageNextFromSource bool
	PreventDamageSourceColors   []color.Color
	// PreventDamageRedirectToSourceController mirrors the parser flag for the
	// Deflecting Palm redirect rider ("If damage is prevented this way, Deflecting
	// Palm deals that much damage to that source's controller."), folded onto the
	// one-shot next-from-source shield. It is only set alongside
	// PreventDamageNextFromSource.
	PreventDamageRedirectToSourceController bool
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
	// SubjectSourceAttached mirrors the parser flag: a resolving continuous effect
	// whose possessive subject is the source's attached permanent ("equipped
	// creature's"/"enchanted creature's"), lowering to SourceAttachedPermanentReference.
	SubjectSourceAttached bool
	// CoordinatedSourceSubject mirrors the parser flag: a power/toughness pump
	// whose subject coordinates the source permanent with a controlled creature
	// group ("Alandra and Drakes you control each get +X/+X …"). StaticSubject
	// carries the group's source-EXCLUDING variant; lowering pumps the source once
	// through a ModifyPT instruction and every other group member through the
	// excluding continuous effect.
	CoordinatedSourceSubject bool
	// DoubleSourceCounters mirrors the parser flag for an EffectDouble whose
	// object is "the number of <kind> counters on <self>" (Mossborn Hydra).
	// Lowering reads it together with DoubleSourceCounterKind to emit a dynamic
	// counter placement that adds counters equal to the source's current count,
	// doubling it; it is false for every other double effect.
	DoubleSourceCounters    bool
	DoubleSourceCounterKind counter.Kind
	// DoubleCountersTarget and DoubleCountersAllKinds mirror the parser flags for
	// the extended counter-doubling forms: a "target ..." object (Gilder Bairn)
	// and the "each kind of counter" all-kinds form (Vorel of the Hull Clade).
	// Lowering reads them to bind the doubling to the sentence's target and to
	// double every counter kind. Both are false for the self single-kind form.
	DoubleCountersTarget   bool
	DoubleCountersAllKinds bool
	// DoubleCountersGroup mirrors the parser flag for the group counter-doubling
	// form ("double the number of +1/+1 counters on each creature you control",
	// Bristly Bill, Spine Sower); the group rides on Selector. Lowering doubles
	// the single kind on each member. It is false for the source and target forms.
	DoubleCountersGroup bool
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
	// PunisherDiscardCount is the number of cards the discard alternative
	// requires; it mirrors the parser field and is only above 1 for an explicit
	// card count ("... unless they discard two cards.").
	PunisherDiscardCount int
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
	// PlayFromTopPayLife mirrors the parser flag for an EffectPlayFromLibraryTop
	// grant carrying the "If you cast a spell this way, pay life equal to its
	// mana value rather than pay its mana cost." rider, so spells cast from the
	// top of the library via the grant pay life equal to their mana value instead
	// of their mana cost. PlayFromTopPayLifeRiderSpan covers the rider sentence so
	// lowering credits it toward source coverage.
	PlayFromTopPayLife          bool
	PlayFromTopPayLifeRiderSpan shared.Span
	// AdditionalCombatPhase mirrors the parser flag for an "After this [main]
	// phase, there is an additional combat phase[ followed by an additional main
	// phase]." effect (Aggravated Assault, Aurelia the Warleader, World at War):
	// the effect inserts an extra combat phase into the current turn.
	// AdditionalMainPhase mirrors the optional "followed by an additional main
	// phase" tail. Both are false for every other effect; AdditionalMainPhase is
	// set only together with AdditionalCombatPhase.
	AdditionalCombatPhase bool
	AdditionalMainPhase   bool
	// AdditionalBeginningPhase mirrors the parser flag for a "there is an
	// additional beginning phase after this phase." effect (Sphinx of the Second
	// Sun, Temple of Atropos, Cyclonus, Cybertronian Fighter): the effect inserts
	// an extra beginning phase into the current turn. It is false for every other
	// effect and never set together with AdditionalCombatPhase or
	// AdditionalMainPhase.
	AdditionalBeginningPhase bool
	// DieSides is the number of faces of the die rolled by an EffectRollDie
	// effect ("roll a d20" sets DieSides to 20). It is zero for every other
	// effect kind.
	DieSides int
	// DiceRow marks an effect that belongs to a die-roll outcome-table row. Its
	// instruction is gated on the rolled value falling within the inclusive
	// interval [DiceRowMin, DiceRowMax]. It is false for every effect that is not
	// part of an outcome table.
	DiceRow    bool
	DiceRowMin int
	DiceRowMax int
	// CoinFlipBranch marks an effect that belongs to a coin-flip branch. A
	// CoinFlipBranchWin effect's instruction is gated on the flip coming up heads
	// and a CoinFlipBranchLose effect's on tails. It is CoinFlipBranchNone for
	// every effect that is not part of a coin flip.
	CoinFlipBranch CoinFlipBranch
	// VoteArm marks an effect that belongs to a vote arm. A marked effect's
	// instruction is gated on the vote tally satisfying the arm: VoteArmOption is
	// the option index whose votes the arm depends on, and VoteArmTieInclusive
	// reports whether a tied vote also satisfies it. VoteArm is false for every
	// effect that is not part of a vote.
	VoteArm             bool
	VoteArmOption       int
	VoteArmTieInclusive bool
	// RequireSourceTrample marks an excess-damage-to-controller redirect
	// (DynamicAmountExcessDamageDealtThisWay) that applies only when the damage
	// source has trample (Ram Through). Lowering gates the excess redirect on the
	// source having trample and emits a plain damage branch otherwise. It is
	// meaningful only when Amount.DynamicKind is DynamicAmountExcessDamageDealtThisWay.
	RequireSourceTrample bool
}

// EntersTappedGroup reports the enters-tapped-group form of a static group
// entry-modification replacement.
func (e *CompiledEffect) EntersTappedGroup() bool {
	return e.GroupEntryModification.Kind == parser.GroupEntryModificationTapped
}

// EntersWithCountersGroup reports the enters-with-counters-group form of a static
// group entry-modification replacement.
func (e *CompiledEffect) EntersWithCountersGroup() bool {
	return e.GroupEntryModification.Kind == parser.GroupEntryModificationWithCounters
}

// CoinFlipBranch identifies which branch of a recognized "Flip a coin." outcome
// an effect belongs to.
type CoinFlipBranch int

const (
	// CoinFlipBranchNone marks an effect that is not part of a coin flip.
	CoinFlipBranchNone CoinFlipBranch = iota
	// CoinFlipBranchWin marks an effect resolved when the controller wins the
	// flip (heads).
	CoinFlipBranchWin
	// CoinFlipBranchLose marks an effect resolved when the controller loses the
	// flip (tails).
	CoinFlipBranchLose
)

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
	// TriggerLandProducedType mirrors the parser's "one mana of any type that
	// land produced" mana-doubler body (Mirari's Wake, Zendikar Resurgent). See
	// parser.EffectManaSyntax.TriggerLandProducedType.
	TriggerLandProducedType bool
	// Combination, CombinationColors, CombinationCount, and CombinationDynamic
	// mirror the parser's "<N> mana in any combination of <colors>" body (Goblin
	// Clearcutter, Manamorphose, Cascading Cataracts). The produced mana is split
	// freely among CombinationColors; CombinationCount holds a fixed cardinal
	// amount (>= 2) while CombinationDynamic instead pairs the split with the
	// effect's dynamic Amount. See parser.EffectManaSyntax.Combination.
	Combination        bool
	CombinationColors  []mana.Color
	CombinationCount   int
	CombinationDynamic bool
	// PersistUntilEndOfTurn mirrors the parser's cross-sentence "Until end of
	// turn, you don't lose this mana as steps and phases end" rider (Grand
	// Warlord Radha): the produced mana does not empty as steps and phases end
	// for the rest of the turn. See parser.EffectManaSyntax.PersistUntilEndOfTurn.
	PersistUntilEndOfTurn bool
}

// CompiledEffectPayment is a typed resolution payment embedded in an effect.
type CompiledEffectPayment struct {
	Span              shared.Span
	Form              parser.EffectPaymentForm
	Payer             parser.EffectPaymentPayerKind
	ManaCost          cost.Mana
	GenericManaAmount CompiledAmount
	// AdditionalCost is a non-mana resolution payment cost (such as "sacrifice a
	// land" or the fixed life portion of "pay {mana} and N life"). It is nil for
	// mana-only payments. ManaCost and AdditionalCost are both set for a combined
	// mana+life payment; otherwise exactly one is set.
	AdditionalCost         *CompiledCost
	SuccessConditionNodeID int
	FailureConditionNodeID int
	// PerCreatureSelector is the folded creature filter of an
	// EffectPaymentFormPerChosenCreature offer: the payer pays ManaCost once for
	// each creature they choose from this selector. It is the zero selector for
	// every other payment form.
	PerCreatureSelector CompiledSelector
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
	StaticSubjectCounter *CompiledStaticSubjectCounter
	StaticSubjectPower   *CompiledStaticSubjectPower
	Symbol               string
}

// CompiledStaticSubjectType preserves a static subject's printed subtype and its
// parser-resolved canonical subtype when known. SubsAny carries a disjunctive
// list of subtypes when the subject names more than one ("... that's a Wolf or a
// Werewolf"); a permanent matches if it has any one of them. When SubsAny is set
// it supersedes the single Sub slot, which still holds the first entry.
type CompiledStaticSubjectType struct {
	Text     string
	Sub      types.Sub
	SubsAny  []types.Sub
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
	// ChosenColorFromEntry constrains the group to permanents whose color
	// matches the source permanent's entry-time color choice ("creatures you
	// control of the chosen color"). It is independent of the printed color
	// qualifiers above.
	ChosenColorFromEntry bool
}

// CompiledStaticSubjectKeyword preserves a static subject's optional single
// keyword filter ("Creatures with flying ...", "Creatures without flying ...").
// Excluded distinguishes the "without" exclusion from the "with" requirement.
type CompiledStaticSubjectKeyword struct {
	Keyword  parser.KeywordKind
	Excluded bool
}

// CompiledStaticSubjectCounter preserves a static subject's optional "with a
// <kind> counter on it/them" filter constraining the affected group to members
// carrying that counter ("Each creature you control with a +1/+1 counter on it
// has ..."). Kind names the required counter; Any marks the kind-agnostic "with
// a counter on it" qualifier (Rishkar), where any counter satisfies the filter.
type CompiledStaticSubjectCounter struct {
	Kind counter.Kind
	Any  bool
}

// CompiledStaticSubjectPower preserves a static subject's optional numeric power
// comparison filter constraining the affected group to members whose power meets
// the comparison ("Each creature you control with power 4 or greater gets ...",
// Goreclaw, Terror of Qal Sisma). Comparison is the recognized power bound.
type CompiledStaticSubjectPower struct {
	Comparison compare.Int
}

func staticSubjectType(text string, sub types.Sub, subsAny []types.Sub, known, excluded bool) *CompiledStaticSubjectType {
	if text == "" && !known {
		return nil
	}
	return &CompiledStaticSubjectType{Text: text, Sub: sub, SubsAny: append([]types.Sub(nil), subsAny...), Known: known, Excluded: excluded}
}

func staticSubjectColors(colors []parser.Color, colorless, multicolored, chosenColorFromEntry bool) *CompiledStaticSubjectColors {
	if len(colors) == 0 && !colorless && !multicolored && !chosenColorFromEntry {
		return nil
	}
	return &CompiledStaticSubjectColors{ColorsAny: colors, Colorless: colorless, Multicolored: multicolored, ChosenColorFromEntry: chosenColorFromEntry}
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

func staticSubjectCounter(required bool, kind counter.Kind, anyKind bool) *CompiledStaticSubjectCounter {
	if !required {
		return nil
	}
	return &CompiledStaticSubjectCounter{Kind: kind, Any: anyKind}
}

func staticSubjectPower(comparison compare.Int, match bool) *CompiledStaticSubjectPower {
	if !match {
		return nil
	}
	return &CompiledStaticSubjectPower{Comparison: comparison}
}

func compiledEffectDetails(staticType *CompiledStaticSubjectType, staticColors *CompiledStaticSubjectColors, staticKeyword *CompiledStaticSubjectKeyword, staticCounter *CompiledStaticSubjectCounter, staticPower *CompiledStaticSubjectPower, symbol string) *CompiledEffectDetails {
	if staticType == nil && staticColors == nil && staticKeyword == nil && staticCounter == nil && staticPower == nil && symbol == "" {
		return nil
	}
	return &CompiledEffectDetails{StaticSubjectType: staticType, StaticSubjectColors: staticColors, StaticSubjectKeyword: staticKeyword, StaticSubjectCounter: staticCounter, StaticSubjectPower: staticPower, Symbol: symbol}
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

// StaticSubjectSubsAny returns the static subject's disjunctive subtype list when
// the subject names more than one creature subtype ("... that's a Wolf or a
// Werewolf"). It is empty for the single-subtype subjects.
func (e *CompiledEffect) StaticSubjectSubsAny() []types.Sub {
	if e.Details == nil || e.Details.StaticSubjectType == nil {
		return nil
	}
	return e.Details.StaticSubjectType.SubsAny
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

// StaticSubjectChosenColorFromEntry reports whether the static subject is
// constrained to permanents whose color matches the source permanent's
// entry-time color choice ("creatures you control of the chosen color").
func (e *CompiledEffect) StaticSubjectChosenColorFromEntry() bool {
	return e.Details != nil && e.Details.StaticSubjectColors != nil && e.Details.StaticSubjectColors.ChosenColorFromEntry
}

// StaticSubjectKeyword returns the static subject's optional single keyword
// filter, whether it is an exclusion, and whether any keyword filter is present.
func (e *CompiledEffect) StaticSubjectKeyword() (keyword parser.KeywordKind, excluded, present bool) {
	if e.Details == nil || e.Details.StaticSubjectKeyword == nil {
		return parser.KeywordUnknown, false, false
	}
	return e.Details.StaticSubjectKeyword.Keyword, e.Details.StaticSubjectKeyword.Excluded, true
}

// StaticSubjectCounter returns the static subject's optional "with a <kind>
// counter on it/them" filter kind, whether the qualifier is kind-agnostic ("with
// a counter on it"), and whether any counter filter is present.
func (e *CompiledEffect) StaticSubjectCounter() (kind counter.Kind, anyKind, present bool) {
	if e.Details == nil || e.Details.StaticSubjectCounter == nil {
		return 0, false, false
	}
	return e.Details.StaticSubjectCounter.Kind, e.Details.StaticSubjectCounter.Any, true
}

// StaticSubjectPower returns the static subject's optional numeric power
// comparison filter ("... with power 4 or greater ...") and whether one is
// present.
func (e *CompiledEffect) StaticSubjectPower() (comparison compare.Int, present bool) {
	if e.Details == nil || e.Details.StaticSubjectPower == nil {
		return compare.Int{}, false
	}
	return e.Details.StaticSubjectPower.Comparison, true
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
	// DynamicAmountTriggeringCombatDamage is the amount of combat damage dealt
	// by the event that triggered the enclosing combat-damage trigger ("that
	// many" in "Whenever a creature you control deals combat damage to a player,
	// create that many Treasure tokens."). It backs the "create that many
	// <predefined> tokens" family (Old Gnawbone). Added last so existing kinds
	// keep their wire values.
	DynamicAmountTriggeringCombatDamage
	// DynamicAmountDestroyedThisWay is the number of permanents destroyed by the
	// immediately preceding destroy effect in the same ability ("for each
	// permanent destroyed this way"). It backs the mass-destroy payoff family
	// (Fumigate, Multani's Decree, Death Begets Life) and is realized by a
	// sequence lowerer that reads the count published by the preceding destroy
	// instruction. Added last so existing kinds keep their wire values.
	DynamicAmountDestroyedThisWay
	// DynamicAmountLifeLostThisTurn is the total life the controller has lost so
	// far this turn ("equal to the life you've lost this turn"). Damage to the
	// controller counts because dealing damage to a player causes that player to
	// lose that much life (CR 120.3). It backs Children of Korlis.
	// DynamicAmountLifeGainedThisTurn is the life-gained sibling. Both are
	// controller-scoped. Added last so existing kinds keep their wire values.
	DynamicAmountLifeLostThisTurn
	DynamicAmountLifeGainedThisTurn
	// DynamicAmountTriggeringPlayerHandSize is the number of cards in the
	// triggering event player's hand ("equal to the number of cards in their
	// hand", Emberwilde Captain). It lowers to a card-count in the event player's
	// hand zone. Added last so existing kinds keep their wire values.
	DynamicAmountTriggeringPlayerHandSize
	// DynamicAmountMaxOf is the greatest value among Operands, the "whichever is
	// greater" combinator over two rules-derived amounts ("equal to the amount
	// of life you gained this turn or the amount of life you lost this turn,
	// whichever is greater." — Willowdusk, Essence Seer). Each operand is itself
	// a CompiledAmount. Added last so existing kinds keep their wire values.
	DynamicAmountMaxOf
	// DynamicAmountTriggeringCounterCount is the number of counters added by the
	// event that triggered the enclosing counter-placement trigger ("that many"
	// in "Whenever you put one or more +1/+1 counters on a creature you control,
	// you may draw that many cards." — Terrasymbiosis). It backs the "draw that
	// many cards" counter payoff; the lowerer accepts it only inside a
	// counter-placement trigger. Added last so existing kinds keep their wire
	// values.
	DynamicAmountTriggeringCounterCount
	// DynamicAmountColorsOfManaSpent is the number of distinct colors of mana
	// spent to cast the source spell ("for each color of mana spent to cast it"
	// — the Converge count). It backs the Converge enters-with-counters quantity
	// (Crystalline Crawler); the runtime records the colors of mana spent as the
	// spell's costs are paid and carries the count to the entering permanent.
	// Added last so existing kinds keep their wire values.
	DynamicAmountColorsOfManaSpent
	// DynamicAmountDieRollResult is the value produced by the immediately
	// preceding die roll in the same ability ("a number of Treasure tokens equal
	// to the result." — Ancient Copper Dragon). The lowerer reads the count the
	// EffectRollDie effect publishes. Added last so existing kinds keep their
	// wire values.
	DynamicAmountDieRollResult
	// DynamicAmountTotalManaValue is the sum of mana value across the selector's
	// battlefield group ("the total mana value of <group>"). It backs the dynamic
	// "where X is the total mana value of noncreature artifacts you control" cost
	// reduction (Metalwork Colossus, Earthquake Dragon, Excalibur, Sword of Eden).
	// Added last so existing kinds keep their wire values.
	DynamicAmountTotalManaValue
	// DynamicAmountTimesKicked is the number of times the source spell was kicked
	// (its Multikicker count, "for each time it was kicked"). It backs Everflowing
	// Chalice's enters-with-counters quantity and Wolfbriar Elemental's Wolf-token
	// count; the runtime records the kick count as the spell is cast. Added last
	// so existing kinds keep their wire values.
	DynamicAmountTimesKicked
	// DynamicAmountOpponentsAttackedThisCombat is the number of the controller's
	// opponents being attacked this combat by creatures the controller controls
	// ("for each opponent you attacked this combat", the Melee count). It is the
	// combat-state sibling of DynamicAmountOpponentCount. Added last so existing
	// kinds keep their wire values.
	DynamicAmountOpponentsAttackedThisCombat
	// DynamicAmountControllerSpeed is the resolving ability controller's current
	// speed ("your speed", the Start your engines! subsystem, CR 702.179). It is
	// controller-scoped; a player with no speed reads zero and speed caps at 4.
	// It backs "where X is your speed" amounts such as The Speed Demon. Added
	// last so existing kinds keep their wire values.
	DynamicAmountControllerSpeed
	// DynamicAmountOpponentControllingCount is the number of the resolving ability
	// controller's opponents who control at least one permanent matching the
	// amount's selector ("the number of opponents who control a creature with
	// power 4 or greater", Summon: Yojimbo chapter IV). The selector is the
	// per-opponent control predicate, evaluated relative to each opponent; it is a
	// player count, not a board count. Added last so existing kinds keep their
	// wire values.
	DynamicAmountOpponentControllingCount
	// DynamicAmountTriggeringEventAmount is the quantity carried by the event
	// that triggered the enclosing trigger, resolved by lowering to the matching
	// per-event amount: combat damage dealt, life gained or lost, counters
	// added, or cards drawn or discarded. It backs the "put that many <kind>
	// counters on <this creature|it>" counter-placement payoff family (Marauding
	// Mako, Necropolis Regent, Ageless Entity, Bioessence Hydra), where "that
	// many" reads whatever the trigger measured. Lowering fails closed outside a
	// trigger whose event publishes such a quantity. Added last so existing
	// kinds keep their wire values.
	DynamicAmountTriggeringEventAmount
	// DynamicAmountCardsDrawnThisTurn is the number of cards the resolving
	// ability's controller has drawn so far this turn ("the number of cards
	// you've drawn this turn"). It backs the draw-payoff family (Thundering
	// Djinn's attack-trigger damage, Duelist of the Mind's characteristic-defining
	// power); the triggering or just-resolved draw counts, since its draw event
	// precedes the resolving ability. It is controller-scoped and carries no
	// in-text referent. Added last so existing kinds keep their wire values.
	DynamicAmountCardsDrawnThisTurn
	// DynamicAmountCardsNamedSelfInGraveyards is the number of cards in every
	// graveyard whose name matches the card's own name ("for each card named
	// Rite of Flame in each graveyard", Rite of Flame). The parser recognizes
	// only the card's own self-name, so the count is the self-named graveyard
	// total across all players; the source name is read at resolution. Added last
	// so existing kinds keep their wire values.
	DynamicAmountCardsNamedSelfInGraveyards
	// DynamicAmountCardsNamedSelfInControllerGraveyard is the number of cards in
	// the controller's graveyard whose name matches the card's own name ("for
	// each card named Compound Fracture in your graveyard", Compound Fracture).
	// The parser recognizes only the card's own self-name, so the count is the
	// self-named total in the controller's graveyard alone, unlike
	// DynamicAmountCardsNamedSelfInGraveyards, which counts every graveyard; the
	// source name is read at resolution. Added last so existing kinds keep their
	// wire values.
	DynamicAmountCardsNamedSelfInControllerGraveyard
	// DynamicAmountHalfPlayerLibrary is half the number of cards in the milling
	// player's library, rounded up or down per CompiledAmount.RoundUp ("mills
	// half their library, rounded down" — Traumatize; "rounded up" — Fleet
	// Swallower; CR 107.4, CR 701.13). The milling player is the effect's subject
	// (target/defending player); lowering counts that player's library. Added
	// last so existing kinds keep their wire values.
	DynamicAmountHalfPlayerLibrary
	// DynamicAmountDamageDealtThisWay is the damage dealt by the immediately
	// preceding damage effect in the same ability ("equal to the damage dealt
	// this way"). Lowering publishes the dealt amount from the damage instruction
	// and consumes it via this amount on a follow-on life gain (drain spells such
	// as Corrupt). Added last so existing kinds keep their wire values.
	DynamicAmountDamageDealtThisWay
	// DynamicAmountExcessDamageDealtThisWay is the excess damage dealt by the
	// immediately preceding damage effect in the same ability ("equal to the
	// excess damage dealt this way") — only the damage beyond what was needed to
	// destroy the recipient. Lowering publishes the excess amount from the damage
	// instruction and consumes it via this amount on a follow-on life gain. Added
	// last so existing kinds keep their wire values.
	DynamicAmountExcessDamageDealtThisWay
	// DynamicAmountCommanderCastCount is the number of times the controller has
	// cast their commander from the command zone this game ("for each time you've
	// cast your commander from the command zone this game"). It backs the
	// command-zone-cast anthem family (Commander's Insignia; Vanguard of the
	// Restless); lowering reads the controller's commander cast count. Added last
	// so existing kinds keep their wire values.
	DynamicAmountCommanderCastCount
	// DynamicAmountReferencedPlayerLifeLostThisTurn is the total life the player
	// named by "that player" lost so far this turn ("target opponent loses life
	// equal to the life that player lost this turn", Blitzwing, Cruel Tormentor).
	// DynamicAmountReferencedPlayerLifeGainedThisTurn is the life-gained sibling.
	// Unlike the controller-scoped DynamicAmountLifeLostThisTurn family, "that
	// player" co-refers with the effect's referenced/target player, so lowering
	// binds the count to that player. Added last so existing kinds keep their
	// wire values.
	DynamicAmountReferencedPlayerLifeLostThisTurn
	DynamicAmountReferencedPlayerLifeGainedThisTurn
	// DynamicAmountCreaturesBlockingSource is the number of creatures blocking
	// the permanent the amount is evaluated for ("for each creature blocking it"
	// — Rabid Elephant, Gang of Elk, Sparring Golem, Elvish Berserker). The "it"
	// names the just-blocked permanent that receives the pump, so lowering binds
	// the count to that pump's object; the runtime reads the current combat's
	// block declarations. It is a combat-state amount, not a board count. Added
	// last so existing kinds keep their wire values.
	DynamicAmountCreaturesBlockingSource
	// DynamicAmountHalfPlayerLife is half the life total of the player who loses
	// the life, rounded up or down per CompiledAmount.RoundUp ("that player loses
	// half their life, rounded up" — Quietus Spike, Virtus the Veiled,
	// Scytheclaw; CR 107.4). The losing player is the effect's subject
	// (target/event player); lowering reads that player's life. Added last so
	// existing kinds keep their wire values.
	DynamicAmountHalfPlayerLife
	// DynamicAmountPartySize is the controller's maximum filled party roles
	// (Cleric, Rogue, Warrior, Wizard), one role per creature.
	DynamicAmountPartySize
	// DynamicAmountDamagePreventedThisWay is the amount of damage prevented by
	// the same card's earlier prevention clause ("For each 1 damage prevented
	// this way, create ..." — Inkshield). Lowering schedules the payoff to
	// resolve after the prevention has applied and reads the running prevented
	// total. Added last so existing kinds keep their wire values.
	DynamicAmountDamagePreventedThisWay
)

// DynamicAmountForm identifies the exact Oracle formula used for an amount.
type DynamicAmountForm uint8

// Dynamic amount forms recognized by the semantic compiler.
const (
	DynamicAmountFormNone DynamicAmountForm = iota
	DynamicAmountEqual
	DynamicAmountForEach
	DynamicAmountWhereX
	// DynamicAmountFormHalfLibrary introduces the "half their library, rounded
	// up/down" mill amount, whose noun is the milling player's library rather
	// than a counted "cards" plural. Added last so existing forms keep their wire
	// values.
	DynamicAmountFormHalfLibrary
)

// CompiledAmount is a fixed or rules-derived amount recognized in an effect.
type CompiledAmount struct {
	Value      int
	Known      bool
	RangeKnown bool
	Minimum    int
	Maximum    int
	VariableX  bool
	// AnyNumber records the unbounded "any number of <noun>" count form; see
	// parser.EffectAmountSyntax.AnyNumber. It is the only positive signal for
	// that form, since "all", "the", and a bare plural noun share the same empty
	// amount shape.
	AnyNumber   bool
	DynamicKind DynamicAmountKind
	DynamicForm DynamicAmountForm
	Multiplier  int
	// RoundUp records that a half-library mill amount rounds up rather than down
	// (DynamicAmountHalfPlayerLibrary). It is false for every other amount.
	RoundUp       bool
	ReferenceSpan shared.Span
	Addend        int
	CounterKind   counter.Kind
	Text          string
	// Colors carries the colors of a devotion amount; empty otherwise.
	Colors []color.Color
	// Operands carries the sub-amounts of a DynamicAmountMaxOf combinator; empty
	// otherwise.
	Operands []CompiledAmount
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
	// YouControl restricts the permanent target to one the enchanting player
	// controls ("Enchant creature or planeswalker you control"). It is set only
	// alongside a card-type/subtype predicate, never with Player or Opponent.
	YouControl bool
}

// CompiledKeyword is a recognized keyword ability.
type CompiledKeyword struct {
	Kind          parser.KeywordKind
	Name          string
	Span          shared.Span
	Text          string
	Parameter     string
	ParameterKind parser.KeywordParameterKind
	ManaCost      cost.Mana
	// WardCost is the typed non-mana or composite payment of a "Ward—<cost>"
	// keyword, or nil for a mana-only Ward whose cost is ManaCost. Its components
	// are lowered through the shared activation-cost kernel into the runtime
	// ward's mana and additional costs.
	WardCost        *CompiledCost
	Integer         int
	EnchantTarget   CompiledEnchantTarget
	Protection      game.ProtectionKeyword
	ProtectionKnown bool
	// Gift is the typed gift a Gift keyword action promises (CR 702.171), or
	// GiftKindNone when the keyword is not Gift. Lowering maps it to the delivery
	// content given to the promised opponent (draw a card / create a token).
	Gift parser.GiftKind
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
	// ReferenceDiedCreature is the explicit "the creature that died" wording, the
	// triggering creature of a dies trigger. Unlike the demonstrative
	// ReferenceThatObject it never names a target, so bindReferences binds it
	// straight to the event permanent rather than a target antecedent.
	ReferenceDiedCreature
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
	// ReferenceBindingEventStackObject binds "that spell"/"it" in a spell-cast
	// trigger body to the spell that was cast. At runtime
	// EventStackObjectReference() resolves this to the triggering event's stack
	// object ("Whenever you cast a spell ..., copy that spell.").
	ReferenceBindingEventStackObject
	// ReferenceBindingSourceAttached binds the permanent the source is attached
	// to, used by Equipment and Aura conditions ("As long as equipped creature
	// is legendary, ..."). At runtime SourceAttachedPermanentReference() resolves
	// this through the source's AttachedTo link.
	ReferenceBindingSourceAttached
	// ReferenceBindingCreatedToken binds "the token" in a resolving condition to
	// a token a prior effect in the same ability just created (Yenna, Redtooth
	// Regent: "If the token is an Aura, ..."). At runtime the lowering resolves
	// this through the linked object the creating effect published.
	ReferenceBindingCreatedToken
	// ReferenceBindingEventRelatedPermanent binds the object demonstrative "that
	// creature" in a combat block trigger body to the triggering event's related
	// permanent, the other creature in the combat ("Whenever this creature blocks
	// or becomes blocked by a creature, ~ deals N damage to that creature.",
	// Inferno Elemental). At runtime EventRelatedPermanentReference() resolves
	// this through the event's RelatedPermanentID, which the block and
	// became-blocked events populate with the opposing combatant.
	ReferenceBindingEventRelatedPermanent
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
