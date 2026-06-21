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
	ConditionPredicateEventSubjectEnteredOrCastFromGraveyard           ConditionPredicateKind = "ConditionPredicateEventSubjectEnteredOrCastFromGraveyard"
	ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard ConditionPredicateKind = "ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard"
	ConditionPredicateEventSubjectHadNoCounter                         ConditionPredicateKind = "ConditionPredicateEventSubjectHadNoCounter"
	ConditionPredicateEventSubjectHadCounters                          ConditionPredicateKind = "ConditionPredicateEventSubjectHadCounters"
	ConditionPredicatePriorInstructionNotAccepted                      ConditionPredicateKind = "ConditionPredicatePriorInstructionNotAccepted"
	ConditionPredicatePriorInstructionAccepted                         ConditionPredicateKind = "ConditionPredicatePriorInstructionAccepted"
	ConditionPredicateEventPlayerDoesNotPay                            ConditionPredicateKind = "ConditionPredicateEventPlayerDoesNotPay"
	ConditionPredicateCounterPlacementOnControlledCreature             ConditionPredicateKind = "ConditionPredicateCounterPlacementOnControlledCreature"
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
)

// ConditionSelection is the source-independent permanent selection used by typed
// condition clauses. Subtype names are canonical typed identities.
type ConditionSelection struct {
	RequiredTypes     []TriggerCardType    `json:",omitempty"`
	Supertypes        []ConditionSupertype `json:",omitempty"`
	SubtypesAny       []types.Sub          `json:",omitempty"`
	ColorsAny         []TriggerColor       `json:",omitempty"`
	Colorless         bool                 `json:",omitempty"`
	Multicolored      bool                 `json:",omitempty"`
	TokenOnly         bool                 `json:",omitempty"`
	ExcludeSource     bool                 `json:",omitempty"`
	Tapped            ConditionTappedState `json:",omitempty"`
	CombatState       ConditionCombatState `json:",omitempty"`
	Keyword           KeywordKind          `json:",omitempty"`
	PowerAtLeast      int                  `json:",omitempty"`
	MatchPowerAtLeast bool                 `json:",omitempty"`
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

	// AnyCounter requires the matched permanent to carry at least one counter of
	// any kind ("if this permanent has counters on it"). It is the kind-agnostic
	// companion to a named-counter requirement.
	AnyCounter bool `json:",omitempty"`
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

	// CounterRecipientTypesAny restricts a
	// ConditionPredicateCounterPlacementOnControlledPermanent clause to a
	// controlled permanent that has at least one of the listed card types ("an
	// artifact or creature you control", Ozolith, the Shattered Spire). It is
	// empty for the unrestricted "a permanent you control" form.
	CounterRecipientTypesAny []TriggerCardType `json:",omitempty"`
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

// isReflexiveWhenYouDoIntro reports whether the tokens at index open the closed
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
		recognizeDestroyedThisWayCondition,
		recognizeEventSubjectCondition,
		recognizeSourceStateCondition,
		recognizeAttachedCreatureStateCondition,
		recognizeSourceCounterStateCondition,
		recognizeControllerResourceCondition,
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
		recognizeCardToGraveyardReplacementCondition,
		recognizeSourceDeathCondition,
		recognizeTargetColorCondition,
		recognizeDrawFromEmptyLibraryCondition,
		recognizeDrawCardReplacementCondition,
		recognizeCastTimingCondition,
	} {
		if clause, ok := recognize(body, atoms); ok {
			return clause, true
		}
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

func recognizePriorInstructionCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "don't") {
		return ConditionClause{Predicate: ConditionPredicatePriorInstructionNotAccepted}, true
	}
	if tokenWordsEqual(body, "you", "do") {
		return ConditionClause{Predicate: ConditionPredicatePriorInstructionAccepted}, true
	}
	return ConditionClause{}, false
}

// recognizeDestroyedThisWayCondition matches the resolving success gate "a
// <permanent noun> is destroyed this way" (and the plural "are" form) that
// follows a preceding optional destroy effect, as in Noxious Gearhulk's "you may
// destroy another target creature. If a creature is destroyed this way, you gain
// life equal to its toughness." It is the outcome-worded equivalent of "if you
// do": the gate holds exactly when the prior destroy actually moved a permanent
// to the graveyard, so it maps to the same prior-instruction success predicate.
// The noun is the descriptive type of what the prior clause destroyed and carries
// no selection of its own. It fails closed on any other wording.
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
	return ConditionClause{Predicate: ConditionPredicatePriorInstructionAccepted}, true
}

// recognizeCastTimingCondition handles the Addendum cast-timing gate "you cast
// this spell during your main phase", which restricts the gated effect to
// spells cast while their controller is the active player in a main phase.
func recognizeCastTimingCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "cast", "this", "spell", "during", "your", "main", "phase") {
		return ConditionClause{Predicate: ConditionPredicateCastDuringControllerMainPhase}, true
	}
	return ConditionClause{}, false
}

func recognizeEventSubjectCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "cast", "it") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasCastByController}, true
	}
	if tokenWordsEqual(body, "it", "was", "kicked") {
		return ConditionClause{Predicate: ConditionPredicateEventSubjectWasKicked}, true
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

// recognizeEnteredOrCastFromGraveyardCondition handles the enters-the-battlefield
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
// enters, draw a card if its power is 3 or greater."). The possessive "its"
// binds the event permanent, so the recognized clause carries a power-at-least
// selection matched against that object.
func recognizeEventSubjectPowerState(body []shared.Token) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "its", "power", "is")
	if !ok {
		return ConditionClause{}, false
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

// recognizeEventSubjectCounterCondition handles "it had no <counter> counters"
// and the optional trailing "on it".
func recognizeEventSubjectCounterCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
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

// inspect the source permanent.
func recognizeSourceStateCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "this")
	if !ok {
		return ConditionClause{}, false
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
// "equipped creature is <state>" / "enchanted creature is <state>" used by
// Equipment and Auras ("As long as equipped creature is legendary, it has
// hexproof."). The subject names the permanent the source is attached to; the
// state is a supertype (e.g. "legendary"), a card type, or a tap/combat state.
// It binds the attached object so a static keyword grant can gate on the
// equipped or enchanted creature's own characteristics.
func recognizeAttachedCreatureStateCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	rest, ok := cutTokenPrefix(body, "equipped", "creature", "is")
	if !ok {
		rest, ok = cutTokenPrefix(body, "enchanted", "creature", "is")
		if !ok {
			return ConditionClause{}, false
		}
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

// applyAttachedCreatureState fills the selection from the state words following
// "equipped/enchanted creature is ...". A bare supertype ("legendary") sets the
// supertype filter; other states fall through to the shared source-state vocab
// ("a <type>", tapped/untapped, attacking/blocking).
func applyAttachedCreatureState(stateTokens []shared.Token, atoms Atoms, selection *ConditionSelection) bool {
	if supertypes, ok := conditionStateSupertypes(stateTokens, atoms); ok {
		selection.Supertypes = append(selection.Supertypes, supertypes...)
		return true
	}
	return applySourceState(stateTokens, atoms, selection)
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

// cutSourceSubjectTokens consumes a leading source self-subject — the card's own
// name or a "this <type>" phrase — and returns the remaining state tokens. It
// fails closed when the body does not begin with a recognized source subject.
func cutSourceSubjectTokens(body []shared.Token, atoms Atoms) ([]shared.Token, bool) {
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
		if count, tail, ok := parseLeadingCount(rest); ok && count.Comparison == ConditionComparisonAtLeast {
			switch {
			case tokenWordsEqual(tail, "cards", "in", "hand"):
				return ConditionClause{Predicate: ConditionPredicateControllerHandSizeAtLeast, Threshold: count.Value}, true
			case tokenWordsEqual(tail, "life"):
				return ConditionClause{Predicate: ConditionPredicateControllerLifeAtLeast, Threshold: count.Value}, true
			case tokenWordsEqual(tail, "opponents"):
				return ConditionClause{Predicate: ConditionPredicateOpponentCountAtLeast, Threshold: count.Value}, true
			}
		}
		// "you have exactly <n> cards in hand" is an equality on hand size, e.g.
		// "Activate only if you have exactly seven cards in hand".
		if exact, ok := cutTokenPrefix(rest, "exactly"); ok && len(exact) >= 1 {
			if value, ok := conditionNumberValue(exact[0]); ok &&
				tokenWordsEqual(exact[1:], "cards", "in", "hand") {
				return ConditionClause{Predicate: ConditionPredicateControllerHandSizeExactly, Threshold: value}, true
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

func recognizeGraveyardCondition(body []shared.Token, _ Atoms) (ConditionClause, bool) {
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
	case tokenWordsEqual(tail, "card", "types", "among", "cards", "in", "your", "graveyard"):
		return ConditionClause{Predicate: ConditionPredicateGraveyardCardTypeCountAtLeast, Threshold: count.Value}, true
	default:
		return ConditionClause{}, false
	}
}

func recognizeCounterPlacementCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	if tokenWordsEqual(body, "you", "would", "put", "one", "or", "more", "counters", "on", "a", "permanent", "or", "player") {
		return ConditionClause{Predicate: ConditionPredicateControllerCounterPlacement}, true
	}
	if tokenWordsEqual(body, "an", "effect", "would", "put", "one", "or", "more", "counters", "on", "a", "permanent", "you", "control") {
		return ConditionClause{Predicate: ConditionPredicateCounterPlacementOnControlledPermanent}, true
	}
	rest, ok := cutTokenPrefix(body, "one", "or", "more")
	if !ok {
		return ConditionClause{}, false
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

func recognizeDamageSourceCondition(body []shared.Token, atoms Atoms) (ConditionClause, bool) {
	var selection ConditionSelection
	rest := body
	if trimmed, ok := cutTokenPrefix(rest, "another"); ok {
		selection.ExcludeSource = true
		rest = trimmed
	} else if trimmed, ok := cutTokenPrefix(rest, "a"); ok {
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
	} else if trimmed, ok := cutTokenPrefix(rest, "creature"); ok {
		selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeCreature)
		rest = trimmed
	} else {
		return ConditionClause{}, false
	}
	if trimmed, ok := cutTokenPrefix(rest, "you", "control"); ok {
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
	leftSub, leftOK := conditionSubtypeAtom(shared.SpanOf(left), atoms, CardTypeLand)
	rightSub, rightOK := conditionSubtypeAtom(shared.SpanOf(right), atoms, CardTypeLand)
	if !leftOK || !rightOK {
		return ConditionSelection{}, false
	}
	selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeLand)
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
