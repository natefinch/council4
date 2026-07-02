package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// ConditionIntroKind identifies the grammatical introducer that opens a typed
// condition clause.
type ConditionIntroKind string

// Condition introducers recognized by the parser.
const (
	ConditionIntroUnknown  ConditionIntroKind = ""
	ConditionIntroIf       ConditionIntroKind = "ConditionIntroIf"
	ConditionIntroUnless   ConditionIntroKind = "ConditionIntroUnless"
	ConditionIntroOnlyIf   ConditionIntroKind = "ConditionIntroOnlyIf"
	ConditionIntroAsLongAs ConditionIntroKind = "ConditionIntroAsLongAs"
)

// ConditionPredicateKind identifies the closed predicate family recognized in a
// condition clause. The compiler maps these mechanically onto its semantic
// predicate vocabulary.
type ConditionPredicateKind string

// Condition predicates recognized by the parser.
const (
	ConditionPredicateUnknown                                          ConditionPredicateKind = ""
	ConditionPredicateControllerLifeAtLeast                            ConditionPredicateKind = "ConditionPredicateControllerLifeAtLeast"
	ConditionPredicateControllerLifeAtMost                             ConditionPredicateKind = "ConditionPredicateControllerLifeAtMost"
	ConditionPredicateControllerLifeAtLeastAboveStarting               ConditionPredicateKind = "ConditionPredicateControllerLifeAtLeastAboveStarting"
	ConditionPredicateControllerHandSizeAtLeast                        ConditionPredicateKind = "ConditionPredicateControllerHandSizeAtLeast"
	ConditionPredicateControllerHandEmpty                              ConditionPredicateKind = "ConditionPredicateControllerHandEmpty"
	ConditionPredicateAnyPlayerLifeAtMost                              ConditionPredicateKind = "ConditionPredicateAnyPlayerLifeAtMost"
	ConditionPredicateOpponentCountAtLeast                             ConditionPredicateKind = "ConditionPredicateOpponentCountAtLeast"
	ConditionPredicateControls                                         ConditionPredicateKind = "ConditionPredicateControls"
	ConditionPredicateGraveyardCardCountAtLeast                        ConditionPredicateKind = "ConditionPredicateGraveyardCardCountAtLeast"
	ConditionPredicateGraveyardCardTypeCountAtLeast                    ConditionPredicateKind = "ConditionPredicateGraveyardCardTypeCountAtLeast"
	ConditionPredicateCreaturePowerDiversityAtLeast                    ConditionPredicateKind = "ConditionPredicateCreaturePowerDiversityAtLeast"
	ConditionPredicateEventSubjectWasKicked                            ConditionPredicateKind = "ConditionPredicateEventSubjectWasKicked"
	ConditionPredicateEventSubjectWasCast                              ConditionPredicateKind = "ConditionPredicateEventSubjectWasCast"
	ConditionPredicateEventSubjectWasCastByController                  ConditionPredicateKind = "ConditionPredicateEventSubjectWasCastByController"
	ConditionPredicateEventSubjectWasCastFromControllerHand            ConditionPredicateKind = "ConditionPredicateEventSubjectWasCastFromControllerHand"
	ConditionPredicateEventSubjectEnteredOrCastFromGraveyard           ConditionPredicateKind = "ConditionPredicateEventSubjectEnteredOrCastFromGraveyard"
	ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard ConditionPredicateKind = "ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard"
	ConditionPredicateEventSubjectHadNoCounter                         ConditionPredicateKind = "ConditionPredicateEventSubjectHadNoCounter"
	ConditionPredicateEventSubjectHadCounter                           ConditionPredicateKind = "ConditionPredicateEventSubjectHadCounter"
	ConditionPredicateEventSubjectHadCounters                          ConditionPredicateKind = "ConditionPredicateEventSubjectHadCounters"
	ConditionPredicatePriorInstructionNotAccepted                      ConditionPredicateKind = "ConditionPredicatePriorInstructionNotAccepted"
	ConditionPredicatePriorInstructionAccepted                         ConditionPredicateKind = "ConditionPredicatePriorInstructionAccepted"
	ConditionPredicateDestroyedThisWay                                 ConditionPredicateKind = "ConditionPredicateDestroyedThisWay"
	ConditionPredicateEventPlayerDoesNotPay                            ConditionPredicateKind = "ConditionPredicateEventPlayerDoesNotPay"
	ConditionPredicateCounterPlacementOnControlledCreature             ConditionPredicateKind = "ConditionPredicateCounterPlacementOnControlledCreature"
	ConditionPredicateCounterPlacementOnSelf                           ConditionPredicateKind = "ConditionPredicateCounterPlacementOnSelf"
	ConditionPredicateControllerCounterPlacement                       ConditionPredicateKind = "ConditionPredicateControllerCounterPlacement"
	ConditionPredicateCounterPlacementOnControlledPermanent            ConditionPredicateKind = "ConditionPredicateCounterPlacementOnControlledPermanent"
	ConditionPredicateDamageByControlledSource                         ConditionPredicateKind = "ConditionPredicateDamageByControlledSource"
	ConditionPredicateTokenCreationUnderController                     ConditionPredicateKind = "ConditionPredicateTokenCreationUnderController"
	ConditionPredicateSourceWouldDie                                   ConditionPredicateKind = "ConditionPredicateSourceWouldDie"
	ConditionPredicateSourceWouldGoToGraveyard                         ConditionPredicateKind = "ConditionPredicateSourceWouldGoToGraveyard"
	ConditionPredicateObjectMatches                                    ConditionPredicateKind = "ConditionPredicateObjectMatches"
	ConditionPredicateObjectExists                                     ConditionPredicateKind = "ConditionPredicateObjectExists"
	ConditionPredicateAnyOpponentPoisonAtLeast                         ConditionPredicateKind = "ConditionPredicateAnyOpponentPoisonAtLeast"
	ConditionPredicateControllerHandSizeExactly                        ConditionPredicateKind = "ConditionPredicateControllerHandSizeExactly"
	ConditionPredicateCreatedTokenThisTurn                             ConditionPredicateKind = "ConditionPredicateCreatedTokenThisTurn"
	ConditionPredicateControllerWouldCreateNamedToken                  ConditionPredicateKind = "ConditionPredicateControllerWouldCreateNamedToken"
	ConditionPredicateControlComparison                                ConditionPredicateKind = "ConditionPredicateControlComparison"
	ConditionPredicateEventSubjectNameUnique                           ConditionPredicateKind = "ConditionPredicateEventSubjectNameUnique"
	ConditionPredicateTargetColor                                      ConditionPredicateKind = "ConditionPredicateTargetColor"
	ConditionPredicateWouldDrawFromEmptyLibrary                        ConditionPredicateKind = "ConditionPredicateWouldDrawFromEmptyLibrary"
	ConditionPredicateCastDuringControllerMainPhase                    ConditionPredicateKind = "ConditionPredicateCastDuringControllerMainPhase"
	ConditionPredicateWouldDrawCard                                    ConditionPredicateKind = "ConditionPredicateWouldDrawCard"
	ConditionPredicateWouldDrawCardExceptFirstInDrawStep               ConditionPredicateKind = "ConditionPredicateWouldDrawCardExceptFirstInDrawStep"
	ConditionPredicateCardWouldGoToGraveyard                           ConditionPredicateKind = "ConditionPredicateCardWouldGoToGraveyard"
	ConditionPredicateControllerLifeGain                               ConditionPredicateKind = "ConditionPredicateControllerLifeGain"
	ConditionPredicateOpponentLifeLossDuringControllerTurn             ConditionPredicateKind = "ConditionPredicateOpponentLifeLossDuringControllerTurn"
	ConditionPredicateOpponentLifeLoss                                 ConditionPredicateKind = "ConditionPredicateOpponentLifeLoss"
	ConditionPredicateAnyPlayerLifeLoss                                ConditionPredicateKind = "ConditionPredicateAnyPlayerLifeLoss"
	ConditionPredicateTokenCreationAnyController                       ConditionPredicateKind = "ConditionPredicateTokenCreationAnyController"
	ConditionPredicateCounterPlacementOnAnyCreature                    ConditionPredicateKind = "ConditionPredicateCounterPlacementOnAnyCreature"
	ConditionPredicateSourceTributeNotPaid                             ConditionPredicateKind = "ConditionPredicateSourceTributeNotPaid"
	ConditionPredicateControllerControlsCommander                      ConditionPredicateKind = "ConditionPredicateControllerControlsCommander"
	ConditionPredicateSpellWasKicked                                   ConditionPredicateKind = "ConditionPredicateSpellWasKicked"
	ConditionPredicateSpellWasCastFromGraveyard                        ConditionPredicateKind = "ConditionPredicateSpellWasCastFromGraveyard"
	ConditionPredicateSourceSaddled                                    ConditionPredicateKind = "ConditionPredicateSourceSaddled"
	ConditionPredicateSourceNotSaddled                                 ConditionPredicateKind = "ConditionPredicateSourceNotSaddled"
	ConditionPredicateAttackersAttackingControllerAtLeast              ConditionPredicateKind = "ConditionPredicateAttackersAttackingControllerAtLeast"
	ConditionPredicateControllerLibrarySizeAtLeast                     ConditionPredicateKind = "ConditionPredicateControllerLibrarySizeAtLeast"
	ConditionPredicateControllerLifeExactly                            ConditionPredicateKind = "ConditionPredicateControllerLifeExactly"
	ConditionPredicateControllerGainedLifeThisTurnAtLeast              ConditionPredicateKind = "ConditionPredicateControllerGainedLifeThisTurnAtLeast"
	ConditionPredicateSpellXAtLeast                                    ConditionPredicateKind = "ConditionPredicateSpellXAtLeast"
	ConditionPredicateGraveyardCardOfTypeCountAtLeast                  ConditionPredicateKind = "ConditionPredicateGraveyardCardOfTypeCountAtLeast"
	ConditionPredicateControllerControlsNamed                          ConditionPredicateKind = "ConditionPredicateControllerControlsNamed"
	ConditionPredicateFirstCombatPhaseOfTurn                           ConditionPredicateKind = "ConditionPredicateFirstCombatPhaseOfTurn"
	ConditionPredicateControlsGreatestPowerCreature                    ConditionPredicateKind = "ConditionPredicateControlsGreatestPowerCreature"
	ConditionPredicateControlsGreatestToughnessCreature                ConditionPredicateKind = "ConditionPredicateControlsGreatestToughnessCreature"
	ConditionPredicateSubjectSharesCreatureTypeWithSource              ConditionPredicateKind = "ConditionPredicateSubjectSharesCreatureTypeWithSource"
	ConditionPredicateControllerIsMonarch                              ConditionPredicateKind = "ConditionPredicateControllerIsMonarch"
	ConditionPredicateControllerHasInitiative                          ConditionPredicateKind = "ConditionPredicateControllerHasInitiative"
	ConditionPredicateControllerHasCityBlessing                        ConditionPredicateKind = "ConditionPredicateControllerHasCityBlessing"
	ConditionPredicateControllerTurn                                   ConditionPredicateKind = "ConditionPredicateControllerTurn"
	ConditionPredicateColoredManaSpentToCastAtLeast                    ConditionPredicateKind = "ConditionPredicateColoredManaSpentToCastAtLeast"
	ConditionPredicateSameColorManaSpentToCastAtLeast                  ConditionPredicateKind = "ConditionPredicateSameColorManaSpentToCastAtLeast"
	ConditionPredicateGraveyardPermanentCardCountAtLeast               ConditionPredicateKind = "ConditionPredicateGraveyardPermanentCardCountAtLeast"
	ConditionPredicateGraveyardManaValueCountAtLeast                   ConditionPredicateKind = "ConditionPredicateGraveyardManaValueCountAtLeast"
	ConditionPredicateAnyOpponentGraveyardCardCountAtLeast             ConditionPredicateKind = "ConditionPredicateAnyOpponentGraveyardCardCountAtLeast"
	ConditionPredicateEventSpellManaSpentToCastAtLeast                 ConditionPredicateKind = "ConditionPredicateEventSpellManaSpentToCastAtLeast"
	ConditionPredicateEventSpellNoManaSpentToCast                      ConditionPredicateKind = "ConditionPredicateEventSpellNoManaSpentToCast"
	ConditionPredicateTriggeringPlayerHandSizeAtMost                   ConditionPredicateKind = "ConditionPredicateTriggeringPlayerHandSizeAtMost"
	ConditionPredicateTriggeringPlayerHandSizeAtLeast                  ConditionPredicateKind = "ConditionPredicateTriggeringPlayerHandSizeAtLeast"
	ConditionPredicateLandEnteredThisTurnOrControlsBasic               ConditionPredicateKind = "ConditionPredicateLandEnteredThisTurnOrControlsBasic"
)

// GraveyardRedirectScope identifies whose graveyard a card-to-graveyard
// replacement watches ("a graveyard" = any player, "your graveyard" = the
// controller, "an opponent's graveyard" = an opponent).
type GraveyardRedirectScope string

// Graveyard redirect scopes recognized by the parser.
const (
	GraveyardRedirectScopeAny      GraveyardRedirectScope = ""
	GraveyardRedirectScopeYou      GraveyardRedirectScope = "GraveyardRedirectScopeYou"
	GraveyardRedirectScopeOpponent GraveyardRedirectScope = "GraveyardRedirectScopeOpponent"
)

// GraveyardRedirectControlScope identifies, for a "would die" graveyard-redirect
// replacement, whose control of the dying permanent the replacement watches
// ("a creature" = any controller, "a creature you control" = the controller, "a
// creature an opponent controls" = an opponent). It is distinct from
// GraveyardRedirectScope, which watches the moving card's owner.
type GraveyardRedirectControlScope string

// Graveyard redirect control scopes recognized by the parser.
const (
	GraveyardRedirectControlScopeAny      GraveyardRedirectControlScope = ""
	GraveyardRedirectControlScopeYou      GraveyardRedirectControlScope = "GraveyardRedirectControlScopeYou"
	GraveyardRedirectControlScopeOpponent GraveyardRedirectControlScope = "GraveyardRedirectControlScopeOpponent"
)

// ConditionControlScope identifies which players' battlefields a "controls"
// predicate counts.
type ConditionControlScope string

// Control scopes recognized by the parser.
const (
	ConditionControlScopeController   ConditionControlScope = ""
	ConditionControlScopeAnyOpponent  ConditionControlScope = "ConditionControlScopeAnyOpponent"
	ConditionControlScopeOpponents    ConditionControlScope = "ConditionControlScopeOpponents"
	ConditionControlScopeEachOpponent ConditionControlScope = "ConditionControlScopeEachOpponent"
	// ConditionControlScopeTriggeringPlayer counts permanents controlled by the
	// player tied to the triggering event ("that player", referring to the
	// controller of the permanent whose entry triggered the ability, as in
	// Archaeomancer's Map: "if that player controls more lands than you").
	ConditionControlScopeTriggeringPlayer ConditionControlScope = "ConditionControlScopeTriggeringPlayer"
	// ConditionControlScopeDefendingPlayer counts permanents controlled by the
	// defending player of an attack ("defending player controls an Island", Sea
	// Monster). It is only meaningful as the guard on a can't-attack static rule,
	// where the defending player is resolved per attack; every other use fails
	// closed downstream.
	ConditionControlScopeDefendingPlayer ConditionControlScope = "ConditionControlScopeDefendingPlayer"
)

// ConditionComparison identifies the numeric comparison a count predicate uses.
type ConditionComparison string

// Condition comparisons recognized by the parser. ConditionComparisonNone marks
// a singular "a"/"an"/"another" selection with no explicit count.
const (
	ConditionComparisonNone    ConditionComparison = ""
	ConditionComparisonAtLeast ConditionComparison = "ConditionComparisonAtLeast"
	ConditionComparisonAtMost  ConditionComparison = "ConditionComparisonAtMost"
)

// ConditionTappedState is a typed tapped-state selection filter.
type ConditionTappedState string

// Tapped-state filters recognized by the parser.
const (
	ConditionTappedAny   ConditionTappedState = ""
	ConditionTappedTrue  ConditionTappedState = "ConditionTappedTrue"
	ConditionTappedFalse ConditionTappedState = "ConditionTappedFalse"
)

// ConditionCombatState is a typed combat-involvement selection filter.
type ConditionCombatState string

// Combat-state filters recognized by the parser.
const (
	ConditionCombatAny                 ConditionCombatState = ""
	ConditionCombatAttacking           ConditionCombatState = "ConditionCombatAttacking"
	ConditionCombatBlocking            ConditionCombatState = "ConditionCombatBlocking"
	ConditionCombatAttackingOrBlocking ConditionCombatState = "ConditionCombatAttackingOrBlocking"
)

// ConditionAttachmentState is a typed attachment-state selection filter testing
// whether the matched permanent has an Aura ("enchanted") or Equipment
// ("equipped") attached to it.
type ConditionAttachmentState string

// Attachment-state filters recognized by the parser.
const (
	ConditionAttachmentAny       ConditionAttachmentState = ""
	ConditionAttachmentEnchanted ConditionAttachmentState = "ConditionAttachmentEnchanted"
	ConditionAttachmentEquipped  ConditionAttachmentState = "ConditionAttachmentEquipped"
)

// ConditionSupertype identifies a supertype in a typed condition selection.
type ConditionSupertype string

// Condition supertypes recognized by the parser.
const (
	ConditionSupertypeUnknown   ConditionSupertype = ""
	ConditionSupertypeBasic     ConditionSupertype = "ConditionSupertypeBasic"
	ConditionSupertypeSnow      ConditionSupertype = "ConditionSupertypeSnow"
	ConditionSupertypeLegendary ConditionSupertype = "ConditionSupertypeLegendary"
)

// ConditionCounterKind identifies a counter mentioned by a condition clause.
type ConditionCounterKind string

// Condition counter kinds recognized by the parser.
const (
	ConditionCounterNone             ConditionCounterKind = ""
	ConditionCounterPlusOnePlusOne   ConditionCounterKind = "ConditionCounterPlusOnePlusOne"
	ConditionCounterMinusOneMinusOne ConditionCounterKind = "ConditionCounterMinusOneMinusOne"
)

// ConditionObjectBinding identifies the object a state predicate inspects.
type ConditionObjectBinding string

// Object bindings recognized by the parser.
const (
	ConditionObjectBindingNone           ConditionObjectBinding = ""
	ConditionObjectBindingSource         ConditionObjectBinding = "ConditionObjectBindingSource"
	ConditionObjectBindingEventPermanent ConditionObjectBinding = "ConditionObjectBindingEventPermanent"
	ConditionObjectBindingSourceAttached ConditionObjectBinding = "ConditionObjectBindingSourceAttached"
	ConditionObjectBindingCreatedToken   ConditionObjectBinding = "ConditionObjectBindingCreatedToken"
	// ConditionObjectBindingTarget binds the condition's object to the spell or
	// activated ability's first target rather than to a triggering event
	// permanent. Used by "if it's a <subtype>" and "if it's legendary" riders
	// that gate a single-target effect clause.
	ConditionObjectBindingTarget ConditionObjectBinding = "ConditionObjectBindingTarget"
)

// ConditionSelection is the source-independent permanent selection used by typed
// condition clauses. Subtype names are canonical typed identities.
type ConditionSelection struct {
	RequiredTypes []TriggerCardType    `json:",omitempty"`
	Supertypes    []ConditionSupertype `json:",omitempty"`
	SubtypesAny   []types.Sub          `json:",omitempty"`
	ColorsAny     []TriggerColor       `json:",omitempty"`
	Colorless     bool                 `json:",omitempty"`
	Multicolored  bool                 `json:",omitempty"`
	TokenOnly     bool                 `json:",omitempty"`
	ExcludeSource bool                 `json:",omitempty"`
	Tapped        ConditionTappedState `json:",omitempty"`
	CombatState   ConditionCombatState `json:",omitempty"`
	// Attachment tests whether the matched permanent has an Aura ("enchanted")
	// or Equipment ("equipped") attached to it ("as long as this creature is
	// equipped"). Its zero value imposes no attachment requirement.
	Attachment        ConditionAttachmentState `json:",omitempty"`
	Keyword           KeywordKind              `json:",omitempty"`
	PowerAtLeast      int                      `json:",omitempty"`
	MatchPowerAtLeast bool                     `json:",omitempty"`
	// TotalPowerAtLeast is the collective-power threshold for a "have total
	// power <n> or greater" qualifier, applied to the selected permanents as a
	// group rather than to each permanent individually. MatchTotalPowerAtLeast
	// marks the threshold present so a zero threshold remains expressible.
	TotalPowerAtLeast      int  `json:",omitempty"`
	MatchTotalPowerAtLeast bool `json:",omitempty"`
	// DistinctNamesAtLeast is the distinct-name threshold for a "with different
	// names" qualifier, counting how many of the selected permanents have
	// distinct names rather than the raw permanent total. MatchDistinctNamesAtLeast
	// marks the threshold present so a zero threshold remains expressible.
	DistinctNamesAtLeast      int  `json:",omitempty"`
	MatchDistinctNamesAtLeast bool `json:",omitempty"`

	// DamageRecipientOpponent, DamageNoncombatOnly, and DamageSourceAnyController
	// qualify a ConditionPredicateDamageByControlledSource clause (CR 614).
	// DamageRecipientOpponent restricts the replacement to damage dealt to an
	// opponent or a permanent an opponent controls; its zero value matches any
	// recipient ("a permanent or player"). DamageNoncombatOnly restricts it to
	// noncombat damage. DamageSourceAnyController marks a source carrying no "you
	// control" qualifier, so the replaced damage may come from any player's
	// source.
	DamageRecipientOpponent   bool `json:",omitempty"`
	DamageNoncombatOnly       bool `json:",omitempty"`
	DamageSourceAnyController bool `json:",omitempty"`

	// DamageRecipientController qualifies a damage-source clause whose recipient
	// is the source permanent's controller alone ("would deal damage to you",
	// Sphere of Law, Urza's Armor). It backs the continuous static
	// damage-prevention statics and is mutually exclusive with
	// DamageRecipientOpponent. DamageSourceControllerOpponent marks a source the
	// clause restricts to one controlled by an opponent ("a source an opponent
	// controls", Protection of the Hekma); its zero value with
	// DamageSourceAnyController matches a source under any player's control.
	DamageRecipientController      bool `json:",omitempty"`
	DamageSourceControllerOpponent bool `json:",omitempty"`

	// AnyCounter requires the matched permanent to carry at least one counter of
	// any kind ("if this permanent has counters on it"). It is the kind-agnostic
	// companion to a named-counter requirement.
	AnyCounter bool `json:",omitempty"`

	// CounterKind, CounterKindKnown, and CounterCountAtLeast express a
	// named-counter-count threshold the matched permanent must satisfy ("has
	// seven or more quest counters on it"). CounterKindKnown marks the kind
	// present; CounterCountAtLeast carries the minimum count. The compiler maps
	// these onto the runtime counter-count selection predicate text-blind.
	CounterKind         counter.Kind `json:",omitempty"`
	CounterKindKnown    bool         `json:",omitempty"`
	CounterCountAtLeast int          `json:",omitempty"`
}

// ConditionClause is composable typed syntax for a supported condition. The
// parser owns the Oracle vocabulary, normalization, and grammar; the compiler
// maps these typed fields mechanically without inspecting source text.
type ConditionClause struct {
	Span      shared.Span            `json:"-"`
	Intro     ConditionIntroKind     `json:",omitempty"`
	Predicate ConditionPredicateKind `json:",omitempty"`

	// Scope, Comparison, and CompareValue describe a "controls" predicate. For
	// other predicates Threshold carries the literal numeric parameter.
	Scope        ConditionControlScope `json:",omitempty"`
	Comparison   ConditionComparison   `json:",omitempty"`
	CompareValue int                   `json:",omitempty"`
	Threshold    int                   `json:",omitempty"`

	Selection     ConditionSelection     `json:",omitzero"`
	Counter       ConditionCounterKind   `json:",omitempty"`
	ObjectBinding ConditionObjectBinding `json:",omitempty"`

	// SubjectSpan is set for source-death predicates so the compiler can confirm
	// the subject binds the source via a typed reference.
	SubjectSpan    shared.Span `json:"-"`
	HasSubjectSpan bool        `json:",omitempty"`
	// SubjectRefID is the parser-assigned NodeID of the reference that fills the
	// subject span for source-death predicates, or -1 when no reference does. The
	// compiler confirms the subject binds the source by matching this identity
	// instead of comparing the reference span to the subject span.
	SubjectRefID int `json:"-"`

	// ControlComparison carries the typed cross-player control-count comparison
	// for ConditionPredicateControlComparison ("an opponent controls more lands
	// than you"). Its zero value is unused.
	ControlComparison ConditionControlComparison `json:",omitzero"`

	// SourceInGraveyard marks a condition introduced by "this card/creature is in
	// your graveyard and ...", as on the Incarnation cycle (Anger, Wonder). It
	// reports that the static ability functions from the graveyard zone; the
	// remaining predicate carries the accompanying runtime condition (e.g. "you
	// control a Mountain").
	SourceInGraveyard bool `json:",omitempty"`

	// GraveyardRedirectScope, GraveyardSubjectTypesAny, and
	// GraveyardFromBattlefieldOnly describe a
	// ConditionPredicateCardWouldGoToGraveyard clause ("If a card would be put
	// into an opponent's graveyard from anywhere, exile it instead."). Scope
	// names whose graveyard is watched; TypesAny restricts the moving card to any
	// of the listed card types (empty matches any card); FromBattlefieldOnly
	// marks the "a permanent" subject, which can only leave the battlefield.
	GraveyardRedirectScope       GraveyardRedirectScope `json:",omitempty"`
	GraveyardSubjectTypesAny     []TriggerCardType      `json:",omitempty"`
	GraveyardFromBattlefieldOnly bool                   `json:",omitempty"`

	// GraveyardRedirectControlScope restricts a
	// ConditionPredicateCardWouldGoToGraveyard clause that watches a dying
	// permanent by who controls it ("If a creature an opponent controls would
	// die, exile it instead."). It is empty for "would be put into a graveyard"
	// forms, which watch the moving card's owner via GraveyardRedirectScope.
	GraveyardRedirectControlScope GraveyardRedirectControlScope `json:",omitempty"`

	// CounterRecipientTypesAny restricts a
	// ConditionPredicateCounterPlacementOnControlledPermanent clause to a
	// controlled permanent that has at least one of the listed card types ("an
	// artifact or creature you control", Ozolith, the Shattered Spire). It is
	// empty for the unrestricted "a permanent you control" form.
	CounterRecipientTypesAny []TriggerCardType `json:",omitempty"`

	// CounterRecipientExcludesSource drops the source permanent from a
	// ConditionPredicateCounterPlacementOnControlledPermanent clause's recipient
	// match ("another creature you control", Benevolent Hydra). It is false for
	// recipient forms that include the source.
	CounterRecipientExcludesSource bool `json:",omitempty"`

	// GraveyardCountCardType carries the single card type counted by a
	// ConditionPredicateGraveyardCardOfTypeCountAtLeast clause ("if twenty or
	// more creature cards are in your graveyard", Mortal Combat). Threshold
	// carries the minimum count. It is TriggerCardTypeUnknown for other clauses.
	GraveyardCountCardType TriggerCardType `json:",omitempty"`

	// ControlledNames carries the card names required by a
	// ConditionPredicateControllerControlsNamed clause ("If you control an
	// Urza's Mine and an Urza's Tower, ..."; the Urza tron lands). The
	// controller must control a permanent matching each listed name. The parser
	// reconstructs each name from the source tokens; matching is normalized
	// downstream.
	ControlledNames []string `json:",omitempty"`

	// ManaSpentColor carries the color required by a
	// ConditionPredicateColoredManaSpentToCastAtLeast clause ("if at least three
	// white mana was spent to cast this spell"; the Adamant ability word).
	// Threshold carries the minimum amount of that color of mana. It is
	// TriggerColorUnknown for the same-color form, which compares the largest
	// single-color tally instead of a named color.
	ManaSpentColor TriggerColor `json:",omitempty"`

	// Negated marks a clause whose recognized wording is the logical negation of
	// its positive predicate ("there are no <kind> counters on this land" is a
	// negated ObjectMatches whose positive form means "has at least one counter
	// of that kind"). The compiler flips Condition.Negated when this is set.
	Negated bool `json:",omitempty"`
}

// ConditionControlComparison describes a cross-player control-count comparison
// ("an opponent controls more lands than you"). LeftScope counts the subject
// player group and RightScope the "than" reference group; Greater is true for
// "more" and false for "fewer"/"less". The counted permanent Selection is
// carried on the enclosing ConditionClause.
type ConditionControlComparison struct {
	LeftScope  ConditionControlScope `json:",omitempty"`
	RightScope ConditionControlScope `json:",omitempty"`
	Greater    bool                  `json:",omitempty"`
}

func emitConditionClauses(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		// Remove a trailing "Activate only …" timing span, including the "and only
		// <timing>" tail peeled off an "Activate only if <condition> and only
		// <timing>" gate, so the condition clause recognizer sees only the
		// "only if <condition>" prefix rather than the conjoined timing wording.
		if span, ok := ability.activationTimingSpan(); ok {
			tokens = tokensOutsideParserSpan(tokens, span)
		}
		if clauses := parseConditionClauses(tokens, ability.Atoms); len(clauses) > 0 {
			ability.ConditionClauses = clauses
		}
		if ability.AlternativeCost != nil {
			ability.ConditionClauses = clausesOutsideSpan(ability.ConditionClauses, ability.AlternativeCost.Span)
		}
		if ability.Modal == nil {
			continue
		}
		for j := range ability.Modal.Options {
			mode := &ability.Modal.Options[j]
			tokens := eventHistorySemanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
			if clauses := parseConditionClauses(tokens, mode.Atoms); len(clauses) > 0 {
				mode.ConditionClauses = clauses
			}
		}
	}
}

// clausesOutsideSpan drops condition clauses whose span falls within enclosing,
// used to suppress a clause an alternative cost already encodes (e.g. the
// commander-control gate of "if you control your commander, you may cast this
// spell without paying its mana cost").
func clausesOutsideSpan(clauses []ConditionClause, enclosing shared.Span) []ConditionClause {
	var kept []ConditionClause
	for _, clause := range clauses {
		if spanContains(enclosing, clause.Span) {
			continue
		}
		kept = append(kept, clause)
	}
	return kept
}

func parseConditionClauses(tokens []shared.Token, atoms Atoms) []ConditionClause {
	var clauses []ConditionClause
	for i := 0; i < len(tokens); i++ {
		intro, width := conditionIntroAt(tokens, i)
		if intro == ConditionIntroUnknown {
			continue
		}
		if effectWordsAt(tokens, i, creatureSpellHasteConditionWords...) {
			continue
		}
		if entersAsCopyCounterRiderConditionAt(tokens, i) || punisherUnlessClauseAt(tokens, i) {
			continue
		}
		if playFromTopPayLifeRiderConditionAt(tokens, i) {
			continue
		}
		end := conditionClauseEnd(tokens, i)
		if clause, ok := parseConditionClause(tokens[i:end], width, intro, atoms); ok {
			clause.Span = shared.SpanOf(tokens[i:end])
			clauses = append(clauses, clause)
		}
		i = end - 1
	}
	return clauses
}

// entersAsCopyCounterRiderConditionAt reports whether the "if" condition intro at
// index i belongs to an enters-as-copy conditional copiable counter rider ("...
// counter on it if it's a creature"; Spark Double). Such an "if" is parsed into
// the enters-as-copy effect's conditional counters, so it must not also surface
// as a standalone intervening condition. It requires both the preceding "counter
// on it" context and a following "if it's a <type>" predicate, so ordinary
// conditional enters-with-counter clauses ("... counter on it if you control
// ...", Ascendant Packleader) keep their condition.
func entersAsCopyCounterRiderConditionAt(tokens []shared.Token, i int) bool {
	if i < 3 || !equalWord(tokens[i], "if") {
		return false
	}
	if !equalWord(tokens[i-3], "counter") || !equalWord(tokens[i-2], "on") || !equalWord(tokens[i-1], "it") {
		return false
	}
	return entersAsCopyConditionalTypePrefix(normalizedWords(tokens[i:]))
}

// conditionLeaveBattlefieldExileReplacementAt reports whether the "if" at index
// opens the leaves-the-battlefield self-replacement clause "if it would leave
// the battlefield, exile it instead [of putting it anywhere else]." (Whip of
// Erebos). That clause is recognized as a whole-sentence replacement effect by
// parseLeaveBattlefieldExileReplacement, so its leading "if" must not also
// surface as a standalone condition. The subject is "it" (a back-reference) or
// "this <type>" (the source).
func conditionLeaveBattlefieldExileReplacementAt(tokens []shared.Token, i int) bool {
	if !equalWord(tokens[i], "if") || i+1 >= len(tokens) {
		return false
	}
	rest := tokens[i+1:]
	subjectWidth := leaveBattlefieldReplacementSubjectWidth(rest)
	if subjectWidth == 0 {
		return false
	}
	return effectWordsAt(rest, subjectWidth, "would", "leave", "the", "battlefield")
}

// leaveBattlefieldReplacementSubjectWidth reports the token width of the subject
// that opens a leaves-the-battlefield replacement clause ("it" → 1, "this
// <type>" → 2), or 0 when tokens do not begin with such a subject.
func leaveBattlefieldReplacementSubjectWidth(tokens []shared.Token) int {
	if len(tokens) == 0 {
		return 0
	}
	if equalWord(tokens[0], "it") {
		return 1
	}
	if len(tokens) >= 2 && equalWord(tokens[0], "this") {
		return 2
	}
	return 0
}

// conditionDieThisTurnExileReplacementAt reports whether the "if" at index opens
// the single-target damage-spell rider "If that creature [or planeswalker] would
// die this turn, exile it instead." (Lava Coil, Obliterating Bolt). That clause
// is recognized as a whole-sentence replacement effect by
// parseDieThisTurnExileReplacement, so its leading "if" must not also surface as
// a standalone condition.
func conditionDieThisTurnExileReplacementAt(tokens []shared.Token, i int) bool {
	if !equalWord(tokens[i], "if") || i+1 >= len(tokens) {
		return false
	}
	rest := tokens[i+1:]
	subjectWidth := dieThisTurnExileSubjectWidth(rest)
	if subjectWidth == 0 {
		return false
	}
	return effectWordsAt(rest, subjectWidth, "would", "die", "this", "turn")
}

// entersAsCopyConditionalTypePrefix reports whether words begins with the
// "if it's a <type>" / "if it is a <type>" predicate of a conditional copiable
// counter rider, where <type> is a recognized card type.
func entersAsCopyConditionalTypePrefix(words []string) bool {
	var typeWord string
	switch {
	case len(words) >= 4 && words[0] == "if" && words[1] == "it's" && (words[2] == "a" || words[2] == "an"):
		typeWord = words[3]
	case len(words) >= 5 && words[0] == "if" && words[1] == "it" && words[2] == "is" && (words[3] == "a" || words[3] == "an"):
		typeWord = words[4]
	default:
		return false
	}
	_, ok := entersAsCopyConditionalTypeWord(typeWord)
	return ok
}
func conditionIntroAt(tokens []shared.Token, index int) (kind ConditionIntroKind, width int) {
	switch {
	case equalWord(tokens[index], "if"):
		return ConditionIntroIf, 1
	case equalWord(tokens[index], "unless"):
		return ConditionIntroUnless, 1
	case index+1 < len(tokens) &&
		equalWord(tokens[index], "only") &&
		equalWord(tokens[index+1], "if"):
		return ConditionIntroOnlyIf, 2
	case index+2 < len(tokens) &&
		equalWord(tokens[index], "as") &&
		equalWord(tokens[index+1], "long") &&
		equalWord(tokens[index+2], "as"):
		return ConditionIntroAsLongAs, 3
	case isControllerTurnIntro(tokens, index):
		// A sentence-leading "During your turn," gates a continuous self-static on
		// the controller being the active player, exactly like an "As long as it's
		// your turn" clause. The "during" introducer alone is consumed; the "your
		// turn" body is recognized as the controller-turn predicate.
		return ConditionIntroAsLongAs, 1
	case isReflexiveWhenYouDoIntro(tokens, index):
		// The reflexive "When you do," preamble gates its trailing effect on the
		// just-performed optional action having been taken, exactly like an "If
		// you do," rider. Oracle text uses the reflexive trigger form when the
		// dependent effect chooses a new target; modelling it as the same
		// prior-instruction-accepted gate lets the existing optional-flow path
		// lower it. The "when" introducer alone is consumed; the "you do" body
		// is recognized as the prior-instruction-accepted predicate.
		return ConditionIntroIf, 1
	default:
		return ConditionIntroUnknown, 0
	}
}

// isControllerTurnIntro reports whether the tokens at index open a
// sentence-leading "During your turn," preamble that gates a continuous static
// on the controller being the active player (Fresh-Faced Recruit, Embereth
// Skyblazer). Only a clause-leading occurrence is treated as a condition; a
// trailing "... during your turn." (Dragonlord Dromoka) and every other use are
// left untouched so existing static recognizers keep owning that wording.
func isControllerTurnIntro(tokens []shared.Token, index int) bool {
	if index != 0 && tokens[index-1].Kind != shared.Period {
		return false
	}
	return index+2 < len(tokens) &&
		equalWord(tokens[index], "during") &&
		equalWord(tokens[index+1], "your") &&
		equalWord(tokens[index+2], "turn")
}

// reflexive preamble "When you do," that gates its trailing effect on a
// preceding optional action having been taken. Only a reflexive that follows an
// earlier "you may" in the same body is treated as a resolving condition gate;
// a "When you do," after a mandatory action (which always happens) is left to
// the existing sequencing so its trailing effect resolves unconditionally, and
// every other "when"/"whenever" clause is left to the trigger machinery.
func isReflexiveWhenYouDoIntro(tokens []shared.Token, index int) bool {
	if index+3 >= len(tokens) ||
		!equalWord(tokens[index], "when") ||
		!equalWord(tokens[index+1], "you") ||
		!equalWord(tokens[index+2], "do") ||
		tokens[index+3].Kind != shared.Comma {
		return false
	}
	return precedingOptionalMayClause(tokens, index)
}

// precedingOptionalMayClause reports whether a "you may" optional action appears
// in the tokens before index, marking a reflexive "When you do," as gating that
// optional rather than an always-performed mandatory action.
func precedingOptionalMayClause(tokens []shared.Token, index int) bool {
	for i := 0; i+1 < index; i++ {
		if equalWord(tokens[i], "you") && equalWord(tokens[i+1], "may") {
			return true
		}
	}
	return false
}

func conditionClauseEnd(tokens []shared.Token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Kind == shared.Period || i > start && tokens[i].Kind == shared.Comma {
			return i
		}
	}
	return len(tokens)
}

func parseConditionClause(
	tokens []shared.Token,
	introWidth int,
	intro ConditionIntroKind,
	atoms Atoms,
) (ConditionClause, bool) {
	body := tokens[introWidth:]
	if len(body) == 0 {
		return ConditionClause{}, false
	}
	clause, ok := recognizeConditionPredicate(body, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	clause.Intro = intro
	return clause, true
}

func recognizeConditionPredicate(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	for _, recognize := range []func([]shared.Token, Atoms) (ConditionClause, bool){
		recognizePriorInstructionCondition,
		recognizeControlsCommanderCondition,
		recognizeLandEnteredOrControlsBasicCondition,
		recognizeControlsGreatestPowerCondition,
		recognizeControlsGreatestToughnessCondition,
		recognizeDestroyedThisWayCondition,
		recognizeTargetObjectMatchCondition,
		recognizeEventSubjectCondition,
		recognizeSourceSaddledCondition,
		recognizeSourceStateCondition,
		recognizeAttachedCreatureStateCondition,
		recognizeSourceNoCounterCondition,
		recognizeSourceCounterStateCondition,
		recognizeControllerResourceCondition,
		recognizeTriggeringPlayerHandSizeCondition,
		recognizeGainedLifeThisTurnCondition,
		recognizeGraveyardCondition,
		recognizeCounterPlacementCondition,
		recognizeDamageSourceCondition,
		recognizeTokenCreationCondition,
		recognizeLifeGainCondition,
		recognizeLifeLossCondition,
		recognizeControlComparisonCondition,
		recognizeGraveyardControlsCondition,
		recognizeControlsCondition,
		recognizeTotalPowerCondition,
		recognizeControlsNamedCondition,
		recognizeCardToGraveyardReplacementCondition,
		recognizeCreatureWouldDieReplacementCondition,
		recognizeSourceDeathCondition,
		recognizeTargetColorCondition,
		recognizeDrawFromEmptyLibraryCondition,
		recognizeDrawCardReplacementCondition,
		recognizeCastTimingCondition,
		recognizeFirstCombatPhaseCondition,
		recognizeControllerTurnCondition,
		recognizeAttackersAttackingControllerCondition,
		recognizeSpellXCondition,
		recognizeAdamantManaSpentCondition,
		recognizeEventSpellManaSpentCondition,
		recognizeCreatedTokenMatchCondition,
		recognizeSharesCreatureTypeCondition,
		recognizeControllerDesignationCondition,
	} {
		if clause, ok := recognize(body, atoms); ok {
			return clause, true
		}
	}
	return ConditionClause{}, false
}

// recognizeAdamantManaSpentCondition matches the Adamant ability word's gate "at
// least <n> <color> mana was spent to cast this spell" ("Adamant — If at least
// three white mana was spent to cast this spell, this creature enters with a
// +1/+1 counter on it.", the Throne of Eldraine Paladin cycle) and its
// same-color form "at least <n> mana of the same color was spent to cast this
// spell" (Henge Walker). It reads the colored mana actually spent to cast the
// resolving spell (CR 702.132), so it gates a resolving-spell replacement. It
// fails closed on any other wording.
func recognizeAdamantManaSpentCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "at", "least")
	if !ok || len(rest) < 2 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok || value <= 0 {
		return ConditionClause{}, false
	}
	rest = rest[1:]
	tail := []string{"mana", "was", "spent", "to", "cast", "this", "spell"}
	if color, ok := recognizeColorWord(rest[0].Text); ok {
		if !tokenWordsEqual(rest[1:], tail...) {
			return ConditionClause{}, false
		}
		return ConditionClause{
			Predicate:      ConditionPredicateColoredManaSpentToCastAtLeast,
			Threshold:      value,
			ManaSpentColor: triggerColorFromAtom(color),
		}, true
	}
	sameColor := append([]string{"mana", "of", "the", "same", "color"}, tail[1:]...)
	if tokenWordsEqual(rest, sameColor...) {
		return ConditionClause{
			Predicate: ConditionPredicateSameColorManaSpentToCastAtLeast,
			Threshold: value,
		}, true
	}
	return ConditionClause{}, false
}

// recognizeEventSpellManaSpentCondition matches a spell-cast trigger's
// intervening-if gate on the total mana spent to cast the triggering spell,
// referenced as "it" or "that spell": "no mana was spent to cast it" / "...that
// spell" (Boromir, Lavinia, Roiling Vortex) and "at least <n> mana was spent to
// cast it" / "...that spell" (Blazing Bomb, Sahagin, Prompto Argentum,
// Raggadragga). It reads the mana actually paid (CR 601.2f-h), so a free cast
// records no mana spent. It fails closed on any other wording.
func recognizeEventSpellManaSpentCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if rest, ok := cutTokenPrefix(body, "no", "mana", "was", "spent", "to", "cast"); ok {
		if eventSpellManaSpentSubject(rest) {
			return ConditionClause{Predicate: ConditionPredicateEventSpellNoManaSpentToCast}, true
		}
		return ConditionClause{}, false
	}
	rest, ok := cutTokenPrefix(body, "at", "least")
	if !ok || len(rest) < 1 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok || value <= 0 {
		return ConditionClause{}, false
	}
	rest, ok = cutTokenPrefix(rest[1:], "mana", "was", "spent", "to", "cast")
	if !ok || !eventSpellManaSpentSubject(rest) {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateEventSpellManaSpentToCastAtLeast,
		Threshold: value,
	}, true
}

// eventSpellManaSpentSubject reports whether the trailing tokens name the
// triggering spell of a spell-cast trigger, spelled "it" or "that spell".
func eventSpellManaSpentSubject(tokens []shared.Token) bool {
	return tokenWordsEqual(tokens, "it") || tokenWordsEqual(tokens, "that", "spell")
}

// recognizeSpellXCondition matches the resolving-spell value-of-X gate "X is
// <n> or more" / "X is <n> or greater" ("If X is 10 or more, ...", the Finale
// cycle). It reads the chosen value of the spell's {X} cost, so it gates only a
// per-effect branch of a resolving spell. It fails closed on any other wording.
func recognizeSpellXCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "x", "is")
	if !ok || len(rest) < 1 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok {
		return ConditionClause{}, false
	}
	tail := rest[1:]
	if tokenWordsEqual(tail, "or", "more") || tokenWordsEqual(tail, "or", "greater") {
		return ConditionClause{Predicate: ConditionPredicateSpellXAtLeast, Threshold: value}, true
	}
	return ConditionClause{}, false
}

// recognizeControlsCommanderCondition matches the Lieutenant intervening/static
// gate "you control your commander" ("Lieutenant — ..., if you control your
// commander, ..." and "Lieutenant — As long as you control your commander,
// ..."). The commander is a single designated object rather than a filterable
// selection, so it maps to its own closed predicate evaluated against runtime
// commander-control state rather than through the generic "controls" selection
// path. It fails closed on any other wording.
func recognizeControlsCommanderCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "control", "your", "commander") {
		return ConditionClause{Predicate: ConditionPredicateControllerControlsCommander}, true
	}
	return ConditionClause{}, false
}

// recognizeLandEnteredOrControlsBasicCondition matches the disjunctive land
// activation gate "this land entered this turn or if you control a basic land"
// (the Mercadian Masques tap-for-two-colors land cycle: Gleaming Bastion,
// Hidden Lair, Dark Fortress, Training Compound, Gathering Place). It holds when
// either disjunct is true and bars the second activated mana ability the turn
// after the land entered unless the controller already has a basic land. It
// fails closed on any other wording.
func recognizeLandEnteredOrControlsBasicCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body,
		"this", "land", "entered", "this", "turn",
		"or", "if", "you", "control", "a", "basic", "land") {
		return ConditionClause{Predicate: ConditionPredicateLandEnteredThisTurnOrControlsBasic}, true
	}
	return ConditionClause{}, false
}

// recognizeControlsGreatestPowerCondition matches the conditional-draw gate "you
// control the creature with the greatest power or tied for the greatest power"
// (Summon: Fenrir chapter III). The predicate holds when the controller controls
// a creature whose power is at least as high as every other creature's power on
// the battlefield (sole highest or tied for highest). It fails closed on any
// other wording.
func recognizeControlsGreatestPowerCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body,
		"you", "control", "the", "creature", "with", "the", "greatest", "power",
		"or", "tied", "for", "the", "greatest", "power") {
		return ConditionClause{Predicate: ConditionPredicateControlsGreatestPowerCreature}, true
	}
	return ConditionClause{}, false
}

// recognizeControlsGreatestToughnessCondition matches the conditional-draw gate
// "you control the creature with the greatest toughness or tied for the greatest
// toughness" (Abzan Beastmaster). The predicate holds when the controller
// controls a creature whose toughness is at least as high as every other
// creature's toughness on the battlefield (sole highest or tied for highest). It
// fails closed on any other wording.
func recognizeControlsGreatestToughnessCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body,
		"you", "control", "the", "creature", "with", "the", "greatest", "toughness",
		"or", "tied", "for", "the", "greatest", "toughness") {
		return ConditionClause{Predicate: ConditionPredicateControlsGreatestToughnessCreature}, true
	}
	return ConditionClause{}, false
}

func recognizePriorInstructionCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "don't") {
		return ConditionClause{Predicate: ConditionPredicatePriorInstructionNotAccepted}, true
	}
	if tokenWordsEqual(body, "you", "do") {
		return ConditionClause{Predicate: ConditionPredicatePriorInstructionAccepted}, true
	}
	return ConditionClause{}, false
}

// recognizeSharesCreatureTypeCondition matches the Kinship resolving gate "it
// shares a creature type with this creature", where "it" is the just-looked-at
// top card of the controller's library and "this creature" is the source
// permanent. The predicate holds when the looked-at card and the source share at
// least one creature type. It fails closed on any other wording.
func recognizeSharesCreatureTypeCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "it", "shares", "a", "creature", "type", "with", "this", "creature") {
		return ConditionClause{Predicate: ConditionPredicateSubjectSharesCreatureTypeWithSource}, true
	}
	return ConditionClause{}, false
}

// recognizeDestroyedThisWayCondition matches the resolving success gate "a
// <permanent noun> is destroyed this way" (and the plural "are" form) that
// follows a preceding optional destroy effect, as in Noxious Gearhulk's "you may
// destroy another target creature. If a creature is destroyed this way, you gain
// life equal to its toughness." It maps to its own predicate distinct from the
// literal "if you do" gate: the noun names only a descriptive subset of what the
// prior clause could have destroyed (e.g. "if an artifact is destroyed this way"
// after "destroy target artifact or land"), so it is the resolving-success
// equivalent of "if you do" only when that noun matches every possible destroyed
// object. The lowering treats it as an "if you do" gate solely for the existing
// optional-destroy shape and fails closed elsewhere. It fails closed on any other
// wording.
func recognizeDestroyedThisWayCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "a")
	if !ok {
		rest, ok = cutTokenPrefix(body, "an")
	}
	if !ok || len(rest) == 0 {
		return ConditionClause{}, false
	}
	plural, ok := destroyedThisWayNounPlural(rest[0])
	if !ok {
		return ConditionClause{}, false
	}
	copula := "is"
	if plural {
		copula = "are"
	}
	if !tokenWordsEqual(rest[1:], copula, "destroyed", "this", "way") {
		return ConditionClause{}, false
	}
	return ConditionClause{Predicate: ConditionPredicateDestroyedThisWay}, true
}

// recognizeCastTimingCondition handles the Addendum cast-timing gate "you cast
// this spell during your main phase", which restricts the gated effect to
// spells cast while their controller is the active player in a main phase.
// recognizeAttackersAttackingControllerCondition matches the intervening-if
// combat gate "<N> or more of those creatures are attacking you and/or
// planeswalkers you control" (Mangara, the Diplomat; Tomik, Wielder of Law).
// "Those creatures" back-references the attackers declared by the trigger's
// "an opponent attacks with creatures" event; the predicate counts how many of
// those attackers are attacking the controller (directly or one of the
// controller's planeswalkers) and holds when that count meets the threshold N.
// It fails closed on any other wording.
// recognizeFirstCombatPhaseCondition matches the turn-structure gate "it's the
// first combat phase of the turn" ("if it's the first combat phase of the turn,
// there is an additional combat phase after this phase"; Raiyuu, Storm's Edge,
// Karlach, Fury of Avernus). It gates the extra-combat insertion so the loop
// fires only once per turn. It fails closed on any other wording.
func recognizeFirstCombatPhaseCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "it's", "the", "first", "combat", "phase", "of", "the", "turn") ||
		tokenWordsEqual(body, "it", "is", "the", "first", "combat", "phase", "of", "the", "turn") {
		return ConditionClause{Predicate: ConditionPredicateFirstCombatPhaseOfTurn}, true
	}
	return ConditionClause{}, false
}

// recognizeControllerTurnCondition matches the controller-turn body "your turn"
// that the "During your turn," introducer leaves behind (Fresh-Faced Recruit,
// Embereth Skyblazer). It gates a conditional self-static on the controller
// being the active player. It fails closed on any other wording.
func recognizeControllerTurnCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "your", "turn") {
		return ConditionClause{Predicate: ConditionPredicateControllerTurn}, true
	}
	return ConditionClause{}, false
}

func recognizeAttackersAttackingControllerCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if len(body) < 1 {
		return ConditionClause{}, false
	}
	count, ok := CardinalWordValue(body[0].Text)
	if !ok || count < 1 {
		return ConditionClause{}, false
	}
	rest, ok := cutTokenPrefix(body[1:],
		"or", "more", "of", "those", "creatures", "are", "attacking", "you", "and")
	if !ok || len(rest) == 0 || rest[0].Kind != shared.Slash {
		return ConditionClause{}, false
	}
	if !tokenWordsEqual(rest[1:], "or", "planeswalkers", "you", "control") {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateAttackersAttackingControllerAtLeast,
		Threshold: count,
	}, true
}

func recognizeCastTimingCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "cast", "this", "spell", "during", "your", "main", "phase") {
		return ConditionClause{Predicate: ConditionPredicateCastDuringControllerMainPhase}, true
	}
	if tokenWordsEqual(body, "this", "spell", "was", "kicked") {
		return ConditionClause{Predicate: ConditionPredicateSpellWasKicked}, true
	}
	if tokenWordsEqual(body, "this", "spell", "was", "cast", "from", "a", "graveyard") {
		return ConditionClause{Predicate: ConditionPredicateSpellWasCastFromGraveyard}, true
	}
	return ConditionClause{}, false
}

func recognizeEventSubjectCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "cast", "it") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasCastByController}, true
	}
	if tokenWordsEqual(body, "you", "cast", "it", "from", "your", "hand") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasCastFromControllerHand}, true
	}
	if tokenWordsEqual(body, "it", "was", "kicked") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasKicked}, true
	}
	if clause, ok := recognizeSourceNamedWasKickedCondition(body, atoms); ok {
		return clause, true
	}
	if tokenWordsEqual(body, "it", "was", "cast") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasCast}, true
	}
	if clause, ok := recognizeEnteredOrCastFromGraveyardCondition(body); ok {
		return clause, true
	}
	if tokenWordsEqual(body, "tribute", "wasn't", "paid") {
		return ConditionClause{Predicate: ConditionPredicateSourceTributeNotPaid}, true
	}
	if tokenWordsEqual(body, "it", "had", "counters", "on", "it") {
		return ConditionClause{
			Predicate:     ConditionPredicateEventSubjectHadCounters,
			ObjectBinding: ConditionObjectBindingEventPermanent,
		}, true
	}
	if clause, ok := recognizeEventSubjectCounterCondition(body, atoms); ok {
		return clause, true
	}
	if clause, ok := recognizeEventSubjectPowerState(body); ok {
		return clause, true
	}
	if clause, ok := recognizeEventSubjectNameUniqueCondition(body); ok {
		return clause, true
	}
	return recognizeEventSubjectMatchCondition(body, atoms)
}

// recognizeSourceNamedWasKickedCondition recognizes a kicker gate whose subject
// names the source permanent by its own card name or by a "this <type>" phrase
// rather than the bare event pronoun ("If this creature was kicked, it enters
// with N +1/+1 counters on it." — the Invasion/Zendikar kicker creature cycle).
// The bare-pronoun "it was kicked" form is recognized by its caller; this
// recognizer covers the leading replacement-clause subject that names the
// source. The subject before "was kicked" must be only the card name or a
// permanent-type noun phrase; any other subject fails closed.
func recognizeSourceNamedWasKickedCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutSourceNamedSubjectTokens(body, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	if !tokenWordsEqual(rest, "was", "kicked") {
		return ConditionClause{}, false
	}
	return ConditionClause{Predicate: ConditionPredicateEventSubjectWasKicked}, true
}

// cutSourceNamedSubjectTokens consumes a leading source self-subject that names
// the source either by its own card name or by a "this <type>" phrase, and
// returns the remaining tokens. Unlike cutSourceSubjectTokens it does not consume
// the bare pronoun "it" (handled by the event-subject recognizers) and does not
// require a following "has"; the subject after "this" must be only a
// permanent-type noun phrase. It fails closed on any other shape.
func cutSourceNamedSubjectTokens(body []shared.Token, atoms Atoms) ([]shared.Token, bool) {
	if len(body) == 0 {
		return nil, false
	}
	if span, ok := atoms.SelfNameSpanStartingAt(body[0].Span); ok {
		i := 0
		for i < len(body) && spanCovers(span, body[i].Span) {
			i++
		}
		if i == 0 {
			return nil, false
		}
		return body[i:], true
	}
	rest, ok := cutTokenPrefix(body, "this")
	if !ok {
		return nil, false
	}
	wasIndex := tokenWordIndex(rest, "was")
	if wasIndex < 1 {
		return nil, false
	}
	selection, ok := parseConditionSelection(rest[:wasIndex], atoms)
	if !ok || !conditionSelectionEmptyExceptType(selection) {
		return nil, false
	}
	return rest[wasIndex:], true
}

// recognizeCreatedTokenMatchCondition handles the resolving gate "the token is a
// <selection>" that inspects the characteristics of a token a prior effect in
// the same ability just created (Yenna, Redtooth Regent: "If the token is an
// Aura, untap Yenna, then scry 2."). The "the token" subject binds the
// just-created token, so the clause carries a permanent selection matched
// against that object. Only the singular "the token" subject is recognized; any
// other subject falls through to the remaining recognizers.
func recognizeCreatedTokenMatchCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "the", "token", "is", "a")
	if !ok {
		if rest, ok = cutTokenPrefix(body, "the", "token", "is", "an"); !ok {
			return ConditionClause{}, false
		}
	}
	selection, ok := parseConditionSelection(rest, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingCreatedToken,
		Selection:     selection,
	}, true
}

// intervening condition that gates a trigger on the entering object(s) having
// come from a graveyard, either by entering directly from a graveyard or by
// being cast from a graveyard. Two oracle wordings carry different ownership
// scopes that lower to distinct predicates: the controller-scoped "your
// graveyard" / "you cast it" form (Prized Amalgam, Archfiend's Vessel) requires
// the source graveyard to belong to the trigger controller, while the
// any-graveyard "a graveyard" form (Twilight Diviner) does not. Both the
// singular ("it") and plural ("they") subjects are recognized. Any other zone
// wording fails closed.
func recognizeEnteredOrCastFromGraveyardCondition(body []shared.Token) (ConditionClause, bool) {
	for _, words := range [][]string{
		{"it", "entered", "from", "your", "graveyard", "or", "you", "cast", "it", "from", "your", "graveyard"},
		{"they", "entered", "from", "your", "graveyard", "or", "you", "cast", "them", "from", "your", "graveyard"},
	} {
		if tokenWordsEqual(body, words...) {
			return ConditionClause{Predicate: ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard}, true
		}
	}
	for _, words := range [][]string{
		{"it", "entered", "or", "was", "cast", "from", "a", "graveyard"},
		{"they", "entered", "or", "were", "cast", "from", "a", "graveyard"},
	} {
		if tokenWordsEqual(body, words...) {
			return ConditionClause{Predicate: ConditionPredicateEventSubjectEnteredOrCastFromGraveyard}, true
		}
	}
	return ConditionClause{}, false
}

// recognizeEventSubjectPowerState handles the triggering object's own power
// threshold "its power is <n> or greater" ("Whenever a creature you control
// enters, draw a card if its power is 3 or greater.") and the past-tense dies
// form "its power was <n> or greater" (Deathknell Berserker). The possessive
// "its" binds the event permanent, so the recognized clause carries a
// power-at-least selection matched against that object; for the dying creature
// the runtime reads its power from last-known information (CR 603.10).
func recognizeEventSubjectPowerState(body []shared.Token) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "its", "power", "is")
	if !ok {
		if rest, ok = cutTokenPrefix(body, "its", "power", "was"); !ok {
			return ConditionClause{}, false
		}
	}
	if len(rest) != 3 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok || !equalWord(rest[1], "or") || !equalWord(rest[2], "greater") {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingEventPermanent,
		Selection:     ConditionSelection{PowerAtLeast: value, MatchPowerAtLeast: true},
	}, true
}

// recognizeEventSubjectNameUniqueCondition handles the name-uniqueness
// intervening condition "it doesn't have the same name as another creature you
// control or a creature card in your graveyard" (Guardian Project). It compares
// the entering creature's name against the other creatures the controller
// controls and the creature cards in their graveyard.
func recognizeEventSubjectNameUniqueCondition(body []shared.Token) (ConditionClause, bool) {
	if tokenWordsEqual(body,
		"it", "doesn't", "have", "the", "same", "name", "as",
		"another", "creature", "you", "control",
		"or", "a", "creature", "card", "in", "your", "graveyard") {
		return ConditionClause{
			Predicate:     ConditionPredicateEventSubjectNameUnique,
			ObjectBinding: ConditionObjectBindingEventPermanent,
		}, true
	}
	return ConditionClause{}, false
}

// recognizeEventSubjectCounterCondition handles the dying creature's last-known
// counter state. The negative form "it had no <counter> counters [on it]"
// (Undying/Persist reminder text) tests the absence of a counter kind; the
// positive form "it had a <counter> counter [on it]" and the equivalent
// "it had one or more <counter> counters [on it]" test the presence of at least
// one counter of that kind ("When this creature dies, if it had a +1/+1 counter
// on it, draw a card." — Promising Duskmage). Both forms read the permanent's
// last-known information at the moment it left the battlefield (CR 603.10).
func recognizeEventSubjectCounterCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if clause, ok := recognizeEventSubjectHadNoCounterCondition(body, atoms); ok {
		return clause, true
	}
	return recognizeEventSubjectHadCounterCondition(body, atoms)
}

func recognizeEventSubjectHadNoCounterCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "it", "had", "no")
	if !ok {
		return ConditionClause{}, false
	}
	if trimmed, ok := stripTokenSuffix(rest, "on", "it"); ok {
		rest = trimmed
	}
	if !tokenSuffixWord(rest, "counters") {
		return ConditionClause{}, false
	}
	counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateEventSubjectHadNoCounter,
		Counter:   counterKind,
	}, true
}

func recognizeEventSubjectHadCounterCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "it", "had", "a")
	if !ok {
		if rest, ok = cutTokenPrefix(body, "it", "had", "one", "or", "more"); !ok {
			return ConditionClause{}, false
		}
	}
	if trimmed, ok := stripTokenSuffix(rest, "on", "it"); ok {
		rest = trimmed
	}
	if !tokenSuffixWord(rest, "counter") && !tokenSuffixWord(rest, "counters") {
		return ConditionClause{}, false
	}
	counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateEventSubjectHadCounter,
		Counter:   counterKind,
	}, true
}

// recognizeEventSubjectMatchCondition handles "it was a <selection>" and the
// "it's a <selection>" contraction, binding the event permanent.
func recognizeEventSubjectMatchCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "it", "was", "a")
	if !ok {
		if rest, ok = cutTokenPrefix(body, "it", "was", "an"); !ok {
			if rest, ok = cutTokenPrefix(body, "it's", "a"); !ok {
				if rest, ok = cutTokenPrefix(body, "it's", "an"); !ok {
					return ConditionClause{}, false
				}
			}
		}
	}
	selection, ok := parseConditionSelection(rest, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingEventPermanent,
		Selection:     selection,
	}, true
}

// recognizeTargetColorCondition handles the bare-color target rider "it's
// <color>" / "it is <color>" that follows a single-target counter or destroy
// effect (Pyroblast, Red Elemental Blast: "Counter target spell if it's blue."
// / "Destroy target permanent if it's blue."). The "it" refers to the effect's
// chosen target, so the predicate is bound to the target by the counter/destroy
// lowering rather than to the source or a triggering event. Only a single bare
// color word is accepted; any noun ("it's a creature") is handled by
// recognizeEventSubjectMatchCondition instead.
func recognizeTargetColorCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "it's")
	if !ok {
		if rest, ok = cutTokenPrefix(body, "it", "is"); !ok {
			return ConditionClause{}, false
		}
	}
	if len(rest) != 1 {
		return ConditionClause{}, false
	}
	color, ok := atoms.ColorAt(rest[0].Span)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateTargetColor,
		Selection: ConditionSelection{ColorsAny: []TriggerColor{triggerColorFromAtom(color)}},
	}, true
}

// recognizeTargetObjectMatchCondition handles "it's a <selection>", "it's an
// <selection>", and the bare supertype form "it's <supertype>" (e.g. "if it's
// legendary"). It binds the condition's object to the spell's target permanent
// rather than to a triggering event permanent; the lowering resolves the target
// at runtime. Only the "it's" contraction is accepted; "it was a <selection>"
// is handled by recognizeEventSubjectMatchCondition for trigger bodies.
//
// Only selections that resolve to subtypes or supertypes (with no required card
// types) are accepted here. Selections with required card types (e.g. "a
// creature") fall through to recognizeEventSubjectCondition so that those forms
// keep their EventPermanent binding in trigger-body contexts.
func recognizeTargetObjectMatchCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "it's", "a")
	if !ok {
		if rest, ok = cutTokenPrefix(body, "it's", "an"); !ok {
			// Bare "it's <supertype>" form — no article required.
			bare, ok2 := cutTokenPrefix(body, "it's")
			if !ok2 || len(bare) == 0 {
				return ConditionClause{}, false
			}
			var supertypes []ConditionSupertype
			for _, tok := range bare {
				st, ok3 := conditionSupertypeAtom(tok.Span, atoms)
				if !ok3 {
					return ConditionClause{}, false
				}
				supertypes = append(supertypes, st)
			}
			return ConditionClause{
				Predicate:     ConditionPredicateObjectMatches,
				ObjectBinding: ConditionObjectBindingTarget,
				Selection:     ConditionSelection{Supertypes: supertypes},
			}, true
		}
	}
	selection, ok := parseConditionSelection(rest, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	// Only accept selections that resolve purely through subtypes/supertypes.
	// Selections with required card types (e.g. "a creature", "a legendary
	// creature") are left for recognizeEventSubjectCondition so they keep the
	// EventPermanent binding used by trigger intervening-if conditions.
	if len(selection.RequiredTypes) > 0 {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingTarget,
		Selection:     selection,
	}, true
}

// recognizeSourceSaddledCondition matches the per-effect gate "this <noun> is
// saddled" / "this <noun> isn't saddled", testing the source Mount's runtime
// saddled state (CR 702.166). It gates Caustic Bronco's split life-loss effect
// ("... if this creature isn't saddled. Otherwise, ...") and the affirmative
// "is saddled" form. The subject noun binds the source, so the predicate alone
// carries the meaning; the "isn't" wording maps to the negated predicate so the
// otherwise/instead negation machinery produces the complementary branch.
func recognizeSourceSaddledCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "this")
	if !ok {
		return ConditionClause{}, false
	}
	if len(rest) != 3 || rest[0].Kind != shared.Word || !equalWord(rest[2], "saddled") {
		return ConditionClause{}, false
	}
	switch {
	case equalWord(rest[1], "is"):
		return ConditionClause{Predicate: ConditionPredicateSourceSaddled}, true
	case equalWord(rest[1], "isn't"):
		return ConditionClause{Predicate: ConditionPredicateSourceNotSaddled}, true
	}
	return ConditionClause{}, false
}

// inspect the source permanent.
func recognizeSourceStateCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if clause, ok := recognizeSourcePronounStateCondition(body, atoms); ok {
		return clause, true
	}
	rest, ok := cutTokenPrefix(body, "this")
	if !ok {
		// A possessive self-name subject ("Kitsa's power is 3 or greater") names
		// the source permanent directly rather than through "this <type>". Keep
		// the possessive name token in rest so the shared "is" split treats
		// "<name>'s power" as the subject, exactly as "this creature's power"
		// does. Any non-self-name body still fails closed here.
		if _, named := atoms.SelfNameSpanStartingAt(body[0].Span); !named {
			return ConditionClause{}, false
		}
		rest = body
	}
	isIndex := tokenWordIndex(rest, "is")
	if isIndex < 1 {
		return ConditionClause{}, false
	}
	subjectTokens := rest[:isIndex]
	stateTokens := rest[isIndex+1:]
	if tokenWordsEqual(stateTokens, "on", "the", "battlefield") {
		selection, ok := parseConditionSelection(subjectTokens, atoms)
		if !ok || !conditionSelectionEmptyExceptType(selection) {
			return ConditionClause{}, false
		}
		return ConditionClause{
			Predicate:     ConditionPredicateObjectExists,
			ObjectBinding: ConditionObjectBindingSource,
		}, true
	}
	// "this <subject>'s power is <n> or greater" inspects the source
	// permanent's own power, e.g. "Activate only if this creature's power is 4
	// or greater". The possessive subject binds the source, so the type filter
	// is redundant with the source binding and only the power threshold is kept.
	if selection, ok := recognizeSourcePowerState(subjectTokens, stateTokens); ok {
		return ConditionClause{
			Predicate:     ConditionPredicateObjectMatches,
			ObjectBinding: ConditionObjectBindingSource,
			Selection:     selection,
		}, true
	}
	selection, ok := parseConditionSelection(subjectTokens, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	if !applySourceState(stateTokens, atoms, &selection) {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingSource,
		Selection:     selection,
	}, true
}

// recognizeSourcePronounStateCondition matches a bare pronoun subject ("it's
// attacking", "it is untapped", "it's equipped") that refers to the source
// permanent in a self-static gate ("This creature has first strike as long as
// it's attacking."). The pronoun carries no subject noun, so the source binding
// alone identifies the inspected permanent and the state words fill the
// selection. Only the source-state vocabulary (tap/combat/attachment) is
// accepted; "it's a <type>" and "it's <color>" are handled by the event-subject
// and target-color recognizers earlier in the dispatch chain.
func recognizeSourcePronounStateCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	stateTokens, ok := cutTokenPrefix(body, "it's")
	if !ok {
		if stateTokens, ok = cutTokenPrefix(body, "it", "is"); !ok {
			return ConditionClause{}, false
		}
	}
	var selection ConditionSelection
	if !applySourceState(stateTokens, atoms, &selection) {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingSource,
		Selection:     selection,
	}, true
}

// recognizeSourcePowerState recognizes a possessive "<subject>'s power" subject
// paired with an "<n> or greater" state, binding the source permanent's power.
func recognizeSourcePowerState(subjectTokens, stateTokens []shared.Token) (ConditionSelection, bool) {
	if len(subjectTokens) != 2 ||
		subjectTokens[0].Kind != shared.Word ||
		!strings.HasSuffix(subjectTokens[0].Text, "'s") ||
		!equalWord(subjectTokens[1], "power") {
		return ConditionSelection{}, false
	}
	if len(stateTokens) != 3 {
		return ConditionSelection{}, false
	}
	value, ok := conditionNumberValue(stateTokens[0])
	if !ok || !equalWord(stateTokens[1], "or") || !equalWord(stateTokens[2], "greater") {
		return ConditionSelection{}, false
	}
	return ConditionSelection{PowerAtLeast: value, MatchPowerAtLeast: true}, true
}

func applySourceState(stateTokens []shared.Token, atoms Atoms, selection *ConditionSelection) bool {
	switch {
	case tokenWordsEqual(stateTokens, "untapped"):
		selection.Tapped = ConditionTappedFalse
		return true
	case tokenWordsEqual(stateTokens, "tapped"):
		selection.Tapped = ConditionTappedTrue
		return true
	case tokenWordsEqual(stateTokens, "attacking"):
		selection.CombatState = ConditionCombatAttacking
		return true
	case tokenWordsEqual(stateTokens, "blocking"):
		selection.CombatState = ConditionCombatBlocking
		return true
	case tokenWordsEqual(stateTokens, "attacking", "or", "blocking"):
		selection.CombatState = ConditionCombatAttackingOrBlocking
		return true
	case tokenWordsEqual(stateTokens, "equipped"):
		selection.Attachment = ConditionAttachmentEquipped
		return true
	case tokenWordsEqual(stateTokens, "enchanted"):
		selection.Attachment = ConditionAttachmentEnchanted
		return true
	}
	// "this permanent is an enchantment": the state is a typed card type.
	var rest []shared.Token
	if trimmed, ok := cutTokenPrefix(stateTokens, "a"); ok {
		rest = trimmed
	} else if trimmed, ok := cutTokenPrefix(stateTokens, "an"); ok {
		rest = trimmed
	} else {
		return false
	}
	typeSelection, ok := parseConditionSelection(rest, atoms)
	if !ok || len(typeSelection.RequiredTypes) == 0 {
		return false
	}
	selection.RequiredTypes = append(selection.RequiredTypes, typeSelection.RequiredTypes...)
	selection.SubtypesAny = append(selection.SubtypesAny, typeSelection.SubtypesAny...)
	selection.ColorsAny = append(selection.ColorsAny, typeSelection.ColorsAny...)
	selection.Supertypes = append(selection.Supertypes, typeSelection.Supertypes...)
	return true
}

// recognizeAttachedCreatureStateCondition matches the conditional-grant gate
// "equipped <subject> is <state>" / "enchanted <subject> is <state>" used by
// Equipment and Auras ("As long as equipped creature is legendary, it has
// hexproof."; "As long as enchanted permanent is a creature, it gets +1/+1.").
// The subject noun ("creature", "permanent", or "land") names the permanent the
// source is attached to; the state is a supertype (e.g. "legendary"), the
// attached object's color(s), card type(s), or subtype(s), or a tap/combat
// state. It binds the attached object so a static grant can gate on the
// equipped or enchanted permanent's own characteristics.
func recognizeAttachedCreatureStateCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutAttachedSubjectPrefix(body)
	if !ok {
		return ConditionClause{}, false
	}
	var selection ConditionSelection
	if !applyAttachedCreatureState(rest, atoms, &selection) {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingSourceAttached,
		Selection:     selection,
	}, true
}

// cutAttachedSubjectPrefix strips the "equipped/enchanted <subject> is" lead-in
// naming the permanent the source is attached to. The subject noun varies with
// the source: Equipment prints "equipped creature", an Aura that enchants any
// permanent prints "enchanted permanent", and an Aura that enchants a land
// prints "enchanted land". All bind the same attached object.
func cutAttachedSubjectPrefix(body []shared.Token) ([]shared.Token, bool) {
	for _, attachment := range []string{"equipped", "enchanted"} {
		for _, subject := range []string{"creature", "permanent", "land"} {
			if rest, ok := cutTokenPrefix(body, attachment, subject, "is"); ok {
				return rest, true
			}
		}
	}
	return nil, false
}

// applyAttachedCreatureState fills the selection from the state words following
// "equipped/enchanted <subject> is ...". A bare supertype ("legendary") sets the
// supertype filter; a characteristic predicate names the attached object's
// color(s), card type(s), or subtype(s) ("a creature", "black", "a Human",
// "red or green"); otherwise the state falls through to the shared source-state
// vocab (tapped/untapped, attacking/blocking).
func applyAttachedCreatureState(stateTokens []shared.Token, atoms Atoms, selection *ConditionSelection) bool {
	if supertypes, ok := conditionStateSupertypes(stateTokens, atoms); ok {
		selection.Supertypes = append(selection.Supertypes, supertypes...)
		return true
	}
	if applyAttachedCharacteristicState(stateTokens, atoms, selection) {
		return true
	}
	return applySourceState(stateTokens, atoms, selection)
}

// applyAttachedCharacteristicState recognizes the attached object's printed
// characteristics as a condition state: a bare color or color disjunction
// ("black", "red or green") or an article-led card-type or subtype predicate
// ("a creature", "a Human", "a Human or an Angel"). It fills the selection's
// color, type, subtype, and supertype filters and reports whether the state was
// a recognized characteristic. A bare color carries no article, so it is tried
// before the article is stripped.
func applyAttachedCharacteristicState(stateTokens []shared.Token, atoms Atoms, selection *ConditionSelection) bool {
	if colors, ok := bareColorList(stateTokens, atoms); ok {
		selection.ColorsAny = append(selection.ColorsAny, colors...)
		return true
	}
	rest := stateTokens
	if trimmed, ok := cutTokenPrefix(rest, "a"); ok {
		rest = trimmed
	} else if trimmed, ok := cutTokenPrefix(rest, "an"); ok {
		rest = trimmed
	} else {
		return false
	}
	parsed, ok := parseConditionSelection(rest, atoms)
	if !ok || conditionSelectionCharacteristicEmpty(parsed) {
		return false
	}
	selection.RequiredTypes = append(selection.RequiredTypes, parsed.RequiredTypes...)
	selection.SubtypesAny = append(selection.SubtypesAny, parsed.SubtypesAny...)
	selection.ColorsAny = append(selection.ColorsAny, parsed.ColorsAny...)
	selection.Supertypes = append(selection.Supertypes, parsed.Supertypes...)
	return true
}

// bareColorList reads one or more color atoms joined by "or" with no trailing
// noun ("black", "red or green"), the form the attached-object color predicate
// "enchanted creature is <color>" prints. It reports false for any token that is
// not a color or the "or" conjunction so non-color states fall through.
func bareColorList(tokens []shared.Token, atoms Atoms) ([]TriggerColor, bool) {
	if len(tokens) == 0 {
		return nil, false
	}
	colors := make([]TriggerColor, 0, len(tokens))
	for _, token := range tokens {
		if equalWord(token, "or") {
			continue
		}
		color, ok := atoms.ColorAt(token.Span)
		if !ok {
			return nil, false
		}
		colors = append(colors, triggerColorFromAtom(color))
	}
	if len(colors) == 0 {
		return nil, false
	}
	return colors, true
}

// conditionSelectionCharacteristicEmpty reports whether a parsed selection
// carries none of the printed characteristic filters (color, card type,
// subtype, supertype) the attached-object state predicate constrains. It rejects
// selections that matched only structural facets so the recognizer falls through
// to the shared tap/combat source-state vocabulary.
func conditionSelectionCharacteristicEmpty(selection ConditionSelection) bool {
	return len(selection.RequiredTypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		len(selection.Supertypes) == 0
}

// conditionStateSupertypes reads one or more bare supertype words ("legendary",
// "snow", "basic") that form a complete predicate state with no trailing noun.
func conditionStateSupertypes(tokens []shared.Token, atoms Atoms) ([]ConditionSupertype, bool) {
	if len(tokens) == 0 {
		return nil, false
	}
	supertypes := make([]ConditionSupertype, 0, len(tokens))
	for _, token := range tokens {
		supertype, ok := conditionSupertypeAtom(token.Span, atoms)
		if !ok {
			return nil, false
		}
		supertypes = append(supertypes, supertype)
	}
	return supertypes, true
}

// recognizeSourceCounterStateCondition handles the source permanent's
// counter-presence intervening condition "<source> has counters on it" /
// "<source> has a counter on it" ("At the beginning of combat on your turn, if
// The Ozolith has counters on it, you may move all counters ..."). The subject
// is the source permanent, named either by "this <type>" or by the card's own
// name; the predicate is the kind-agnostic any-counter presence test bound to
// the source.
func recognizeSourceCounterStateCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutSourceSubjectTokens(body, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	if selection, ok := sourceCounterCountSelection(rest, atoms); ok {
		return ConditionClause{
			Predicate:     ConditionPredicateObjectMatches,
			ObjectBinding: ConditionObjectBindingSource,
			Selection:     selection,
		}, true
	}
	if !tokenWordsEqual(rest, "has", "counters", "on", "it") &&
		!tokenWordsEqual(rest, "has", "a", "counter", "on", "it") {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingSource,
		Selection:     ConditionSelection{AnyCounter: true},
	}, true
}

// sourceCounterCountSelection recognizes a kind-specific source counter-state
// body. It accepts the named-counter-count threshold "has <n> or more <kind>
// counters on it" ("As long as ~ has seven or more quest counters on it, ...",
// the Ascension cycle) and the singular kind-specific presence "has a <kind>
// counter on it" ("If this creature has a +1/+1 counter on it, ...", Incubation
// Druid), which means one or more counters of that kind. It returns a Selection
// carrying the counter kind and minimum count.
func sourceCounterCountSelection(rest []shared.Token, atoms Atoms) (ConditionSelection, bool) {
	after, ok := cutTokenPrefix(rest, "has")
	if !ok {
		return ConditionSelection{}, false
	}
	after, ok = stripTokenSuffix(after, "on", "it")
	if !ok {
		return ConditionSelection{}, false
	}
	atLeast, ok := sourceCounterThreshold(after)
	if !ok {
		return ConditionSelection{}, false
	}
	kind, _, ok := atoms.CounterIn(shared.SpanOf(after))
	if !ok {
		return ConditionSelection{}, false
	}
	return ConditionSelection{
		CounterKind:         kind,
		CounterKindKnown:    true,
		CounterCountAtLeast: atLeast,
	}, true
}

// sourceCounterThreshold reads the minimum count of a kind-specific source
// counter-state body whose tokens follow "has" and precede "on it". It accepts
// the plural threshold "<n> or more <kind> counters" (n or more of that kind)
// and the singular presence "a <kind> counter" (one or more of that kind). It
// fails closed on any other shape.
func sourceCounterThreshold(after []shared.Token) (int, bool) {
	if tokenSuffixWord(after, "counters") {
		count, _, ok := parseLeadingCount(after)
		if !ok || count.Comparison != ConditionComparisonAtLeast {
			return 0, false
		}
		return count.Value, true
	}
	if tokenSuffixWord(after, "counter") &&
		startsWithWord(after, "a", "an") {
		return 1, true
	}
	return 0, false
}

// recognizeSourceNoCounterCondition matches the negated source counter-state
// form "there are no <kind> counters on this <type>" (Mercadian Masques depletion
// taplands: "If there are no depletion counters on this land, sacrifice it."). It
// produces a negated ObjectMatches clause with Source binding whose Selection
// carries the counter kind and a minimum count of 1, so the condition means
// "source does NOT have >= 1 counter of kind" = "source has zero counters of kind".
func recognizeSourceNoCounterCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "there", "are", "no")
	if !ok {
		return ConditionClause{}, false
	}
	// Strip trailing source subject: "on this <type>" or "on it".
	inner, ok := stripSourceSuffix(rest)
	if !ok {
		return ConditionClause{}, false
	}
	if !tokenSuffixWord(inner, "counters") {
		return ConditionClause{}, false
	}
	kind, _, ok := atoms.CounterIn(shared.SpanOf(rest))
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:     ConditionPredicateObjectMatches,
		ObjectBinding: ConditionObjectBindingSource,
		Selection: ConditionSelection{
			CounterKind:         kind,
			CounterKindKnown:    true,
			CounterCountAtLeast: 1,
		},
		Negated: true,
	}, true
}

// stripSourceSuffix strips a trailing "on this <type>" or "on it" source subject
// from a token slice, returning the remaining inner tokens. It fails closed when
// neither form is present.
func stripSourceSuffix(tokens []shared.Token) ([]shared.Token, bool) {
	if trimmed, ok := stripTokenSuffix(tokens, "on", "it"); ok {
		return trimmed, true
	}
	// Try "on this <type>": require at least 3 trailing tokens "on this <noun>".
	n := len(tokens)
	if n >= 3 &&
		strings.EqualFold(tokens[n-3].Text, "on") &&
		strings.EqualFold(tokens[n-2].Text, "this") {
		return tokens[:n-3], true
	}
	return nil, false
}

// cutSourceSubjectTokens consumes a leading source self-subject — the card's own
// name, a "this <type>" phrase, or the bare pronoun "it" that refers back to the
// source in a self gate ("This creature has trample as long as it has a +1/+1
// counter on it.") — and returns the remaining state tokens. It fails closed
// when the body does not begin with a recognized source subject.
func cutSourceSubjectTokens(body []shared.Token, atoms Atoms) ([]shared.Token, bool) {
	if len(body) == 0 {
		return nil, false
	}
	if rest, ok := cutTokenPrefix(body, "it"); ok {
		return rest, true
	}
	if span, ok := atoms.SelfNameSpanStartingAt(body[0].Span); ok {
		i := 0
		for i < len(body) && spanCovers(span, body[i].Span) {
			i++
		}
		if i == 0 {
			return nil, false
		}
		return body[i:], true
	}
	rest, ok := cutTokenPrefix(body, "this")
	if !ok {
		return nil, false
	}
	hasIndex := tokenWordIndex(rest, "has")
	if hasIndex < 1 {
		return nil, false
	}
	selection, ok := parseConditionSelection(rest[:hasIndex], atoms)
	if !ok || !conditionSelectionEmptyExceptType(selection) {
		return nil, false
	}
	return rest[hasIndex:], true
}

func recognizeControllerResourceCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "have", "no", "cards", "in", "hand") {
		return ConditionClause{Predicate: ConditionPredicateControllerHandEmpty}, true
	}
	rest, ok := cutTokenPrefix(body, "you", "have")
	if ok {
		// "you have at least <n> life [more than your starting life total]".
		if atLeast, ok := cutTokenPrefix(rest, "at", "least"); ok && len(atLeast) >= 1 {
			if value, ok := conditionNumberValue(atLeast[0]); ok {
				tail := atLeast[1:]
				switch {
				case tokenWordsEqual(tail, "life", "more", "than", "your", "starting", "life", "total"):
					return ConditionClause{Predicate: ConditionPredicateControllerLifeAtLeastAboveStarting, Threshold: value}, true
				case tokenWordsEqual(tail, "life"):
					return ConditionClause{Predicate: ConditionPredicateControllerLifeAtLeast, Threshold: value}, true
				}
			}
		}
		if count, tail, ok := parseLeadingCount(rest); ok {
			switch count.Comparison {
			case ConditionComparisonAtLeast:
				switch {
				case tokenWordsEqual(tail, "cards", "in", "hand"),
					tokenWordsEqual(tail, "cards", "in", "your", "hand"):
					return ConditionClause{Predicate: ConditionPredicateControllerHandSizeAtLeast, Threshold: count.Value}, true
				case tokenWordsEqual(tail, "cards", "in", "your", "library"):
					return ConditionClause{Predicate: ConditionPredicateControllerLibrarySizeAtLeast, Threshold: count.Value}, true
				case tokenWordsEqual(tail, "life"):
					return ConditionClause{Predicate: ConditionPredicateControllerLifeAtLeast, Threshold: count.Value}, true
				case tokenWordsEqual(tail, "opponents"):
					return ConditionClause{Predicate: ConditionPredicateOpponentCountAtLeast, Threshold: count.Value}, true
				}
			case ConditionComparisonAtMost:
				if tokenWordsEqual(tail, "life") {
					return ConditionClause{Predicate: ConditionPredicateControllerLifeAtMost, Threshold: count.Value}, true
				}
			default:
			}
		}
		// "you have exactly <n> cards in hand" is an equality on hand size, e.g.
		// "Activate only if you have exactly seven cards in hand".
		if exact, ok := cutTokenPrefix(rest, "exactly"); ok && len(exact) >= 1 {
			if value, ok := conditionNumberValue(exact[0]); ok {
				tail := exact[1:]
				switch {
				case tokenWordsEqual(tail, "cards", "in", "hand"),
					tokenWordsEqual(tail, "cards", "in", "your", "hand"):
					return ConditionClause{Predicate: ConditionPredicateControllerHandSizeExactly, Threshold: value}, true
				case tokenWordsEqual(tail, "life"):
					return ConditionClause{Predicate: ConditionPredicateControllerLifeExactly, Threshold: value}, true
				}
			}
		}
	}
	rest, ok = cutTokenPrefix(body, "an", "opponent", "has")
	if ok {
		if count, tail, ok := parseLeadingCount(rest); ok &&
			count.Comparison == ConditionComparisonAtLeast &&
			tokenWordsEqual(tail, "poison", "counters") {
			return ConditionClause{Predicate: ConditionPredicateAnyOpponentPoisonAtLeast, Threshold: count.Value}, true
		}
	}
	rest, ok = cutTokenPrefix(body, "a", "player", "has")
	if ok {
		if count, tail, ok := parseLeadingCount(rest); ok &&
			count.Comparison == ConditionComparisonAtMost &&
			tokenWordsEqual(tail, "life") {
			return ConditionClause{Predicate: ConditionPredicateAnyPlayerLifeAtMost, Threshold: count.Value}, true
		}
	}
	return ConditionClause{}, false
}

// recognizeTriggeringPlayerHandSizeCondition matches an intervening-if body that
// compares the triggering player's hand size against a threshold: "that player
// has no cards in hand", "that player has two or fewer cards in hand", or "that
// player has five or more cards in hand". "that player" denotes the player whose
// step began the trigger (the active player on each opponent's or each player's
// upkeep), so the condition resolves against the triggering event's player.
func recognizeTriggeringPlayerHandSizeCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "that", "player", "has")
	if !ok {
		return ConditionClause{}, false
	}
	if tokenWordsEqual(rest, "no", "cards", "in", "hand") {
		return ConditionClause{Predicate: ConditionPredicateTriggeringPlayerHandSizeAtMost, Threshold: 0}, true
	}
	count, tail, ok := parseLeadingCount(rest)
	if !ok || !tokenWordsEqual(tail, "cards", "in", "hand") {
		return ConditionClause{}, false
	}
	switch count.Comparison {
	case ConditionComparisonAtMost:
		return ConditionClause{Predicate: ConditionPredicateTriggeringPlayerHandSizeAtMost, Threshold: count.Value}, true
	case ConditionComparisonAtLeast:
		return ConditionClause{Predicate: ConditionPredicateTriggeringPlayerHandSizeAtLeast, Threshold: count.Value}, true
	default:
		return ConditionClause{}, false
	}
}

// recognizeGainedLifeThisTurnCondition matches the intervening-if body
// "you gained <n> or more life this turn", e.g. Angelic Accord's
// "At the beginning of each end step, if you gained 3 or more life this turn,
// create a 4/4 white Angel creature token with flying."
func recognizeGainedLifeThisTurnCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "you", "gained")
	if !ok {
		return ConditionClause{}, false
	}
	count, tail, ok := parseLeadingCount(rest)
	if !ok || count.Comparison != ConditionComparisonAtLeast {
		return ConditionClause{}, false
	}
	if !tokenWordsEqual(tail, "life", "this", "turn") {
		return ConditionClause{}, false
	}
	return ConditionClause{Predicate: ConditionPredicateControllerGainedLifeThisTurnAtLeast, Threshold: count.Value}, true
}

// recognizeControllerDesignationCondition matches an intervening-if body that
// tests whether the controller currently holds a player designation: the
// monarch (CR 720), the initiative (CR 720/dungeon), or the city's blessing
// (CR 702.131 ascend). These are live single-player game-state predicates that
// the runtime evaluates against the ability controller's designation flags.
func recognizeControllerDesignationCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	switch {
	case tokenWordsEqual(body, "you're", "the", "monarch"),
		tokenWordsEqual(body, "you", "are", "the", "monarch"):
		return ConditionClause{Predicate: ConditionPredicateControllerIsMonarch}, true
	case tokenWordsEqual(body, "you", "have", "the", "initiative"):
		return ConditionClause{Predicate: ConditionPredicateControllerHasInitiative}, true
	case tokenWordsEqual(body, "you", "have", "the", "city's", "blessing"):
		return ConditionClause{Predicate: ConditionPredicateControllerHasCityBlessing}, true
	default:
		return ConditionClause{}, false
	}
}

func recognizeGraveyardCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if rest, ok := cutTokenPrefix(body, "an", "opponent", "has"); ok {
		if count, tail, ok := parseLeadingCount(rest); ok &&
			count.Comparison == ConditionComparisonAtLeast &&
			tokenWordsEqual(tail, "cards", "in", "their", "graveyard") {
			return ConditionClause{Predicate: ConditionPredicateAnyOpponentGraveyardCardCountAtLeast, Threshold: count.Value}, true
		}
		return ConditionClause{}, false
	}
	rest := body
	if trimmed, ok := cutTokenPrefix(body, "there", "are"); ok {
		rest = trimmed
	}
	count, tail, ok := parseLeadingCount(rest)
	if !ok || count.Comparison != ConditionComparisonAtLeast {
		return ConditionClause{}, false
	}
	switch {
	case tokenWordsEqual(tail, "cards", "in", "your", "graveyard"),
		tokenWordsEqual(tail, "cards", "are", "in", "your", "graveyard"):
		return ConditionClause{Predicate: ConditionPredicateGraveyardCardCountAtLeast, Threshold: count.Value}, true
	case tokenWordsEqual(tail, "permanent", "cards", "in", "your", "graveyard"),
		tokenWordsEqual(tail, "permanent", "cards", "are", "in", "your", "graveyard"):
		return ConditionClause{Predicate: ConditionPredicateGraveyardPermanentCardCountAtLeast, Threshold: count.Value}, true
	case tokenWordsEqual(tail, "mana", "values", "among", "cards", "in", "your", "graveyard"):
		return ConditionClause{Predicate: ConditionPredicateGraveyardManaValueCountAtLeast, Threshold: count.Value}, true
	case tokenWordsEqual(tail, "card", "types", "among", "cards", "in", "your", "graveyard"):
		return ConditionClause{Predicate: ConditionPredicateGraveyardCardTypeCountAtLeast, Threshold: count.Value}, true
	}
	if cardType, ok := graveyardCountCardType(tail, atoms); ok {
		return ConditionClause{
			Predicate:              ConditionPredicateGraveyardCardOfTypeCountAtLeast,
			Threshold:              count.Value,
			GraveyardCountCardType: cardType,
		}, true
	}
	return ConditionClause{}, false
}

// graveyardCountCardType recognizes the tail "<card type> cards [are] in your
// graveyard" of a graveyard card-count condition filtered by a single card type
// ("twenty or more creature cards are in your graveyard", Mortal Combat). It
// fails closed when the noun is not a single recognized card type.
func graveyardCountCardType(tail []shared.Token, atoms Atoms) (TriggerCardType, bool) {
	typeTokens, ok := stripTokenSuffix(tail, "cards", "are", "in", "your", "graveyard")
	if !ok {
		typeTokens, ok = stripTokenSuffix(tail, "cards", "in", "your", "graveyard")
	}
	if !ok || len(typeTokens) != 1 {
		return TriggerCardTypeUnknown, false
	}
	cardType, ok := atoms.CardTypeAt(typeTokens[0].Span)
	if !ok {
		return TriggerCardTypeUnknown, false
	}
	mapped := triggerCardTypeFromAtom(cardType)
	if mapped == TriggerCardTypeUnknown {
		return TriggerCardTypeUnknown, false
	}
	return mapped, true
}

func recognizeCounterPlacementCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "would", "put", "one", "or", "more", "counters", "on", "a", "permanent", "or", "player") {
		return ConditionClause{Predicate: ConditionPredicateControllerCounterPlacement}, true
	}
	if tokenWordsEqual(body, "an", "effect", "would", "put", "one", "or", "more", "counters", "on", "a", "permanent", "you", "control") {
		return ConditionClause{Predicate: ConditionPredicateCounterPlacementOnControlledPermanent}, true
	}
	if tokenWordsEqual(body, "you", "would", "put", "one", "or", "more", "counters", "on", "a", "permanent", "you", "control") {
		return ConditionClause{Predicate: ConditionPredicateCounterPlacementOnControlledPermanent}, true
	}
	rest, ok := cutTokenPrefix(body, "one", "or", "more")
	if !ok {
		return ConditionClause{}, false
	}
	if clause, ok := recognizeSelfCounterPlacement(rest, body, atoms); ok {
		return clause, true
	}
	if tail, ok := stripTokenSuffix(rest, "counters", "would", "be", "put", "on", "a", "permanent", "you", "control"); ok && len(tail) > 0 {
		counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
		if !ok {
			return ConditionClause{}, false
		}
		return ConditionClause{
			Predicate: ConditionPredicateCounterPlacementOnControlledPermanent,
			Counter:   counterKind,
		}, true
	}
	if tail, ok := stripTokenSuffix(rest, "counters", "would", "be", "put", "on", "a", "creature", "you", "control"); ok && len(tail) > 0 {
		counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
		if !ok {
			return ConditionClause{}, false
		}
		return ConditionClause{
			Predicate: ConditionPredicateCounterPlacementOnControlledCreature,
			Counter:   counterKind,
		}, true
	}
	if clause, ok := recognizeControlledTypeUnionCounterPlacement(rest, body, atoms); ok {
		return clause, true
	}
	if clause, ok := recognizeControlledRecipientCounterPlacement(rest, body, atoms); ok {
		return clause, true
	}
	tail, ok := stripTokenSuffix(rest, "counters", "would", "be", "put", "on", "a", "creature")
	if !ok || len(tail) == 0 {
		return ConditionClause{}, false
	}
	counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateCounterPlacementOnAnyCreature,
		Counter:   counterKind,
	}, true
}

// recognizeSelfCounterPlacement recognizes the self recipient of a
// counter-placement replacement, as on Mowu, Loyal Companion ("If one or more
// +1/+1 counters would be put on Mowu, that many plus one +1/+1 counters are put
// on it instead."). The recipient after "would be put on" must be a self
// reference (the card's own name, "this creature", or "it"). It fails closed
// when the counter kind is unknown.
func recognizeSelfCounterPlacement(rest, body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	split := tokenSubsequenceIndex(rest, "counters", "would", "be", "put", "on")
	if split < 1 {
		return ConditionClause{}, false
	}
	recipient := rest[split+5:]
	if len(recipient) == 0 || !costSelfReference(recipient, atoms, true) {
		return ConditionClause{}, false
	}
	counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateCounterPlacementOnSelf,
		Counter:   counterKind,
	}, true
}

// recognizeControlledTypeUnionCounterPlacement recognizes the type-restricted
// controlled-permanent recipient of a counter-placement replacement, as on
// Ozolith, the Shattered Spire ("... would be put on an artifact or creature you
// control ..."). It accepts an "or"-joined card-type union ("artifact",
// "artifact or creature") before "you control"; the plain "a permanent you
// control" and "a creature you control" forms are handled by their own branches.
func recognizeControlledTypeUnionCounterPlacement(rest, body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	split := tokenSubsequenceIndex(rest, "counters", "would", "be", "put", "on")
	if split < 1 {
		return ConditionClause{}, false
	}
	recipient := rest[split+5:]
	inner, ok := stripTokenSuffix(recipient, "you", "control")
	if !ok || len(inner) < 2 {
		return ConditionClause{}, false
	}
	if !equalWord(inner[0], "a") && !equalWord(inner[0], "an") {
		return ConditionClause{}, false
	}
	cardTypes, ok := parseGraveyardRedirectSubjectTypes(inner[1:], atoms)
	if !ok {
		return ConditionClause{}, false
	}
	counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:                ConditionPredicateCounterPlacementOnControlledPermanent,
		Counter:                  counterKind,
		CounterRecipientTypesAny: cardTypes,
	}, true
}

// recognizeControlledRecipientCounterPlacement recognizes a counter-placement
// replacement whose controlled-permanent recipient excludes the source
// permanent ("another creature you control", Benevolent Hydra). The plain "a
// permanent you control", "a creature you control", and card-type union
// recipients are handled by their own branches.
func recognizeControlledRecipientCounterPlacement(rest, body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	split := tokenSubsequenceIndex(rest, "counters", "would", "be", "put", "on")
	if split < 1 {
		return ConditionClause{}, false
	}
	recipient := rest[split+5:]
	inner, ok := stripTokenSuffix(recipient, "you", "control")
	if !ok || len(inner) < 2 {
		return ConditionClause{}, false
	}
	if !equalWord(inner[0], "another") {
		return ConditionClause{}, false
	}
	noun := inner[1:]
	counterKind, ok := conditionCounterAtom(shared.SpanOf(body), atoms)
	if !ok {
		return ConditionClause{}, false
	}
	cardTypes, ok := parseGraveyardRedirectSubjectTypes(noun, atoms)
	if !ok || len(cardTypes) == 0 {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:                      ConditionPredicateCounterPlacementOnControlledPermanent,
		Counter:                        counterKind,
		CounterRecipientTypesAny:       cardTypes,
		CounterRecipientExcludesSource: true,
	}, true
}

func recognizeDamageSourceCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	var selection ConditionSelection
	rest := body
	if trimmed, ok := cutTokenPrefix(rest, "another"); ok {
		selection.ExcludeSource = true
		rest = trimmed
	} else if trimmed, ok := cutTokenPrefix(rest, "a"); ok {
		rest = trimmed
	} else if trimmed, ok := cutTokenPrefix(rest, "an"); ok {
		rest = trimmed
	} else {
		return ConditionClause{}, false
	}
	for len(rest) > 0 {
		color, ok := atoms.ColorAt(rest[0].Span)
		if !ok {
			break
		}
		selection.ColorsAny = append(selection.ColorsAny, triggerColorFromAtom(color))
		rest = rest[1:]
	}
	if trimmed, ok := cutTokenPrefix(rest, "source"); ok {
		rest = trimmed
	} else if len(rest) > 0 {
		cardType, ok := atoms.CardTypeAt(rest[0].Span)
		if !ok {
			return ConditionClause{}, false
		}
		mapped := triggerCardTypeFromAtom(cardType)
		if mapped == TriggerCardTypeUnknown {
			return ConditionClause{}, false
		}
		selection.RequiredTypes = append(selection.RequiredTypes, mapped)
		rest = rest[1:]
	} else {
		return ConditionClause{}, false
	}
	if trimmed, ok := cutTokenPrefix(rest, "you", "control"); ok {
		rest = trimmed
	} else if trimmed, ok := cutTokenPrefix(rest, "an", "opponent", "controls"); ok {
		selection.DamageSourceControllerOpponent = true
		rest = trimmed
	} else {
		selection.DamageSourceAnyController = true
	}
	trimmed, ok := cutTokenPrefix(rest, "would", "deal")
	if !ok {
		return ConditionClause{}, false
	}
	rest = trimmed
	if trimmed, ok := cutTokenPrefix(rest, "noncombat"); ok {
		selection.DamageNoncombatOnly = true
		rest = trimmed
	}
	trimmed, ok = cutTokenPrefix(rest, "damage", "to")
	if !ok {
		return ConditionClause{}, false
	}
	rest = trimmed
	switch {
	case tokenWordsEqual(rest, "a", "permanent", "or", "player"):
	case tokenWordsEqual(rest, "an", "opponent", "or", "a", "permanent", "an", "opponent", "controls"),
		tokenWordsEqual(rest, "an", "opponent", "or", "a", "permanent", "or", "planeswalker", "an", "opponent", "controls"):
		selection.DamageRecipientOpponent = true
	case tokenWordsEqual(rest, "you"):
		selection.DamageRecipientController = true
	default:
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateDamageByControlledSource,
		Selection: selection,
	}, true
}

func recognizeTokenCreationCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "an", "effect", "would", "create", "one", "or", "more", "tokens", "under", "your", "control") ||
		tokenWordsEqual(body, "one", "or", "more", "tokens", "would", "be", "created", "under", "your", "control") {
		return ConditionClause{Predicate: ConditionPredicateTokenCreationUnderController}, true
	}
	if tokenWordsEqual(body, "an", "effect", "would", "create", "one", "or", "more", "tokens") ||
		tokenWordsEqual(body, "one", "or", "more", "tokens", "would", "be", "created") {
		return ConditionClause{Predicate: ConditionPredicateTokenCreationAnyController}, true
	}
	// The passive would-create wording may carry a card-type filter ("one or
	// more artifact tokens would be created under your control"). The type word
	// is tolerated here and carried downstream by the would-create group's
	// selector, mirroring the active "you would create one or more <type>
	// tokens" form handled just below.
	if rest, ok := cutTokenPrefix(body, "one", "or", "more"); ok {
		if _, ok := stripTokenSuffix(rest, "tokens", "would", "be", "created", "under", "your", "control"); ok {
			return ConditionClause{Predicate: ConditionPredicateTokenCreationUnderController}, true
		}
		if _, ok := stripTokenSuffix(rest, "tokens", "would", "be", "created"); ok {
			return ConditionClause{Predicate: ConditionPredicateTokenCreationAnyController}, true
		}
	}
	if rest, ok := cutTokenPrefix(body, "you", "would", "create", "one", "or", "more"); ok {
		if _, ok := stripTokenSuffix(rest, "tokens"); ok {
			return ConditionClause{Predicate: ConditionPredicateTokenCreationUnderController}, true
		}
	}
	if tokenWordsEqual(body, "you", "created", "a", "token", "this", "turn") {
		return ConditionClause{Predicate: ConditionPredicateCreatedTokenThisTurn}, true
	}
	if _, ok := cutTokenPrefix(body, "you", "would", "create", "a"); ok {
		return ConditionClause{Predicate: ConditionPredicateControllerWouldCreateNamedToken}, true
	}
	return ConditionClause{}, false
}

// recognizeLifeGainCondition matches the intervening condition that gates a
// life-gain replacement: "you would gain life" ("If you would gain life, you
// gain twice that much life instead.", Boon Reflection, Angel of Vitality). The
// matching replacement amount ("twice that much" / "that much plus N") is
// recognized on the gain-life effect.
func recognizeLifeGainCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "would", "gain", "life") {
		return ConditionClause{Predicate: ConditionPredicateControllerLifeGain}, true
	}
	return ConditionClause{}, false
}

// recognizeLifeLossCondition matches the intervening condition that gates a
// life-loss replacement: "an opponent would lose life during your turn" (the
// controller's-turn-gated opponent form, Bloodletter of Aclazotz), "an opponent
// would lose life" (any time), and "a player would lose life" (any player). The
// matching replacement amount ("twice that much" / "that much plus N") is
// recognized on the accompanying lose-life effect.
func recognizeLifeLossCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "an", "opponent", "would", "lose", "life", "during", "your", "turn") {
		return ConditionClause{Predicate: ConditionPredicateOpponentLifeLossDuringControllerTurn}, true
	}
	if tokenWordsEqual(body, "an", "opponent", "would", "lose", "life") {
		return ConditionClause{Predicate: ConditionPredicateOpponentLifeLoss}, true
	}
	if tokenWordsEqual(body, "a", "player", "would", "lose", "life") {
		return ConditionClause{Predicate: ConditionPredicateAnyPlayerLifeLoss}, true
	}
	return ConditionClause{}, false
}

// recognizeDrawFromEmptyLibraryCondition matches the intervening condition that
// gates the draw-from-empty-library win replacement: "you would draw a card
// while your library has no cards in it" (Laboratory Maniac, Jace, Wielder of
// Mysteries). The matching replacement result ("you win the game instead") is
// recognized separately by parseDrawEmptyLibraryWinReplacement.
func recognizeDrawFromEmptyLibraryCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body,
		"you", "would", "draw", "a", "card",
		"while", "your", "library", "has", "no", "cards", "in", "it") {
		return ConditionClause{Predicate: ConditionPredicateWouldDrawFromEmptyLibrary}, true
	}
	return ConditionClause{}, false
}

// recognizeDrawCardReplacementCondition matches the intervening condition that
// gates the draw-doubling replacement: the plain "you would draw a card"
// (Thought Reflection) and the draw-step exception form "you would draw a card
// except the first one you draw in each of your draw steps" (Teferi's Ageless
// Insight). The matching replacement result ("draw two cards instead") is
// recognized separately by parseDrawDoublingReplacement.
func recognizeDrawCardReplacementCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "would", "draw", "a", "card") {
		return ConditionClause{Predicate: ConditionPredicateWouldDrawCard}, true
	}
	if tokenWordsEqual(body,
		"you", "would", "draw", "a", "card",
		"except", "the", "first", "one", "you", "draw",
		"in", "each", "of", "your", "draw", "steps") {
		return ConditionClause{Predicate: ConditionPredicateWouldDrawCardExceptFirstInDrawStep}, true
	}
	return ConditionClause{}, false
}

// recognizeControlComparisonCondition matches a cross-player control-count
// comparison: "<subject> controls more|fewer <selection> than <reference>",
// where subject and reference each name the controller ("you") or an opponent.
// It fails closed unless exactly one side is the controller and the other an
// opponent scope, so the comparison has a well-defined direction.
func recognizeControlComparisonCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	thanIdx := tokenWordIndex(body, "than")
	if thanIdx <= 0 || thanIdx == len(body)-1 {
		return ConditionClause{}, false
	}
	leftScope, afterScope, ok := cutComparisonSubjectScope(body[:thanIdx])
	if !ok {
		return ConditionClause{}, false
	}
	greater, nounTokens, ok := cutComparisonDirection(afterScope)
	if !ok || len(nounTokens) == 0 {
		return ConditionClause{}, false
	}
	rightScope, ok := comparisonReferenceScope(body[thanIdx+1:])
	if !ok || !validComparisonScopes(leftScope, rightScope) {
		return ConditionClause{}, false
	}
	selection, ok := parseConditionSelection(nounTokens, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate: ConditionPredicateControlComparison,
		Selection: selection,
		ControlComparison: ConditionControlComparison{
			LeftScope:  leftScope,
			RightScope: rightScope,
			Greater:    greater,
		},
	}, true
}

// cutComparisonSubjectScope reads the subject player scope opening a control
// comparison: "you control" (controller), "an opponent controls" (opponent), or
// "that player controls" (the triggering event's player).
func cutComparisonSubjectScope(tokens []shared.Token) (ConditionControlScope, []shared.Token, bool) {
	if rest, ok := cutTokenPrefix(tokens, "you", "control"); ok {
		return ConditionControlScopeController, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "an", "opponent", "controls"); ok {
		return ConditionControlScopeAnyOpponent, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "that", "player", "controls"); ok {
		return ConditionControlScopeTriggeringPlayer, rest, true
	}
	return ConditionControlScopeController, nil, false
}

// cutComparisonDirection reads the comparison direction word: "more" (greater)
// or "fewer"/"less" (lesser).
func cutComparisonDirection(tokens []shared.Token) (bool, []shared.Token, bool) {
	if rest, ok := cutTokenPrefix(tokens, "more"); ok {
		return true, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "fewer"); ok {
		return false, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "less"); ok {
		return false, rest, true
	}
	return false, nil, false
}

// comparisonReferenceScope reads the "than" reference player scope, which must
// consume every reference token: "you", "an"/"any opponent", "each opponent",
// or "that player" (the triggering event's player).
func comparisonReferenceScope(tokens []shared.Token) (ConditionControlScope, bool) {
	switch {
	case tokenWordsEqual(tokens, "you"):
		return ConditionControlScopeController, true
	case tokenWordsEqual(tokens, "an", "opponent"), tokenWordsEqual(tokens, "any", "opponent"):
		return ConditionControlScopeAnyOpponent, true
	case tokenWordsEqual(tokens, "each", "opponent"):
		return ConditionControlScopeEachOpponent, true
	case tokenWordsEqual(tokens, "that", "player"):
		return ConditionControlScopeTriggeringPlayer, true
	default:
		return ConditionControlScopeController, false
	}
}

// validComparisonScopes requires exactly one side to be the controller so the
// comparison contrasts the controller against an opponent scope.
func validComparisonScopes(left, right ConditionControlScope) bool {
	return (left == ConditionControlScopeController) != (right == ConditionControlScopeController)
}

// recognizeGraveyardControlsCondition matches the Incarnation-cycle condition
// "this card/creature is in your graveyard and <controls condition>" (Anger,
// Brawn, Filth, Valor, Wonder). The leading clause reports that the static
// ability functions from the graveyard; the trailing clause is delegated to
// recognizeControlsCondition for the accompanying "you control a <land>"
// requirement, which becomes the runtime condition.
func recognizeGraveyardControlsCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "this", "card", "is", "in", "your", "graveyard", "and")
	if !ok {
		rest, ok = cutTokenPrefix(body, "this", "creature", "is", "in", "your", "graveyard", "and")
		if !ok {
			return ConditionClause{}, false
		}
	}
	clause, ok := recognizeControlsCondition(rest, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	clause.SourceInGraveyard = true
	return clause, true
}

func recognizeControlsCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	scope, rest, ok := cutControlScope(body)
	if !ok {
		return ConditionClause{}, false
	}
	determiner, ok := parseControlsDeterminer(rest)
	if !ok {
		return ConditionClause{}, false
	}
	count, exclude, tail := determiner.Count, determiner.Exclude, determiner.Rest
	if scope == ConditionControlScopeController &&
		count.Comparison == ConditionComparisonAtLeast &&
		tokenWordsEqual(tail, "creatures", "with", "different", "powers") {
		return ConditionClause{
			Predicate: ConditionPredicateCreaturePowerDiversityAtLeast,
			Threshold: count.Value,
		}, true
	}
	nameDiversity := false
	if count.Comparison == ConditionComparisonAtLeast {
		if stripped, ok := stripTokenSuffix(tail, "with", "different", "names"); ok && len(stripped) > 0 {
			tail = stripped
			nameDiversity = true
		}
	}
	selection, ok := parseConditionSelection(tail, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	selection.ExcludeSource = selection.ExcludeSource || exclude
	if nameDiversity {
		selection.MatchDistinctNamesAtLeast = true
		selection.DistinctNamesAtLeast = count.Value
	}
	return ConditionClause{
		Predicate:    ConditionPredicateControls,
		Scope:        scope,
		Comparison:   count.Comparison,
		CompareValue: count.Value,
		Selection:    selection,
	}, true
}

// recognizeTotalPowerCondition matches "<selection> you control have total
// power <n> or greater", a collective-power predicate (the "Formidable" ability
// word). The selected permanents the controller controls must have a combined
// power of at least the threshold. Only the controller scope is recognized.
func recognizeTotalPowerCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	haveIndex := tokenWordIndex(body, "have")
	if haveIndex <= 0 {
		return ConditionClause{}, false
	}
	rest, ok := cutTokenPrefix(body[haveIndex+1:], "total", "power")
	if !ok || len(rest) != 3 {
		return ConditionClause{}, false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok || !equalWord(rest[1], "or") || !equalWord(rest[2], "greater") {
		return ConditionClause{}, false
	}
	nounTokens, ok := stripTokenSuffix(body[:haveIndex], "you", "control")
	if !ok || len(nounTokens) == 0 {
		return ConditionClause{}, false
	}
	selection, ok := parseConditionSelection(nounTokens, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	selection.TotalPowerAtLeast = value
	selection.MatchTotalPowerAtLeast = true
	return ConditionClause{
		Predicate:  ConditionPredicateControls,
		Scope:      ConditionControlScopeController,
		Comparison: ConditionComparisonNone,
		Selection:  selection,
	}, true
}

// recognizeControlsNamedCondition matches a "you control" gate whose objects
// are named permanents rather than card types ("If you control an Urza's Mine
// and an Urza's Tower, ..."; the Urza tron lands). It splits the noun list on
// "and", strips each segment's "a"/"an" determiner, and records the remaining
// tokens as a literal card name. A segment must begin with a capitalized word
// and must not parse as a typed condition selection, so type-based "you control
// a creature" gates fall through to recognizeControlsCondition. Name matching is
// normalized downstream, so the printed Oracle spelling matches the card name.
func recognizeControlsNamedCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "you", "control")
	if !ok {
		return ConditionClause{}, false
	}
	segments := splitTokensOnWord(rest, "and")
	if len(segments) == 0 {
		return ConditionClause{}, false
	}
	names := make([]string, 0, len(segments))
	for _, segment := range segments {
		if trimmed, cut := cutTokenPrefix(segment, "a"); cut {
			segment = trimmed
		} else if trimmed, cut := cutTokenPrefix(segment, "an"); cut {
			segment = trimmed
		}
		if !controlledNamePhrase(segment, atoms) {
			return ConditionClause{}, false
		}
		names = append(names, joinTokens(segment))
	}
	return ConditionClause{
		Predicate:       ConditionPredicateControllerControlsNamed,
		ControlledNames: names,
	}, true
}

// controlledNamePhrase reports whether tokens name a specific card (a proper
// noun) rather than a card type or subtype. It requires a leading capitalized
// word and rejects any phrase that parses as a typed condition selection.
func controlledNamePhrase(tokens []shared.Token, atoms Atoms) bool {
	if len(tokens) == 0 {
		return false
	}
	first := tokens[0].Text
	if first == "" || first[0] < 'A' || first[0] > 'Z' {
		return false
	}
	if _, ok := parseConditionSelection(tokens, atoms); ok {
		return false
	}
	return true
}

// splitTokensOnWord splits tokens into the segments separated by the given
// connector word, dropping empty segments.
func splitTokensOnWord(tokens []shared.Token, word string) [][]shared.Token {
	var segments [][]shared.Token
	start := 0
	for i := range tokens {
		if equalWord(tokens[i], word) {
			if i > start {
				segments = append(segments, tokens[start:i])
			}
			start = i + 1
		}
	}
	if start < len(tokens) {
		segments = append(segments, tokens[start:])
	}
	return segments
}

func recognizeSourceDeathCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	for _, candidate := range []struct {
		suffix    []string
		predicate ConditionPredicateKind
	}{
		{[]string{"would", "die"}, ConditionPredicateSourceWouldDie},
		{[]string{"would", "be", "put", "into", "a", "graveyard", "from", "anywhere"}, ConditionPredicateSourceWouldGoToGraveyard},
	} {
		subject, ok := stripTokenSuffix(body, candidate.suffix...)
		if !ok || len(subject) == 0 {
			continue
		}
		subjectSpan := shared.SpanOf(subject)
		return ConditionClause{
			Predicate:      candidate.predicate,
			SubjectSpan:    subjectSpan,
			HasSubjectSpan: true,
			SubjectRefID:   atoms.ReferenceIDAt(subjectSpan),
		}, true
	}
	return ConditionClause{}, false
}

// recognizeCardToGraveyardReplacementCondition recognizes the global
// graveyard-redirect replacement condition "[a/an] <subject> would be put into
// [a/your/an opponent's] graveyard [from anywhere]" (Leyline of the Void,
// Samurai of the Pale Curtain, Dryad Militant, Rest in Peace, Dauthi
// Voidwalker). It is distinct from recognizeSourceDeathCondition, whose subject
// binds the ability's own source. The subject "a permanent" leaves only the
// battlefield; "a card"/"a card or token"/typed "<types> card" subjects come
// from anywhere.
func recognizeCardToGraveyardReplacementCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	pivot := tokenSubsequenceIndex(body, "would", "be", "put", "into")
	if pivot <= 0 {
		return ConditionClause{}, false
	}
	destination, ok := parseGraveyardRedirectDestination(body[pivot+4:])
	if !ok {
		return ConditionClause{}, false
	}
	subjectTypes, subjectFromBattlefield, ok := parseGraveyardRedirectSubject(body[:pivot], atoms)
	if !ok {
		return ConditionClause{}, false
	}
	fromBattlefieldOnly := subjectFromBattlefield || destination.fromBattlefield
	if !fromBattlefieldOnly && !destination.fromAnywhere {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:                    ConditionPredicateCardWouldGoToGraveyard,
		GraveyardRedirectScope:       destination.scope,
		GraveyardSubjectTypesAny:     subjectTypes,
		GraveyardFromBattlefieldOnly: fromBattlefieldOnly,
	}, true
}

// recognizeCreatureWouldDieReplacementCondition recognizes the "would die"
// graveyard-redirect replacement condition "[a/an] <type> [you control / an
// opponent controls] would die" (Stone of Erech, Misery's Shadow, Gisa,
// Liesa, Nemata). "X would die" means "X would be put into a graveyard from the
// battlefield" (CR 700.4), so it always restricts the source zone to the
// battlefield. The optional control qualifier watches who controls the dying
// permanent, distinct from the owner-scoped "would be put into a graveyard"
// forms. It fails closed on self ("this creature"), exclude-self ("another
// creature"), duration ("this turn"), and counter-conditioned subjects, leaving
// those to other recognizers or as unsupported.
func recognizeCreatureWouldDieReplacementCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	subject, ok := stripTokenSuffix(body, "would", "die")
	if !ok || len(subject) == 0 {
		return ConditionClause{}, false
	}
	controlScope := GraveyardRedirectControlScopeAny
	if trimmed, trimmedOK := stripTokenSuffix(subject, "an", "opponent", "controls"); trimmedOK {
		subject = trimmed
		controlScope = GraveyardRedirectControlScopeOpponent
	} else if trimmed, trimmedOK := stripTokenSuffix(subject, "you", "control"); trimmedOK {
		subject = trimmed
		controlScope = GraveyardRedirectControlScopeYou
	}
	subjectTypes, ok := parseGraveyardDeathSubject(subject, atoms)
	if !ok {
		return ConditionClause{}, false
	}
	return ConditionClause{
		Predicate:                     ConditionPredicateCardWouldGoToGraveyard,
		GraveyardSubjectTypesAny:      subjectTypes,
		GraveyardFromBattlefieldOnly:  true,
		GraveyardRedirectControlScope: controlScope,
	}, true
}

// parseGraveyardDeathSubject parses the dying-object subject of a "would die"
// graveyard-redirect condition, returning the typed card-type filter (empty for
// "a permanent"). It accepts only "a permanent" or a single permanent card-type
// noun ("a creature", "an artifact"); it fails closed on anything else.
func parseGraveyardDeathSubject(subject []shared.Token, atoms Atoms) (cardTypes []TriggerCardType, ok bool) {
	if len(subject) != 2 || (!equalWord(subject[0], "a") && !equalWord(subject[0], "an")) {
		return nil, false
	}
	if equalWord(subject[1], "permanent") {
		return nil, true
	}
	cardType, ok := atoms.CardTypeAt(subject[1].Span)
	if !ok {
		return nil, false
	}
	mapped := triggerCardTypeFromAtom(cardType)
	if mapped == TriggerCardTypeUnknown {
		return nil, false
	}
	return []TriggerCardType{mapped}, true
}

// graveyardRedirectDestination is the parsed destination phrase of a
// graveyard-redirect condition: the watched graveyard scope, whether "from
// anywhere" was present, and whether the source zone was restricted to the
// battlefield ("from the battlefield").
type graveyardRedirectDestination struct {
	scope           GraveyardRedirectScope
	fromAnywhere    bool
	fromBattlefield bool
}

// parseGraveyardRedirectDestination parses the destination phrase of a
// graveyard-redirect condition. It fails closed for unrecognized phrases.
func parseGraveyardRedirectDestination(rest []shared.Token) (graveyardRedirectDestination, bool) {
	result := graveyardRedirectDestination{}
	if trimmed, ok := stripTokenSuffix(rest, "from", "anywhere"); ok {
		rest = trimmed
		result.fromAnywhere = true
	} else if trimmed, ok := stripTokenSuffix(rest, "from", "the", "battlefield"); ok {
		rest = trimmed
		result.fromBattlefield = true
	}
	switch {
	case tokenWordsEqual(rest, "a", "graveyard"):
		result.scope = GraveyardRedirectScopeAny
	case tokenWordsEqual(rest, "your", "graveyard"):
		result.scope = GraveyardRedirectScopeYou
	case tokenWordsEqual(rest, "an", "opponent's", "graveyard"):
		result.scope = GraveyardRedirectScopeOpponent
	default:
		return graveyardRedirectDestination{}, false
	}
	return result, true
}

// parseGraveyardRedirectSubject parses the moving-object subject of a
// graveyard-redirect condition, returning the typed card-type filter (empty for
// any card) and whether the subject can only leave the battlefield ("a
// permanent"). It fails closed for unrecognized subjects.
func parseGraveyardRedirectSubject(subject []shared.Token, atoms Atoms) (cardTypes []TriggerCardType, fromBattlefieldOnly, ok bool) {
	if len(subject) < 2 || (!equalWord(subject[0], "a") && !equalWord(subject[0], "an")) {
		return nil, false, false
	}
	noun := subject[1:]
	switch {
	case tokenWordsEqual(noun, "permanent"):
		return nil, true, true
	case tokenWordsEqual(noun, "card"), tokenWordsEqual(noun, "card", "or", "token"):
		return nil, false, true
	}
	typeTokens, ok := stripTokenSuffix(noun, "card")
	if !ok || len(typeTokens) == 0 {
		return nil, false, false
	}
	subjectTypes, ok := parseGraveyardRedirectSubjectTypes(typeTokens, atoms)
	if !ok {
		return nil, false, false
	}
	return subjectTypes, false, true
}

// parseGraveyardRedirectSubjectTypes parses an "or"-joined card-type list
// ("instant or sorcery", "creature") into typed card types.
func parseGraveyardRedirectSubjectTypes(tokens []shared.Token, atoms Atoms) ([]TriggerCardType, bool) {
	var result []TriggerCardType
	for i := range tokens {
		if i%2 == 1 {
			if !equalWord(tokens[i], "or") {
				return nil, false
			}
			continue
		}
		cardType, ok := atoms.CardTypeAt(tokens[i].Span)
		if !ok {
			return nil, false
		}
		mapped := triggerCardTypeFromAtom(cardType)
		if mapped == TriggerCardTypeUnknown || slices.Contains(result, mapped) {
			return nil, false
		}
		result = append(result, mapped)
	}
	if len(result) == 0 {
		return nil, false
	}
	return result, true
}

// tokenSubsequenceIndex returns the start index of the first occurrence of the
// word subsequence in tokens, or -1 when absent.
func tokenSubsequenceIndex(tokens []shared.Token, words ...string) int {
	for start := 0; start+len(words) <= len(tokens); start++ {
		if equalWordSequence(tokens, start, words...) {
			return start
		}
	}
	return -1
}

type conditionCount struct {
	Comparison ConditionComparison `json:",omitempty"`
	Value      int                 `json:",omitempty"`
}

// parseLeadingCount parses a leading "<n> or more|fewer|less" count, returning
// the typed comparison and the remaining tokens.
func parseLeadingCount(tokens []shared.Token) (conditionCount, []shared.Token, bool) {
	if len(tokens) < 3 {
		return conditionCount{}, nil, false
	}
	value, ok := conditionNumberValue(tokens[0])
	if !ok || !equalWord(tokens[1], "or") {
		return conditionCount{}, nil, false
	}
	switch {
	case equalWord(tokens[2], "more"):
		return conditionCount{Comparison: ConditionComparisonAtLeast, Value: value}, tokens[3:], true
	case equalWord(tokens[2], "fewer"), equalWord(tokens[2], "less"):
		return conditionCount{Comparison: ConditionComparisonAtMost, Value: value}, tokens[3:], true
	default:
		return conditionCount{}, nil, false
	}
}

// controlsDeterminer holds the parsed opening determiner of a controlled-permanent
// noun phrase.
type controlsDeterminer struct {
	Count   conditionCount `json:",omitzero"`
	Exclude bool           `json:",omitempty"`
	Rest    []shared.Token `json:"-"`
}

// parseControlsDeterminer parses the determiner that opens a controlled-permanent
// noun phrase: "a"/"an"/"another"/"no"/"<n> or more|fewer". It returns the count,
// whether the source is excluded, and the remaining noun tokens.
func parseControlsDeterminer(tokens []shared.Token) (controlsDeterminer, bool) {
	var rest []shared.Token
	count := conditionCount{Comparison: ConditionComparisonNone}
	exclude := false
	switch {
	case startsWithWord(tokens, "another"):
		exclude = true
		rest = tokens[1:]
	case startsWithWord(tokens, "a"), startsWithWord(tokens, "an"):
		rest = tokens[1:]
	case startsWithWord(tokens, "no"):
		count = conditionCount{Comparison: ConditionComparisonAtMost, Value: 0}
		rest = tokens[1:]
	default:
		parsed, tail, ok := parseLeadingCount(tokens)
		if !ok {
			return controlsDeterminer{}, false
		}
		count = parsed
		rest = tail
	}
	if trimmed, ok := cutTokenPrefix(rest, "other"); ok {
		exclude = true
		rest = trimmed
	}
	if len(rest) == 0 {
		return controlsDeterminer{}, false
	}
	return controlsDeterminer{Count: count, Exclude: exclude, Rest: rest}, true
}

// parseConditionSelection parses a permanent noun phrase into a typed selection,
// consuming card-type, subtype, color, and supertype atoms by span. It fails
// closed unless every token belongs to a recognized production.
func parseConditionSelection(tokens []shared.Token, atoms Atoms) (ConditionSelection, bool) {
	if len(tokens) == 0 {
		return ConditionSelection{}, false
	}
	var selection ConditionSelection
	// Trailing "with <qualifier>" clause: either "with power <n> or greater" or
	// "with <keyword>" (e.g. "a creature with flying").
	if idx := tokenWordIndex(tokens, "with"); idx >= 0 {
		qualifier := tokens[idx+1:]
		if !parseConditionPowerQualifier(qualifier, &selection) &&
			!parseConditionKeywordQualifier(qualifier, &selection) {
			return ConditionSelection{}, false
		}
		tokens = tokens[:idx]
	}
	if len(tokens) == 0 {
		return ConditionSelection{}, false
	}
	// Leading tapped/untapped state.
	switch {
	case equalWord(tokens[0], "tapped"):
		selection.Tapped = ConditionTappedTrue
		tokens = tokens[1:]
	case equalWord(tokens[0], "untapped"):
		selection.Tapped = ConditionTappedFalse
		tokens = tokens[1:]
	default:
	}
	// Leading supertypes (basic/snow/legendary).
	for len(tokens) > 0 {
		supertype, ok := conditionSupertypeAtom(tokens[0].Span, atoms)
		if !ok {
			break
		}
		selection.Supertypes = append(selection.Supertypes, supertype)
		tokens = tokens[1:]
	}
	if len(tokens) == 0 {
		return selection, false
	}
	return parseConditionNoun(tokens, atoms, selection)
}

func parseConditionNoun(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if orIndex := tokenWordIndex(tokens, "or"); orIndex >= 0 {
		return parseConditionAlternativeNoun(tokens, orIndex, atoms, selection)
	}
	// Color-qualified "<colors> creature|permanent".
	if clause, ok := parseConditionColorQualified(tokens, atoms, selection); ok {
		return clause, true
	}
	// One or more distinct card-type words.
	cardTypes := make([]TriggerCardType, 0, len(tokens))
	allTypes := true
	for _, token := range tokens {
		cardType, ok := atoms.CardTypeAt(token.Span)
		if !ok {
			allTypes = false
			break
		}
		mapped := triggerCardTypeFromAtom(cardType)
		if slices.Contains(cardTypes, mapped) {
			return ConditionSelection{}, false
		}
		cardTypes = append(cardTypes, mapped)
	}
	if allTypes {
		selection.RequiredTypes = append(selection.RequiredTypes, cardTypes...)
		return selection, len(selection.RequiredTypes) > 0
	}
	// A bare permanent (no required type), e.g. "permanent" or "permanents".
	if tokenWordsEqual(tokens, "permanent") || tokenWordsEqual(tokens, "permanents") {
		return selection, true
	}
	// A bare token, e.g. "you control a token".
	if tokenWordsEqual(tokens, "token") || tokenWordsEqual(tokens, "tokens") {
		selection.TokenOnly = true
		return selection, true
	}
	// A subtype noun: creature, land, or "<name> planeswalker".
	return parseConditionSubtypeNoun(tokens, atoms, selection)
}

func parseConditionSubtypeNoun(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	span := shared.SpanOf(tokens)
	if subtype, ok := atoms.SubtypeAt(span); ok {
		selection.SubtypesAny = append(selection.SubtypesAny, subtype)
		return selection, true
	}
	if len(tokens) >= 2 && equalWord(tokens[len(tokens)-1], "planeswalker") {
		nameSpan := shared.SpanOf(tokens[:len(tokens)-1])
		if subtype, ok := conditionSubtypeAtom(nameSpan, atoms, CardTypePlaneswalker); ok {
			selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypePlaneswalker)
			selection.SubtypesAny = append(selection.SubtypesAny, subtype)
			return selection, true
		}
	}
	// A typed subtype noun "<name> creature", e.g. "a Griffin creature".
	if len(tokens) >= 2 &&
		(equalWord(tokens[len(tokens)-1], "creature") || equalWord(tokens[len(tokens)-1], "creatures")) {
		nameSpan := shared.SpanOf(tokens[:len(tokens)-1])
		if subtype, ok := conditionSubtypeAtom(nameSpan, atoms, CardTypeCreature); ok {
			selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeCreature)
			selection.SubtypesAny = append(selection.SubtypesAny, subtype)
			return selection, true
		}
	}
	return ConditionSelection{}, false
}

func parseConditionAlternativeNoun(tokens []shared.Token, orIndex int, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if clause, ok := parseConditionColorQualified(tokens, atoms, selection); ok {
		return clause, true
	}
	left := tokens[:orIndex]
	right := tokens[orIndex+1:]
	if trimmed, ok := cutTokenPrefix(right, "a"); ok {
		right = trimmed
	} else if trimmed, ok := cutTokenPrefix(right, "an"); ok {
		right = trimmed
	}
	// Land subtype disjunction ("a Forest or an Island") carries the Land card
	// type so the matched permanent must be a land of either basic type.
	leftLand, leftLandOK := conditionSubtypeAtom(shared.SpanOf(left), atoms, CardTypeLand)
	rightLand, rightLandOK := conditionSubtypeAtom(shared.SpanOf(right), atoms, CardTypeLand)
	if leftLandOK && rightLandOK {
		selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeLand)
		selection.SubtypesAny = append(selection.SubtypesAny, leftLand, rightLand)
		return selection, true
	}
	// Generic subtype disjunction ("another Wolf or Werewolf"). Each side names a
	// subtype of any card type and the match constrains only the subtype, exactly
	// like the single-subtype noun production, so a permanent matches if it has
	// either named subtype.
	leftSub, leftOK := atoms.SubtypeAt(shared.SpanOf(left))
	rightSub, rightOK := atoms.SubtypeAt(shared.SpanOf(right))
	if !leftOK || !rightOK {
		return ConditionSelection{}, false
	}
	selection.SubtypesAny = append(selection.SubtypesAny, leftSub, rightSub)
	return selection, true
}

// parseConditionColorQualified handles "<colors> creature(s)" and "<colors>
// permanent(s)", where colors are one or more color atoms joined by "or", or the
// "colorless"/"multicolored" qualifier.
func parseConditionColorQualified(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	if len(tokens) < 2 {
		return ConditionSelection{}, false
	}
	last := tokens[len(tokens)-1]
	colorTokens := tokens[:len(tokens)-1]
	switch {
	case equalWord(last, "creature"), equalWord(last, "creatures"):
		selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeCreature)
	case equalWord(last, "permanent"), equalWord(last, "permanents"):
	default:
		return ConditionSelection{}, false
	}
	if tokenWordsEqual(colorTokens, "colorless") {
		selection.Colorless = true
		return selection, true
	}
	if tokenWordsEqual(colorTokens, "multicolored") {
		selection.Multicolored = true
		return selection, true
	}
	for _, token := range colorTokens {
		if equalWord(token, "or") {
			continue
		}
		color, ok := atoms.ColorAt(token.Span)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, triggerColorFromAtom(color))
	}
	if len(selection.ColorsAny) == 0 {
		return ConditionSelection{}, false
	}
	return selection, true
}

func parseConditionPowerQualifier(tokens []shared.Token, selection *ConditionSelection) bool {
	rest, ok := cutTokenPrefix(tokens, "power")
	if !ok || len(rest) != 3 {
		return false
	}
	value, ok := conditionNumberValue(rest[0])
	if !ok || !equalWord(rest[1], "or") || !equalWord(rest[2], "greater") {
		return false
	}
	selection.PowerAtLeast = value
	selection.MatchPowerAtLeast = true
	return true
}

// parseConditionKeywordQualifier recognizes a single keyword name following
// "with" (e.g. "a creature with flying"). The qualifier tokens must form exactly
// one keyword name; trailing text fails closed.
func parseConditionKeywordQualifier(tokens []shared.Token, selection *ConditionSelection) bool {
	kind, length, ok := recognizeKeywordNameAt(tokens, 0)
	if !ok || length != len(tokens) {
		return false
	}
	selection.Keyword = kind
	return true
}

func cutControlScope(tokens []shared.Token) (ConditionControlScope, []shared.Token, bool) {
	if rest, ok := cutTokenPrefix(tokens, "you", "control"); ok {
		return ConditionControlScopeController, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "an", "opponent", "controls"); ok {
		return ConditionControlScopeAnyOpponent, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "your", "opponents", "control"); ok {
		return ConditionControlScopeOpponents, rest, true
	}
	if rest, ok := cutTokenPrefix(tokens, "defending", "player", "controls"); ok {
		return ConditionControlScopeDefendingPlayer, rest, true
	}
	return ConditionControlScopeController, nil, false
}

func conditionCounterAtom(span shared.Span, atoms Atoms) (ConditionCounterKind, bool) {
	kind, _, ok := atoms.CounterIn(span)
	if !ok {
		return ConditionCounterNone, false
	}
	switch kind {
	case counter.PlusOnePlusOne:
		return ConditionCounterPlusOnePlusOne, true
	case counter.MinusOneMinusOne:
		return ConditionCounterMinusOneMinusOne, true
	default:
		return ConditionCounterNone, false
	}
}

func conditionSupertypeAtom(span shared.Span, atoms Atoms) (ConditionSupertype, bool) {
	supertype, ok := atoms.SupertypeAt(span)
	if !ok {
		return ConditionSupertypeUnknown, false
	}
	switch supertype {
	case SupertypeBasic:
		return ConditionSupertypeBasic, true
	case SupertypeSnow:
		return ConditionSupertypeSnow, true
	case SupertypeLegendary:
		return ConditionSupertypeLegendary, true
	default:
		return ConditionSupertypeUnknown, false
	}
}

func conditionSubtypeAtom(span shared.Span, atoms Atoms, cardType CardType) (types.Sub, bool) {
	sub, ok := atoms.SubtypeAt(span)
	if !ok {
		return "", false
	}
	if !SubtypeMatchesCardType(sub, cardType) {
		return "", false
	}
	return sub, true
}

func conditionNumberValue(token shared.Token) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		return value, err == nil
	}
	if token.Kind == shared.Word {
		return CardinalWordValue(token.Text)
	}
	return 0, false
}

func triggerCardTypeFromAtom(cardType CardType) TriggerCardType {
	switch cardType {
	case CardTypeArtifact:
		return TriggerCardTypeArtifact
	case CardTypeBattle:
		return TriggerCardTypeBattle
	case CardTypeCreature:
		return TriggerCardTypeCreature
	case CardTypeEnchantment:
		return TriggerCardTypeEnchantment
	case CardTypeInstant:
		return TriggerCardTypeInstant
	case CardTypeLand:
		return TriggerCardTypeLand
	case CardTypePlaneswalker:
		return TriggerCardTypePlaneswalker
	case CardTypeSorcery:
		return TriggerCardTypeSorcery
	default:
		return TriggerCardTypeUnknown
	}
}

func triggerColorFromAtom(color Color) TriggerColor {
	switch color {
	case ColorWhite:
		return TriggerColorWhite
	case ColorBlue:
		return TriggerColorBlue
	case ColorBlack:
		return TriggerColorBlack
	case ColorRed:
		return TriggerColorRed
	case ColorGreen:
		return TriggerColorGreen
	default:
		return TriggerColorUnknown
	}
}

func conditionSelectionEmptyExceptType(selection ConditionSelection) bool {
	return len(selection.RequiredTypes) == 1 &&
		len(selection.Supertypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		!selection.Colorless &&
		!selection.ExcludeSource &&
		selection.Tapped == ConditionTappedAny &&
		selection.Attachment == ConditionAttachmentAny &&
		!selection.MatchPowerAtLeast
}

func tokenSuffixWord(tokens []shared.Token, word string) bool {
	return len(tokens) > 0 && equalWord(tokens[len(tokens)-1], word)
}

func tokenWordIndex(tokens []shared.Token, word string) int {
	for i := range tokens {
		if equalWord(tokens[i], word) {
			return i
		}
	}
	return -1
}
