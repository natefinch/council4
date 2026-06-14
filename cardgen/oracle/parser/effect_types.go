package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// EffectKind identifies a resolving instruction. The parser owns the Oracle
// vocabulary which selects these values; consumers only map the typed value.
type EffectKind uint8

// Resolving effect kinds recognized by the parser.
const (
	EffectUnknown EffectKind = iota
	EffectAddMana
	EffectAttach
	EffectCast
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
	EffectGainControl
	EffectGrantKeyword
	EffectInvestigate
	EffectExplore
	EffectLose
	EffectManifest
	EffectManifestDread
	EffectMill
	EffectModifyPT
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

// EffectDurationKind identifies a resolving effect's duration.
type EffectDurationKind uint8

// Resolving effect durations recognized by the parser.
const (
	EffectDurationNone EffectDurationKind = iota
	EffectDurationUntilEndOfTurn
	EffectDurationUntilYourNextTurn
	EffectDurationThisTurn
	EffectDurationThisCombat
	EffectDurationWhileSourceOnBattlefield
	EffectDurationWhileYouControlSource
)

// DelayedTimingKind identifies a delayed resolving instruction suffix.
type DelayedTimingKind uint8

// Delayed timings recognized by resolving-effect grammar.
const (
	DelayedTimingNone DelayedTimingKind = iota
	DelayedTimingNextEndStep
	DelayedTimingNextUpkeep
)

// EffectDestinationPosition identifies an ordered position in a destination
// zone.
type EffectDestinationPosition uint8

// Ordered destination positions recognized by resolving-effect grammar.
const (
	EffectDestinationUnspecified EffectDestinationPosition = iota
	EffectDestinationTop
	EffectDestinationBottom
)

// EffectDynamicAmountKind identifies a rules-derived amount.
type EffectDynamicAmountKind uint8

// Dynamic resolving amounts recognized by the parser.
const (
	EffectDynamicAmountNone EffectDynamicAmountKind = iota
	EffectDynamicAmountCount
	EffectDynamicAmountControllerLife
	EffectDynamicAmountOpponentCount
	EffectDynamicAmountSourcePower
	EffectDynamicAmountBasicLandTypes
)

// EffectDynamicAmountForm identifies how a dynamic amount is introduced.
type EffectDynamicAmountForm uint8

// Dynamic amount forms recognized by the parser.
const (
	EffectDynamicAmountFormNone EffectDynamicAmountForm = iota
	EffectDynamicAmountFormEqual
	EffectDynamicAmountFormForEach
	EffectDynamicAmountFormWhereX
)

// EffectAmountSyntax is a fixed or rules-derived source-spanned amount.
type EffectAmountSyntax struct {
	Span          shared.Span
	Text          string
	Value         int
	Known         bool
	VariableX     bool
	DynamicKind   EffectDynamicAmountKind
	DynamicForm   EffectDynamicAmountForm
	Multiplier    int
	ReferenceSpan shared.Span
	Selection     *SelectionSyntax
}

// EffectReplacementKind identifies how an instruction replaces an event.
type EffectReplacementKind uint8

// Resolving replacement modifiers recognized by the parser.
const (
	EffectReplacementNone EffectReplacementKind = iota
	EffectReplacementInstead
	EffectReplacementTwiceThatMany
	EffectReplacementThatMuchPlus
	EffectReplacementDoubleThat
)

// EffectReplacementSyntax is a source-spanned replacement modifier.
type EffectReplacementSyntax struct {
	Kind            EffectReplacementKind
	Span            shared.Span
	Amount          int
	EachCounterKind bool
}

// EffectManaSyntax describes exact add-mana output.
type EffectManaSyntax struct {
	Span            shared.Span
	Symbols         []string
	Choice          bool
	AnyColor        bool
	LegacyBodyExact bool
}

// EffectContextKind identifies the grammatical subject performing or receiving
// a resolving instruction.
type EffectContextKind uint8

// Resolving-effect contexts recognized by the parser.
const (
	EffectContextUnknown EffectContextKind = iota
	EffectContextController
	EffectContextTarget
	EffectContextEachOpponent
	EffectContextEachPlayer
	EffectContextEventPlayer
	EffectContextSource
	EffectContextReferencedObject
	EffectContextReferencedPlayer
	EffectContextPriorSubject
)

// SignedAmountSyntax is one signed half of a power/toughness change.
type SignedAmountSyntax struct {
	Span     shared.Span
	Value    int
	Known    bool
	Negative bool
}

// SelectionController identifies a selected object's controller.
type SelectionController uint8

// Selection controller relations.
const (
	SelectionControllerAny SelectionController = iota
	SelectionControllerYou
	SelectionControllerOpponent
	SelectionControllerNotYou
)

// SelectionKind identifies the broad object selected by a phrase.
type SelectionKind uint8

// Selection kinds recognized by resolving-effect grammar.
const (
	SelectionUnknown SelectionKind = iota
	SelectionAny
	SelectionPlayer
	SelectionOpponent
	SelectionArtifact
	SelectionCreature
	SelectionEnchantment
	SelectionLand
	SelectionPermanent
	SelectionCard
	SelectionSpell
	SelectionActivatedAbility
	SelectionTriggeredAbility
	SelectionActivatedOrTriggeredAbility
	SelectionSpellActivatedOrTriggeredAbility
	SelectionPlaneswalker
	SelectionBattle
)

// SelectionSyntax is a typed, source-spanned noun phrase.
type SelectionSyntax struct {
	Span             shared.Span
	Text             string
	Kind             SelectionKind
	Controller       SelectionController
	All              bool
	Another          bool
	Other            bool
	Attacking        bool
	Blocking         bool
	Tapped           bool
	Untapped         bool
	Keyword          KeywordKind
	Zone             zone.Type
	RequiredTypesAny []CardType
	ExcludedTypes    []CardType
	Supertypes       []Supertype
	ColorsAny        []Color
	ExcludedColors   []Color
	SubtypesAny      []types.Sub
	ManaValue        compare.Int
	MatchManaValue   bool
	Power            compare.Int
	MatchPower       bool
	Toughness        compare.Int
	MatchToughness   bool
}

// TargetCardinalitySyntax is an inclusive target-count range.
type TargetCardinalitySyntax struct {
	Min int
	Max int
}

// TargetSyntax is one typed target production.
type TargetSyntax struct {
	Span        shared.Span
	Text        string
	Cardinality TargetCardinalitySyntax
	Selection   SelectionSyntax
	Exact       bool
}

// EffectConnectionKind identifies how a resolving instruction is coordinated
// with the preceding instruction in the same sentence.
type EffectConnectionKind uint8

// Resolving-instruction connections recognized by the parser.
const (
	EffectConnectionNone EffectConnectionKind = iota
	EffectConnectionAnd
	EffectConnectionThen
)

// EffectSyntax is one typed resolving instruction. Text and Tokens remain
// lossless metadata; all meaning consumed downstream is carried by typed fields.
type EffectSyntax struct {
	Kind                    EffectKind
	Context                 EffectContextKind
	Connection              EffectConnectionKind
	ConnectionSpan          shared.Span
	Span                    shared.Span
	VerbSpan                shared.Span
	ClauseSpan              shared.Span
	Text                    string
	Tokens                  []shared.Token
	Duration                EffectDurationKind
	DelayedTiming           DelayedTimingKind
	Selection               SelectionSyntax
	Amount                  EffectAmountSyntax
	PowerDelta              SignedAmountSyntax
	ToughnessDelta          SignedAmountSyntax
	StaticSubject           EffectStaticSubjectSyntax
	CounterKind             counter.Kind
	CounterKnown            bool
	FromZone                zone.Type
	ToZone                  zone.Type
	Destination             EffectDestinationPosition
	EntersTapped            bool
	EntersTappedSelf        bool
	EntersWithCounters      bool
	UnderYourControl        bool
	CastAsAdventure         bool
	Negated                 bool
	Optional                bool
	OptionalSpan            shared.Span
	Symbol                  string
	Mana                    EffectManaSyntax
	Replacement             EffectReplacementSyntax
	References              []Reference
	SubjectReferences       []Reference
	Targets                 []TargetSyntax
	SubjectTargets          []TargetSyntax
	Payment                 EffectPaymentSyntax
	Exact                   bool
	RequiresOrderedLowering bool
	HasUnrecognizedSibling  bool
	UnsupportedDetail       string
}

// EffectPaymentPayerKind identifies who may pay a cost embedded in an effect.
type EffectPaymentPayerKind uint8

// Embedded-effect payers recognized by the parser.
const (
	EffectPaymentPayerUnknown EffectPaymentPayerKind = iota
	EffectPaymentPayerTargetController
)

// EffectPaymentSyntax is a source-spanned typed resolution payment.
type EffectPaymentSyntax struct {
	Span     shared.Span
	Payer    EffectPaymentPayerKind
	ManaCost cost.Mana
}

// EffectStaticSubjectKind identifies the group affected by a static resolving
// effect production.
type EffectStaticSubjectKind uint8

// Static effect subjects recognized by resolving-effect grammar.
const (
	EffectStaticSubjectNone EffectStaticSubjectKind = iota
	EffectStaticSubjectAttachedObject
	EffectStaticSubjectControlledCreatures
	EffectStaticSubjectOtherControlledCreatures
	EffectStaticSubjectControlledWalls
	EffectStaticSubjectControlledArtifacts
	EffectStaticSubjectControlledTokens
	EffectStaticSubjectOpponentControlledCreatures
	EffectStaticSubjectControlledCreatureSubtype
	EffectStaticSubjectOtherControlledCreatureSubtype
)

// EffectStaticSubjectSyntax is a source-spanned typed static-effect subject.
type EffectStaticSubjectSyntax struct {
	Kind         EffectStaticSubjectKind
	Span         shared.Span
	Subtype      types.Sub
	SubtypeText  string
	SubtypeKnown bool
}
