package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// EffectKind identifies a resolving instruction. The parser owns the Oracle
// vocabulary which selects these values; consumers only map the typed value.
type EffectKind string

// Resolving effect kinds recognized by the parser.
const (
	EffectUnknown       EffectKind = ""
	EffectAddMana       EffectKind = "EffectAddMana"
	EffectAttach        EffectKind = "EffectAttach"
	EffectCast          EffectKind = "EffectCast"
	EffectCounter       EffectKind = "EffectCounter"
	EffectCreate        EffectKind = "EffectCreate"
	EffectDealDamage    EffectKind = "EffectDealDamage"
	EffectDestroy       EffectKind = "EffectDestroy"
	EffectDiscard       EffectKind = "EffectDiscard"
	EffectDiscover      EffectKind = "EffectDiscover"
	EffectDouble        EffectKind = "EffectDouble"
	EffectDraw          EffectKind = "EffectDraw"
	EffectEnterTapped   EffectKind = "EffectEnterTapped"
	EffectEnterPrepared EffectKind = "EffectEnterPrepared"
	EffectExile         EffectKind = "EffectExile"
	EffectFight         EffectKind = "EffectFight"
	EffectGain          EffectKind = "EffectGain"
	EffectGainControl   EffectKind = "EffectGainControl"
	EffectGrantKeyword  EffectKind = "EffectGrantKeyword"
	EffectInvestigate   EffectKind = "EffectInvestigate"
	EffectExplore       EffectKind = "EffectExplore"
	EffectLose          EffectKind = "EffectLose"
	EffectManifest      EffectKind = "EffectManifest"
	EffectManifestDread EffectKind = "EffectManifestDread"
	EffectMill          EffectKind = "EffectMill"
	EffectModifyPT      EffectKind = "EffectModifyPT"
	EffectPut           EffectKind = "EffectPut"
	EffectProliferate   EffectKind = "EffectProliferate"
	EffectRegenerate    EffectKind = "EffectRegenerate"
	EffectReturn        EffectKind = "EffectReturn"
	EffectReveal        EffectKind = "EffectReveal"
	EffectSacrifice     EffectKind = "EffectSacrifice"
	EffectScry          EffectKind = "EffectScry"
	EffectSurveil       EffectKind = "EffectSurveil"
	EffectSearch        EffectKind = "EffectSearch"
	EffectShuffle       EffectKind = "EffectShuffle"
	EffectTap           EffectKind = "EffectTap"
	EffectUntap         EffectKind = "EffectUntap"
	EffectTransform     EffectKind = "EffectTransform"
)

// EffectDurationKind identifies a resolving effect's duration.
type EffectDurationKind string

// Resolving effect durations recognized by the parser.
const (
	EffectDurationNone                     EffectDurationKind = ""
	EffectDurationUntilEndOfTurn           EffectDurationKind = "EffectDurationUntilEndOfTurn"
	EffectDurationUntilYourNextTurn        EffectDurationKind = "EffectDurationUntilYourNextTurn"
	EffectDurationThisTurn                 EffectDurationKind = "EffectDurationThisTurn"
	EffectDurationThisCombat               EffectDurationKind = "EffectDurationThisCombat"
	EffectDurationWhileSourceOnBattlefield EffectDurationKind = "EffectDurationWhileSourceOnBattlefield"
	EffectDurationWhileYouControlSource    EffectDurationKind = "EffectDurationWhileYouControlSource"
	// EffectDurationWhileControlledCreatureEnchanted matches the
	// attachment-dependent wording "for as long as that creature is enchanted".
	// The effect expires when the affected creature is no longer enchanted.
	EffectDurationWhileControlledCreatureEnchanted EffectDurationKind = "EffectDurationWhileControlledCreatureEnchanted"
)

// DelayedTimingKind identifies a delayed resolving instruction suffix.
type DelayedTimingKind string

// Delayed timings recognized by resolving-effect grammar.
const (
	DelayedTimingNone        DelayedTimingKind = ""
	DelayedTimingNextEndStep DelayedTimingKind = "DelayedTimingNextEndStep"
	DelayedTimingNextUpkeep  DelayedTimingKind = "DelayedTimingNextUpkeep"
)

// EffectDestinationPosition identifies an ordered position in a destination
// zone.
type EffectDestinationPosition string

// Ordered destination positions recognized by resolving-effect grammar.
const (
	EffectDestinationUnspecified EffectDestinationPosition = ""
	EffectDestinationTop         EffectDestinationPosition = "EffectDestinationTop"
	EffectDestinationBottom      EffectDestinationPosition = "EffectDestinationBottom"
)

// EffectDynamicAmountKind identifies a rules-derived amount.
type EffectDynamicAmountKind string

// Dynamic resolving amounts recognized by the parser.
const (
	EffectDynamicAmountNone           EffectDynamicAmountKind = ""
	EffectDynamicAmountCount          EffectDynamicAmountKind = "EffectDynamicAmountCount"
	EffectDynamicAmountControllerLife EffectDynamicAmountKind = "EffectDynamicAmountControllerLife"
	EffectDynamicAmountOpponentCount  EffectDynamicAmountKind = "EffectDynamicAmountOpponentCount"
	EffectDynamicAmountSourcePower    EffectDynamicAmountKind = "EffectDynamicAmountSourcePower"
	EffectDynamicAmountBasicLandTypes EffectDynamicAmountKind = "EffectDynamicAmountBasicLandTypes"
	EffectDynamicAmountEventCardCount EffectDynamicAmountKind = "EffectDynamicAmountEventCardCount"
)

// EffectDynamicAmountForm identifies how a dynamic amount is introduced.
type EffectDynamicAmountForm string

// Dynamic amount forms recognized by the parser.
const (
	EffectDynamicAmountFormNone    EffectDynamicAmountForm = ""
	EffectDynamicAmountFormEqual   EffectDynamicAmountForm = "EffectDynamicAmountFormEqual"
	EffectDynamicAmountFormForEach EffectDynamicAmountForm = "EffectDynamicAmountFormForEach"
	EffectDynamicAmountFormWhereX  EffectDynamicAmountForm = "EffectDynamicAmountFormWhereX"
)

// EffectAmountSyntax is a fixed or rules-derived source-spanned amount.
type EffectAmountSyntax struct {
	Span          shared.Span             `json:"-"`
	Text          string                  `json:",omitempty"`
	Value         int                     `json:",omitempty"`
	Known         bool                    `json:",omitempty"`
	VariableX     bool                    `json:",omitempty"`
	DynamicKind   EffectDynamicAmountKind `json:",omitempty"`
	DynamicForm   EffectDynamicAmountForm `json:",omitempty"`
	Multiplier    int                     `json:",omitempty"`
	ReferenceSpan shared.Span             `json:"-"`
	Selection     *SelectionSyntax        `json:",omitempty"`
}

// EffectReplacementKind identifies how an instruction replaces an event.
type EffectReplacementKind string

// Resolving replacement modifiers recognized by the parser.
const (
	EffectReplacementNone          EffectReplacementKind = ""
	EffectReplacementInstead       EffectReplacementKind = "EffectReplacementInstead"
	EffectReplacementTwiceThatMany EffectReplacementKind = "EffectReplacementTwiceThatMany"
	EffectReplacementThatMuchPlus  EffectReplacementKind = "EffectReplacementThatMuchPlus"
	EffectReplacementDoubleThat    EffectReplacementKind = "EffectReplacementDoubleThat"
)

// EffectReplacementSyntax is a source-spanned replacement modifier.
type EffectReplacementSyntax struct {
	Kind            EffectReplacementKind `json:",omitempty"`
	Span            shared.Span           `json:"-"`
	Amount          int                   `json:",omitempty"`
	EachCounterKind bool                  `json:",omitempty"`
}

// EffectManaSyntax describes exact add-mana output.
type EffectManaSyntax struct {
	Span    shared.Span `json:"-"`
	Symbols []string    `json:",omitempty"`
	// Colors are the typed mana colors recognized from Symbols, in order, when
	// every symbol is a basic color token ({W}{U}{B}{R}{G}{C}). They let a
	// consumer build add-mana content from typed values instead of re-parsing the
	// rendered symbol strings. Colors is populated only when ColorsKnown is true.
	Colors      []mana.Color `json:"-"`
	ColorsKnown bool         `json:",omitempty"`
	Choice      bool         `json:",omitempty"`
	AnyColor    bool         `json:",omitempty"`
	// ChosenColor reports the exact body "one mana of the chosen color", which
	// adds one mana of the color chosen as the source permanent entered (CR
	// 614.12) rather than a fixed or freely-chosen color.
	ChosenColor bool `json:",omitempty"`
	// ChosenColorFixed is the fixed alternative basic color of the composite body
	// "{C} or one mana of the chosen color." (the Gate/Thriving land cycle, e.g.
	// "Add {W} or one mana of the chosen color."). It is set together with
	// ChosenColor and ChosenColorFixedKnown; it is empty for the plain chosen
	// color body.
	ChosenColorFixed      mana.Color `json:"-"`
	ChosenColorFixedKnown bool       `json:",omitempty"`
	// CommanderIdentity reports the exact body "one mana of any color in your
	// commander's color identity" (CR 903.4). The choosable colors are the
	// controller's commander color identity, resolved dynamically at activation.
	CommanderIdentity bool `json:",omitempty"`
	LegacyBodyExact   bool `json:",omitempty"`
}

// EffectContextKind identifies the grammatical subject performing or receiving
// a resolving instruction.
type EffectContextKind string

// Resolving-effect contexts recognized by the parser.
const (
	EffectContextUnknown          EffectContextKind = ""
	EffectContextController       EffectContextKind = "EffectContextController"
	EffectContextTarget           EffectContextKind = "EffectContextTarget"
	EffectContextEachOpponent     EffectContextKind = "EffectContextEachOpponent"
	EffectContextEachPlayer       EffectContextKind = "EffectContextEachPlayer"
	EffectContextEventPlayer      EffectContextKind = "EffectContextEventPlayer"
	EffectContextSource           EffectContextKind = "EffectContextSource"
	EffectContextReferencedObject EffectContextKind = "EffectContextReferencedObject"
	EffectContextReferencedPlayer EffectContextKind = "EffectContextReferencedPlayer"
	// EffectContextReferencedObjectController marks an effect whose subject is the
	// controller of a referenced object ("Its controller creates …", "That
	// creature's controller creates …"). The recipient is the controller of the
	// object the subject reference resolves to.
	EffectContextReferencedObjectController EffectContextKind = "EffectContextReferencedObjectController"
	EffectContextPriorSubject               EffectContextKind = "EffectContextPriorSubject"
)

// DamageRecipientReferenceKind identifies a damage recipient that is the
// controller or owner of a referenced object (the prior removal target), as in
// "deals 2 damage to that land's controller" or "deals 2 damage to its owner".
// It is None for every other recipient (a target, a group, or any target).
type DamageRecipientReferenceKind uint8

// Damage recipient reference kinds.
const (
	DamageRecipientReferenceNone DamageRecipientReferenceKind = iota
	DamageRecipientReferenceController
	DamageRecipientReferenceOwner
)

// SignedAmountSyntax is one signed half of a power/toughness change.
type SignedAmountSyntax struct {
	Span     shared.Span `json:"-"`
	Value    int         `json:",omitempty"`
	Known    bool        `json:",omitempty"`
	Negative bool        `json:",omitempty"`
	// VariableX marks a side written as the variable "X" (as in "+X/+0"), whose
	// magnitude is supplied by the effect's dynamic amount rather than a fixed
	// Value. Known stays false for an X side.
	VariableX bool `json:",omitempty"`
}

// SelectionController identifies a selected object's controller.
type SelectionController string

// Selection controller relations.
const (
	SelectionControllerAny      SelectionController = ""
	SelectionControllerYou      SelectionController = "SelectionControllerYou"
	SelectionControllerOpponent SelectionController = "SelectionControllerOpponent"
	SelectionControllerNotYou   SelectionController = "SelectionControllerNotYou"
)

// SelectionKind identifies the broad object selected by a phrase.
type SelectionKind string

// Selection kinds recognized by resolving-effect grammar.
const (
	SelectionUnknown                          SelectionKind = ""
	SelectionAny                              SelectionKind = "SelectionAny"
	SelectionPlayer                           SelectionKind = "SelectionPlayer"
	SelectionOpponent                         SelectionKind = "SelectionOpponent"
	SelectionArtifact                         SelectionKind = "SelectionArtifact"
	SelectionCreature                         SelectionKind = "SelectionCreature"
	SelectionEnchantment                      SelectionKind = "SelectionEnchantment"
	SelectionLand                             SelectionKind = "SelectionLand"
	SelectionPermanent                        SelectionKind = "SelectionPermanent"
	SelectionCard                             SelectionKind = "SelectionCard"
	SelectionSpell                            SelectionKind = "SelectionSpell"
	SelectionActivatedAbility                 SelectionKind = "SelectionActivatedAbility"
	SelectionTriggeredAbility                 SelectionKind = "SelectionTriggeredAbility"
	SelectionActivatedOrTriggeredAbility      SelectionKind = "SelectionActivatedOrTriggeredAbility"
	SelectionSpellActivatedOrTriggeredAbility SelectionKind = "SelectionSpellActivatedOrTriggeredAbility"
	SelectionTriggeredAbilityOrSpell          SelectionKind = "SelectionTriggeredAbilityOrSpell"
	SelectionPlaneswalker                     SelectionKind = "SelectionPlaneswalker"
	SelectionBattle                           SelectionKind = "SelectionBattle"
)

// SelectionSyntax is a typed, source-spanned noun phrase.
type SelectionSyntax struct {
	Span         shared.Span         `json:"-"`
	Text         string              `json:",omitempty"`
	Kind         SelectionKind       `json:",omitempty"`
	Controller   SelectionController `json:",omitempty"`
	All          bool                `json:",omitempty"`
	Another      bool                `json:",omitempty"`
	Other        bool                `json:",omitempty"`
	Attacking    bool                `json:",omitempty"`
	Blocking     bool                `json:",omitempty"`
	Tapped       bool                `json:",omitempty"`
	Untapped     bool                `json:",omitempty"`
	Colorless    bool                `json:",omitempty"`
	Multicolored bool                `json:",omitempty"`
	// PlayerOrPlaneswalker marks the combined "player or planeswalker" /
	// "opponent or planeswalker" combined damage target. Kind stays
	// SelectionPlayer or SelectionOpponent for the player half; this flag records
	// the additional planeswalker-permanent half the merged Kind cannot express.
	PlayerOrPlaneswalker bool `json:",omitempty"`
	// MatchManaValue, MatchPower, and MatchToughness record whether their paired
	// ManaValue/Power/Toughness comparison below is active. They are grouped with
	// the other booleans to keep the struct compact.
	MatchManaValue bool        `json:",omitempty"`
	MatchPower     bool        `json:",omitempty"`
	MatchToughness bool        `json:",omitempty"`
	Keyword        KeywordKind `json:",omitempty"`
	// ExcludedKeyword records a "without <keyword>" selector qualifier (e.g.
	// "each creature without flying"); it is mutually exclusive with Keyword.
	ExcludedKeyword    KeywordKind `json:",omitempty"`
	Zone               zone.Type   `json:",omitempty"`
	RequiredTypesAny   []CardType  `json:",omitempty"`
	ExcludedTypes      []CardType  `json:",omitempty"`
	SourceTypes        []CardType  `json:",omitempty"`
	Supertypes         []Supertype `json:",omitempty"`
	ExcludedSupertypes []Supertype `json:",omitempty"`
	ColorsAny          []Color     `json:",omitempty"`
	ExcludedColors     []Color     `json:",omitempty"`
	SubtypesAny        []types.Sub `json:",omitempty"`
	ManaValue          compare.Int `json:",omitzero"`
	Power              compare.Int `json:",omitzero"`
	Toughness          compare.Int `json:",omitzero"`
}

// TargetCardinalitySyntax is an inclusive target-count range.
type TargetCardinalitySyntax struct {
	Min int `json:",omitempty"`
	Max int `json:",omitempty"`
}

// TargetSyntax is one typed target production.
type TargetSyntax struct {
	Span        shared.Span             `json:"-"`
	Text        string                  `json:",omitempty"`
	Cardinality TargetCardinalitySyntax `json:",omitzero"`
	Selection   SelectionSyntax         `json:",omitzero"`
	Exact       bool                    `json:",omitempty"`
	// Order is the target's dense source-order rank, used downstream to bind
	// references to their closest preceding target without byte offsets.
	Order shared.SourceOrder `json:"-"`
}

// EffectConnectionKind identifies how a resolving instruction is coordinated
// with the preceding instruction in the same sentence.
type EffectConnectionKind string

// Resolving-instruction connections recognized by the parser.
const (
	EffectConnectionNone EffectConnectionKind = ""
	EffectConnectionAnd  EffectConnectionKind = "EffectConnectionAnd"
	EffectConnectionThen EffectConnectionKind = "EffectConnectionThen"
)

// EffectSyntax is one typed resolving instruction. Text and Tokens remain
// lossless metadata; all meaning consumed downstream is carried by typed fields.
type EffectSyntax struct {
	Kind           EffectKind           `json:",omitempty"`
	Context        EffectContextKind    `json:",omitempty"`
	Connection     EffectConnectionKind `json:",omitempty"`
	ConnectionSpan shared.Span          `json:"-"`
	Span           shared.Span          `json:"-"`
	VerbSpan       shared.Span          `json:"-"`
	ClauseSpan     shared.Span          `json:"-"`
	Text           string               `json:",omitempty"`
	Tokens         []shared.Token       `json:"-"`
	Duration       EffectDurationKind   `json:",omitempty"`
	DelayedTiming  DelayedTimingKind    `json:",omitempty"`
	Selection      SelectionSyntax      `json:",omitzero"`
	// DamageRecipientPair holds the two recipient groups of a dual-recipient
	// fixed group-damage effect ("deals N damage to each X and each Y"). It is
	// populated only when the recipient is exactly two "each <group>" phrases
	// joined by "and"; it is empty for every other recipient. The single
	// merged Selection cannot represent two distinct groups, so lowering emits
	// one damage instruction per recipient in Oracle order instead.
	DamageRecipientPair []SelectionSyntax `json:",omitempty"`
	// DamageRecipientReference marks a damage recipient that is the controller or
	// owner of a referenced object (the prior removal target), as in "deals 2
	// damage to that land's controller". It is None for every other recipient.
	DamageRecipientReference DamageRecipientReferenceKind `json:",omitempty"`
	Amount                   EffectAmountSyntax           `json:",omitzero"`
	PowerDelta               SignedAmountSyntax           `json:",omitzero"`
	ToughnessDelta           SignedAmountSyntax           `json:",omitzero"`
	// TokenPower/TokenToughness/TokenPTKnown hold a created token's fixed
	// power/toughness (e.g. "1/1"). Known is false for tokens with no printed
	// power/toughness (named artifact tokens like Treasure).
	TokenPower     int  `json:",omitempty"`
	TokenToughness int  `json:",omitempty"`
	TokenPTKnown   bool `json:",omitempty"`
	// TokenCopyOfTarget reports that the created token is a copy of the effect's
	// single target object ("Create a token that's a copy of target creature you
	// control."). The copy source is the effect's lone target, captured in
	// Targets; the token has no printed power/toughness of its own.
	TokenCopyOfTarget  bool                      `json:",omitempty"`
	StaticSubject      EffectStaticSubjectSyntax `json:",omitzero"`
	CounterKind        counter.Kind              `json:",omitempty"`
	CounterKnown       bool                      `json:",omitempty"`
	FromZone           zone.Type                 `json:",omitempty"`
	ToZone             zone.Type                 `json:",omitempty"`
	Destination        EffectDestinationPosition `json:",omitempty"`
	EntersTapped       bool                      `json:",omitempty"`
	EntersTappedSelf   bool                      `json:",omitempty"`
	EntersWithCounters bool                      `json:",omitempty"`
	// EntersColorChoice reports a self entry replacement of the form "As this
	// <permanent> enters, choose a color." or "... choose a color other than
	// <color>." The enters verb is shared by several entry constructs, so this is
	// set only for those exact color-choice clauses (not a non-color choice).
	EntersColorChoice bool `json:",omitempty"`
	// EntersColorChoiceExclude is the single forbidden basic color of an "As this
	// <permanent> enters, choose a color other than <color>." clause (the
	// Gate/Thriving land cycle). It is empty for the unconstrained "choose a
	// color." form.
	EntersColorChoiceExclude mana.Color `json:",omitempty"`
	// EntersTypeChoice reports a self entry replacement of the form "As this
	// <permanent> enters, choose a creature type." The enters verb is shared by
	// several entry constructs, so this is set only for that exact clause.
	EntersTypeChoice bool `json:",omitempty"`
	UnderYourControl bool `json:",omitempty"`
	CastAsAdventure  bool `json:",omitempty"`
	Negated          bool `json:",omitempty"`
	Optional         bool `json:",omitempty"`
	// Divided reports a "deals N damage divided as you choose among <targets>"
	// effect: a fixed total split among the chosen targets, at least one each.
	Divided                 bool                    `json:",omitempty"`
	OptionalSpan            shared.Span             `json:"-"`
	Symbol                  string                  `json:",omitempty"`
	Mana                    EffectManaSyntax        `json:",omitzero"`
	Replacement             EffectReplacementSyntax `json:",omitzero"`
	References              []Reference             `json:",omitempty"`
	SubjectReferences       []Reference             `json:",omitempty"`
	Targets                 []TargetSyntax          `json:",omitempty"`
	SubjectTargets          []TargetSyntax          `json:",omitempty"`
	Payment                 EffectPaymentSyntax     `json:",omitzero"`
	Exact                   bool                    `json:",omitempty"`
	RequiresOrderedLowering bool                    `json:",omitempty"`
	HasUnrecognizedSibling  bool                    `json:",omitempty"`
	UnsupportedDetail       string                  `json:",omitempty"`
	// Order is the effect's dense source-order rank (of Span); VerbOrder is the
	// rank of VerbSpan. Downstream stages compare these ranks to order effects
	// and bind references to effect verbs without inspecting byte offsets.
	Order     shared.SourceOrder `json:"-"`
	VerbOrder shared.SourceOrder `json:"-"`
	// LifeObject reports that a gain/lose effect's grammatical object is the
	// player's life (e.g. "gain 3 life", "loses that much life"), as opposed to
	// a keyword or quoted ability ("gains shadow", "loses protection from
	// black"). It lets consumers route only true life changes to the life
	// lowerer rather than misclassifying keyword/ability grants and losses.
	LifeObject bool `json:",omitempty"`
	// PreventRegeneration reports that a destroy effect is followed by a
	// regeneration rider ("It/They can't be regenerated."). The rider is a
	// separate zero-effect sentence whose pronoun refers to the destroyed
	// permanents; the parser folds it onto the destroy effect so lowering
	// emits a destruction that bypasses regeneration shields.
	PreventRegeneration bool `json:",omitempty"`
	// RegenerationRiderSpan covers the rider sentence's semantic tokens so the
	// lowerer can credit them toward source coverage. It is set only when
	// PreventRegeneration is true.
	RegenerationRiderSpan shared.Span `json:"-"`
}

// EffectPaymentPayerKind identifies who may pay a cost embedded in an effect.
type EffectPaymentPayerKind string

// Embedded-effect payers recognized by the parser.
const (
	EffectPaymentPayerUnknown          EffectPaymentPayerKind = ""
	EffectPaymentPayerTargetController EffectPaymentPayerKind = "EffectPaymentPayerTargetController"
)

// EffectPaymentSyntax is a source-spanned typed resolution payment.
type EffectPaymentSyntax struct {
	Span     shared.Span            `json:"-"`
	Payer    EffectPaymentPayerKind `json:",omitempty"`
	ManaCost cost.Mana              `json:",omitempty"`
	// Order is the payment's dense source-order rank, used downstream to test
	// condition containment without byte offsets.
	Order shared.SourceOrder `json:"-"`
}

// EffectStaticSubjectKind identifies the group affected by a static resolving
// effect production.
type EffectStaticSubjectKind string

// Static effect subjects recognized by resolving-effect grammar.
const (
	EffectStaticSubjectNone                           EffectStaticSubjectKind = ""
	EffectStaticSubjectAttachedObject                 EffectStaticSubjectKind = "EffectStaticSubjectAttachedObject"
	EffectStaticSubjectAllCreatures                   EffectStaticSubjectKind = "EffectStaticSubjectAllCreatures"
	EffectStaticSubjectAllOtherCreatures              EffectStaticSubjectKind = "EffectStaticSubjectAllOtherCreatures"
	EffectStaticSubjectAttackingCreatures             EffectStaticSubjectKind = "EffectStaticSubjectAttackingCreatures"
	EffectStaticSubjectBlockingCreatures              EffectStaticSubjectKind = "EffectStaticSubjectBlockingCreatures"
	EffectStaticSubjectControlledCreatures            EffectStaticSubjectKind = "EffectStaticSubjectControlledCreatures"
	EffectStaticSubjectOtherControlledCreatures       EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledCreatures"
	EffectStaticSubjectControlledWalls                EffectStaticSubjectKind = "EffectStaticSubjectControlledWalls"
	EffectStaticSubjectControlledArtifacts            EffectStaticSubjectKind = "EffectStaticSubjectControlledArtifacts"
	EffectStaticSubjectControlledTokens               EffectStaticSubjectKind = "EffectStaticSubjectControlledTokens"
	EffectStaticSubjectOpponentControlledCreatures    EffectStaticSubjectKind = "EffectStaticSubjectOpponentControlledCreatures"
	EffectStaticSubjectControlledCreatureSubtype      EffectStaticSubjectKind = "EffectStaticSubjectControlledCreatureSubtype"
	EffectStaticSubjectOtherControlledCreatureSubtype EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledCreatureSubtype"
	EffectStaticSubjectAllCreatureSubtype             EffectStaticSubjectKind = "EffectStaticSubjectAllCreatureSubtype"
	EffectStaticSubjectOtherCreatureSubtype           EffectStaticSubjectKind = "EffectStaticSubjectOtherCreatureSubtype"
	EffectStaticSubjectControlledAttackingCreatures   EffectStaticSubjectKind = "EffectStaticSubjectControlledAttackingCreatures"
	EffectStaticSubjectControlledCreatureTokens       EffectStaticSubjectKind = "EffectStaticSubjectControlledCreatureTokens"
	EffectStaticSubjectBattlefieldCreatureTokens      EffectStaticSubjectKind = "EffectStaticSubjectBattlefieldCreatureTokens"
	EffectStaticSubjectControlledLegendaryCreatures   EffectStaticSubjectKind = "EffectStaticSubjectControlledLegendaryCreatures"
	EffectStaticSubjectControlledUntappedCreatures    EffectStaticSubjectKind = "EffectStaticSubjectControlledUntappedCreatures"
	EffectStaticSubjectOtherControlledTappedCreatures EffectStaticSubjectKind = "EffectStaticSubjectOtherControlledTappedCreatures"
)

// EffectStaticSubjectSyntax is a source-spanned typed static-effect subject.
type EffectStaticSubjectSyntax struct {
	Kind         EffectStaticSubjectKind `json:",omitempty"`
	Span         shared.Span             `json:"-"`
	Subtype      types.Sub               `json:",omitempty"`
	SubtypeText  string                  `json:",omitempty"`
	SubtypeKnown bool                    `json:",omitempty"`

	// Colors, Colorless, and Multicolored carry an optional color filter
	// constraining the affected creature group ("Other red creatures you
	// control ..."). Colors lists single-color words matched disjunctively;
	// Colorless and Multicolored are the color-family qualifiers. They are
	// mutually exclusive shapes downstream maps onto a Selection color filter.
	Colors       []Color `json:",omitempty"`
	Colorless    bool    `json:",omitempty"`
	Multicolored bool    `json:",omitempty"`
}
