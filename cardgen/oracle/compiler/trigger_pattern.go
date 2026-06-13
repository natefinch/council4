package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// TriggerEvent identifies a representable rules event without depending on
// runtime game values.
type TriggerEvent uint8

// Trigger event families recognized by the semantic compiler.
const (
	TriggerEventUnknown TriggerEvent = iota
	TriggerEventSpellCast
	TriggerEventPermanentEnteredBattlefield
	TriggerEventPermanentDied
	TriggerEventZoneChanged
	TriggerEventCountersAdded
	TriggerEventDamageDealt
	TriggerEventCardDrawn
	TriggerEventAttackerDeclared
	TriggerEventBlockerDeclared
	TriggerEventCardDiscarded
	TriggerEventCycled
	TriggerEventBeginningOfStep
	TriggerEventLifeGained
	TriggerEventLifeLost
	TriggerEventPermanentTapped
	TriggerEventPermanentUntapped
	TriggerEventPermanentTurnedFaceUp
	TriggerEventPermanentSacrificed
	TriggerEventScry
	TriggerEventSurveil
	TriggerEventAbilityActivated
	TriggerEventObjectBecameTarget
	TriggerEventPermanentMutated
	TriggerEventAttackerBecameBlocked
)

// TriggerSourceRelation identifies the event object's relationship to the
// ability source.
type TriggerSourceRelation uint8

// Trigger source relations.
const (
	TriggerSourceAny TriggerSourceRelation = iota
	TriggerSourceSelf
	TriggerSourceAttachedPermanent
)

// TriggerSubject identifies the event permanent role used for source and
// controller matching.
type TriggerSubject uint8

// Trigger subject roles.
const (
	TriggerSubjectDefault TriggerSubject = iota
	TriggerSubjectPermanent
	TriggerSubjectBlockedAttacker
	TriggerSubjectDamageSource
)

// TriggerPlayerRelation identifies an affected player's relationship to the
// ability controller.
type TriggerPlayerRelation uint8

// Trigger player relations.
const (
	TriggerPlayerAny TriggerPlayerRelation = iota
	TriggerPlayerYou
	TriggerPlayerOpponent
)

// TriggerZone identifies a zone used by a trigger pattern.
type TriggerZone uint8

// Trigger zones.
const (
	TriggerZoneNone TriggerZone = iota
	TriggerZoneGraveyard
	TriggerZoneBattlefield
	TriggerZoneHand
	TriggerZoneExile
	TriggerZoneLibrary
	TriggerZoneStack
	TriggerZoneCommand
)

// TriggerStep identifies a phase or step boundary used by a trigger pattern.
type TriggerStep uint8

// Trigger steps.
const (
	TriggerStepNone TriggerStep = iota
	TriggerStepUpkeep
	TriggerStepDraw
	TriggerStepBeginningOfCombat
	TriggerStepEndOfCombat
	TriggerStepEnd
	TriggerStepPrecombatMain
	TriggerStepPostcombatMain
)

// TriggerCombatQualifier identifies a combat-specific event restriction.
type TriggerCombatQualifier uint8

// Trigger combat qualifiers.
const (
	TriggerCombatAny TriggerCombatQualifier = iota
	TriggerCombatDamage
	TriggerNonCombatDamage
)

// TriggerAttackRecipient identifies what an attacker was declared against.
type TriggerAttackRecipient uint8

// Trigger attack recipient values are flags so exact recipient unions remain
// representable.
const (
	TriggerAttackRecipientAny    TriggerAttackRecipient = 0
	TriggerAttackRecipientPlayer TriggerAttackRecipient = 1 << (iota - 1)
	TriggerAttackRecipientPlaneswalker
	TriggerAttackRecipientBattle
)

// TriggerDamageRecipient identifies what received damage. Values are flags so a
// pattern can match either kind.
type TriggerDamageRecipient uint8

// Trigger damage recipient kinds.
const (
	TriggerDamageRecipientAny TriggerDamageRecipient = iota
	TriggerDamageRecipientPlayer
	TriggerDamageRecipientPermanent
)

// TriggerStackObject identifies a stack object involved in an event.
type TriggerStackObject uint8

// Trigger stack object kinds.
const (
	TriggerStackObjectAny TriggerStackObject = iota
	TriggerStackObjectSpell
)

// TriggerCounter identifies a counter kind used by a trigger pattern.
type TriggerCounter uint8

// Trigger counter kinds.
const (
	TriggerCounterAny TriggerCounter = iota
	TriggerCounterPlusOnePlusOne
	TriggerCounterMinusOneMinusOne
)

// TriggerCardType identifies a card type used by a semantic trigger Selection.
type TriggerCardType uint8

// Trigger card types.
const (
	TriggerCardTypeUnknown TriggerCardType = iota
	TriggerCardTypeArtifact
	TriggerCardTypeBattle
	TriggerCardTypeCreature
	TriggerCardTypeEnchantment
	TriggerCardTypeInstant
	TriggerCardTypeLand
	TriggerCardTypePlaneswalker
	TriggerCardTypeSorcery
)

// TriggerColor identifies a color used by a semantic trigger Selection.
type TriggerColor uint8

// Trigger colors.
const (
	TriggerColorUnknown TriggerColor = iota
	TriggerColorWhite
	TriggerColorBlue
	TriggerColorBlack
	TriggerColorRed
	TriggerColorGreen
)

// TriggerSubtype identifies a typed subtype used by a semantic trigger Selection.
type TriggerSubtype = types.Sub

// Trigger subtypes.
const (
	TriggerSubtypeUnknown TriggerSubtype = ""
	TriggerSubtypeSpirit  TriggerSubtype = types.Spirit
	TriggerSubtypeArcane  TriggerSubtype = types.Arcane
)

// TriggerSupertype identifies a supertype used by a semantic trigger Selection.
type TriggerSupertype uint8

// Trigger supertypes.
const (
	TriggerSupertypeUnknown TriggerSupertype = iota
	TriggerSupertypeLegendary
	TriggerSupertypeSnow
)

// TriggerKeyword identifies a keyword used by a semantic trigger Selection.
type TriggerKeyword uint8

// Trigger keywords.
const (
	TriggerKeywordUnknown TriggerKeyword = iota
	TriggerKeywordDefender
	TriggerKeywordFlash
	TriggerKeywordFlying
	TriggerKeywordHaste
	TriggerKeywordShadow
)

// TriggerTriState is a closed semantic true/false filter.
type TriggerTriState uint8

// Trigger tri-state values.
const (
	TriggerTriAny TriggerTriState = iota
	TriggerTriTrue
	TriggerTriFalse
)

// TriggerCombatState identifies a permanent's combat involvement.
type TriggerCombatState uint8

// Trigger combat-state values.
const (
	TriggerCombatStateAny TriggerCombatState = iota
	TriggerCombatStateAttacking
	TriggerCombatStateBlocking
)

// TriggerComparison identifies an integer-comparison relation.
type TriggerComparison uint8

// Trigger comparison relations.
const (
	TriggerComparisonUnknown TriggerComparison = iota
	TriggerComparisonEqual
	TriggerComparisonAtMost
	TriggerComparisonAtLeast
)

// TriggerNumberFilter is a closed semantic integer predicate.
type TriggerNumberFilter struct {
	Comparison TriggerComparison
	Value      int
}

// TriggerSelection is the closed semantic Selection vocabulary currently used
// by representable event subjects and cast spells. Its zero value is a
// wildcard.
type TriggerSelection struct {
	RequiredTypes    []TriggerCardType
	RequiredTypesAny []TriggerCardType
	ExcludedTypes    []TriggerCardType
	Supertypes       []TriggerSupertype
	SubtypesAny      []TriggerSubtype
	ColorsAny        []TriggerColor
	ExcludedColors   []TriggerColor
	Colorless        bool
	Multicolored     bool
	Tapped           TriggerTriState
	CombatState      TriggerCombatState
	Keyword          TriggerKeyword
	ExcludedKeyword  TriggerKeyword
	NonToken         bool
	TokenOnly        bool
	ManaValueAtLeast int
	MatchManaValue   bool
	ManaValue        TriggerNumberFilter
	Power            TriggerNumberFilter
	Toughness        TriggerNumberFilter
	Controller       ControllerKind
}

// TriggerPattern is a source-spanned semantic description of a representable
// event trigger. Raw trigger event text is deliberately not part of this
// lowering interface.
type TriggerPattern struct {
	Span shared.Span
	Kind TriggerKind

	Event      TriggerEvent
	Source     TriggerSourceRelation
	Subject    TriggerSubject
	Controller ControllerKind
	// CauseController identifies the controller of the spell or ability that
	// caused an event, independently from the event subject's controller.
	CauseController ControllerKind
	Player          TriggerPlayerRelation

	SubjectSelection         TriggerSelection
	RelatedSubjectSelection  TriggerSelection
	CardSelection            TriggerSelection
	DamageRecipientSelection TriggerSelection
	DamageSourceSelection    TriggerSelection
	AttackRecipientSelection TriggerSelection

	MatchFromZone bool
	FromZone      TriggerZone
	MatchToZone   bool
	ToZone        TriggerZone
	ExcludeToZone bool

	MatchFaceDown bool
	FaceDown      bool

	Step                              TriggerStep
	StepPlayerSourceAttachedSelection TriggerSelection
	CombatQualifier                   TriggerCombatQualifier
	DamageRecipient                   TriggerDamageRecipient
	DamageRecipientIsSource           bool
	DamageSourceIsStackObject         bool
	AttackRecipient                   TriggerAttackRecipient
	StackObject                       TriggerStackObject
	Counter                           TriggerCounter

	ExcludeSelf                bool
	OneOrMore                  bool
	OneOrMorePerAttackTarget   bool
	RequireKickerPaid          bool
	RequireHistoric            bool
	ExcludeManaAbility         bool
	PlayerEventOrdinalThisTurn int

	InterveningCondition *CompiledCondition
}

func compilePlayerEventTriggerPattern(
	clause *parser.PlayerEventTriggerClause,
	kind TriggerKind,
	condition *CompiledCondition,
) TriggerPattern {
	pattern := TriggerPattern{
		Span:                 clause.Span,
		Kind:                 kind,
		InterveningCondition: condition,
	}
	if kind != TriggerWhen && kind != TriggerWhenever {
		return pattern
	}
	event, ok := compilePlayerEventAction(clause.Action.Kind)
	if !ok {
		return pattern
	}
	player, ok := compilePlayerEventPlayer(clause.Player)
	if !ok {
		return pattern
	}
	modifiers := compilePlayerEventModifiers(
		clause.Action.Kind,
		clause.Player.Kind,
		clause.Card,
		clause.Occurrence,
	)
	if !modifiers.ok || occurrenceRequiresWhenever(clause.Occurrence.Kind) && kind != TriggerWhenever {
		return pattern
	}
	pattern.Event = event
	pattern.Player = player
	pattern.OneOrMore = modifiers.oneOrMore
	pattern.ExcludeSelf = modifiers.excludeSelf
	pattern.PlayerEventOrdinalThisTurn = modifiers.ordinal
	return pattern
}

func compilePlayerEventAction(action parser.PlayerEventActionKind) (TriggerEvent, bool) {
	switch action {
	case parser.PlayerEventActionDraw:
		return TriggerEventCardDrawn, true
	case parser.PlayerEventActionDiscard, parser.PlayerEventActionCycleOrDiscard:
		return TriggerEventCardDiscarded, true
	case parser.PlayerEventActionCycle:
		return TriggerEventCycled, true
	case parser.PlayerEventActionScry:
		return TriggerEventScry, true
	case parser.PlayerEventActionSurveil:
		return TriggerEventSurveil, true
	case parser.PlayerEventActionGainLife:
		return TriggerEventLifeGained, true
	case parser.PlayerEventActionLoseLife:
		return TriggerEventLifeLost, true
	default:
		return TriggerEventUnknown, false
	}
}

func compilePlayerEventPlayer(player parser.TriggerPlayerSelector) (TriggerPlayerRelation, bool) {
	switch player.Kind {
	case parser.TriggerPlayerSelectorAny:
		return TriggerPlayerAny, true
	case parser.TriggerPlayerSelectorYou:
		return TriggerPlayerYou, true
	case parser.TriggerPlayerSelectorOpponent:
		return TriggerPlayerOpponent, true
	default:
		return TriggerPlayerAny, false
	}
}

type compiledPlayerEventModifiers struct {
	oneOrMore   bool
	excludeSelf bool
	ordinal     int
	ok          bool
}

func compilePlayerEventModifiers(
	action parser.PlayerEventActionKind,
	player parser.TriggerPlayerSelectorKind,
	card parser.PlayerEventCard,
	occurrence parser.PlayerEventOccurrence,
) compiledPlayerEventModifiers {
	compiledCard := compilePlayerEventCard(action, card.Kind)
	if !compiledCard.ok {
		return compiledPlayerEventModifiers{}
	}
	ordinal, ok := compilePlayerEventOccurrence(action, player, occurrence)
	return compiledPlayerEventModifiers{
		oneOrMore:   compiledCard.oneOrMore,
		excludeSelf: compiledCard.excludeSelf,
		ordinal:     ordinal,
		ok:          ok,
	}
}

type compiledPlayerEventCard struct {
	oneOrMore   bool
	excludeSelf bool
	ok          bool
}

func compilePlayerEventCard(action parser.PlayerEventActionKind, card parser.PlayerEventCardKind) compiledPlayerEventCard {
	if !playerEventActionHasCard(action) {
		return compiledPlayerEventCard{ok: card == parser.PlayerEventCardNone}
	}
	switch card {
	case parser.PlayerEventCardSingle:
		return compiledPlayerEventCard{ok: true}
	case parser.PlayerEventCardOneOrMore:
		ok := action == parser.PlayerEventActionDiscard
		return compiledPlayerEventCard{oneOrMore: ok, ok: ok}
	case parser.PlayerEventCardAnother:
		ok := action == parser.PlayerEventActionDiscard ||
			action == parser.PlayerEventActionCycle ||
			action == parser.PlayerEventActionCycleOrDiscard
		return compiledPlayerEventCard{excludeSelf: ok, ok: ok}
	default:
		return compiledPlayerEventCard{}
	}
}

func compilePlayerEventOccurrence(
	action parser.PlayerEventActionKind,
	player parser.TriggerPlayerSelectorKind,
	occurrence parser.PlayerEventOccurrence,
) (int, bool) {
	switch occurrence.Kind {
	case parser.PlayerEventOccurrenceAny:
		return 0, occurrence.Ordinal == 0
	case parser.PlayerEventOccurrenceFirstEachTurn:
		return 1, occurrence.Ordinal == 1 && playerEventFirstEachTurnAllowed(action, player)
	case parser.PlayerEventOccurrenceOrdinalEachTurn:
		return occurrence.Ordinal, action == parser.PlayerEventActionDraw &&
			occurrence.Ordinal >= 1 &&
			occurrence.Ordinal <= 5
	default:
		return 0, false
	}
}

func occurrenceRequiresWhenever(occurrence parser.PlayerEventOccurrenceKind) bool {
	return occurrence == parser.PlayerEventOccurrenceAny
}

func compilePhaseStepTriggerPattern(
	clause *parser.PhaseStepTriggerClause,
	kind TriggerKind,
	condition *CompiledCondition,
) TriggerPattern {
	pattern := TriggerPattern{
		Span:                 clause.Span,
		Kind:                 kind,
		InterveningCondition: condition,
	}
	if kind != TriggerAt || !knownPhaseStepQuantifier(clause.Quantifier.Kind) {
		return pattern
	}
	step, ok := compilePhaseStepName(clause.Name.Kind)
	if !ok {
		return pattern
	}
	controller, attached, ok := compilePhaseStepPlayer(clause.Player)
	if !ok {
		return pattern
	}
	pattern.Event = TriggerEventBeginningOfStep
	pattern.Step = step
	pattern.Controller = controller
	pattern.StepPlayerSourceAttachedSelection = attached
	return pattern
}

func knownPhaseStepQuantifier(kind parser.PhaseStepQuantifierKind) bool {
	switch kind {
	case parser.PhaseStepQuantifierNone,
		parser.PhaseStepQuantifierSingle,
		parser.PhaseStepQuantifierEach,
		parser.PhaseStepQuantifierEachOf:
		return true
	default:
		return false
	}
}

func compilePhaseStepName(name parser.PhaseStepNameKind) (TriggerStep, bool) {
	switch name {
	case parser.PhaseStepNameUpkeep:
		return TriggerStepUpkeep, true
	case parser.PhaseStepNameDrawStep:
		return TriggerStepDraw, true
	case parser.PhaseStepNameEndStep:
		return TriggerStepEnd, true
	case parser.PhaseStepNameCombat, parser.PhaseStepNameCombatStep:
		return TriggerStepBeginningOfCombat, true
	case parser.PhaseStepNameEndOfCombat, parser.PhaseStepNameEndOfCombatStep:
		return TriggerStepEndOfCombat, true
	case parser.PhaseStepNamePrecombatMainPhase, parser.PhaseStepNameFirstMainPhase:
		return TriggerStepPrecombatMain, true
	case parser.PhaseStepNamePostcombatMainPhase, parser.PhaseStepNameSecondMainPhase:
		return TriggerStepPostcombatMain, true
	default:
		return TriggerStepNone, false
	}
}

func compilePhaseStepPlayer(player parser.TriggerPlayerSelector) (ControllerKind, TriggerSelection, bool) {
	switch player.Kind {
	case parser.TriggerPlayerSelectorAny:
		return ControllerAny, TriggerSelection{}, true
	case parser.TriggerPlayerSelectorYou, parser.TriggerPlayerSelectorSourceController:
		return ControllerYou, TriggerSelection{}, true
	case parser.TriggerPlayerSelectorOpponent:
		return ControllerOpponent, TriggerSelection{}, true
	case parser.TriggerPlayerSelectorAttachedController:
		selection, ok := compilePhaseStepAttachedSubject(player.AttachedSubject)
		return ControllerAny, selection, ok
	default:
		return ControllerAny, TriggerSelection{}, false
	}
}

func compilePhaseStepAttachedSubject(subject parser.TriggerAttachedSubject) (TriggerSelection, bool) {
	selection, ok := compileTriggerSelection(subject.Selection)
	if !ok || phaseStepAttachedSelectionEmpty(selection) {
		// A wildcard Selection is empty, which the runtime interprets as no
		// attached-controller relation rather than any attached permanent.
		return TriggerSelection{}, false
	}
	return selection, true
}

func compileTriggerSelection(syntax parser.TriggerSelection) (TriggerSelection, bool) {
	selection := TriggerSelection{
		Colorless:    syntax.Colorless,
		Multicolored: syntax.Multicolored,
		NonToken:     syntax.NonToken,
		TokenOnly:    syntax.TokenOnly,
	}
	var ok bool
	for _, value := range syntax.RequiredTypes {
		selection.RequiredTypes = append(selection.RequiredTypes, compileTriggerCardType(value))
		if selection.RequiredTypes[len(selection.RequiredTypes)-1] == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.RequiredTypesAny {
		selection.RequiredTypesAny = append(selection.RequiredTypesAny, compileTriggerCardType(value))
		if selection.RequiredTypesAny[len(selection.RequiredTypesAny)-1] == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ExcludedTypes {
		selection.ExcludedTypes = append(selection.ExcludedTypes, compileTriggerCardType(value))
		if selection.ExcludedTypes[len(selection.ExcludedTypes)-1] == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.Supertypes {
		selection.Supertypes = append(selection.Supertypes, compileTriggerSupertype(value))
		if selection.Supertypes[len(selection.Supertypes)-1] == TriggerSupertypeUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ColorsAny {
		selection.ColorsAny = append(selection.ColorsAny, compileTriggerColor(value))
		if selection.ColorsAny[len(selection.ColorsAny)-1] == TriggerColorUnknown {
			return TriggerSelection{}, false
		}
	}
	for _, value := range syntax.ExcludedColors {
		selection.ExcludedColors = append(selection.ExcludedColors, compileTriggerColor(value))
		if selection.ExcludedColors[len(selection.ExcludedColors)-1] == TriggerColorUnknown {
			return TriggerSelection{}, false
		}
	}
	if len(syntax.SubtypesAny) > 0 {
		selection.SubtypesAny = make([]TriggerSubtype, len(syntax.SubtypesAny))
		copy(selection.SubtypesAny, syntax.SubtypesAny)
	}
	selection.Controller, ok = compileTriggerController(syntax.Controller)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.Tapped, ok = compileTriggerSelectionTapped(syntax.Tapped)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.CombatState, ok = compileTriggerSelectionCombatState(syntax.CombatState)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.Keyword, ok = compileTriggerSelectionKeyword(syntax.Keyword)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.ExcludedKeyword, ok = compileTriggerSelectionKeyword(syntax.ExcludedKeyword)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.ManaValue, ok = compileTriggerSelectionNumber(syntax.ManaValue)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.Power, ok = compileTriggerSelectionNumber(syntax.Power)
	if !ok {
		return TriggerSelection{}, false
	}
	selection.Toughness, ok = compileTriggerSelectionNumber(syntax.Toughness)
	return selection, ok
}

func compileTriggerCardType(value parser.TriggerCardType) TriggerCardType {
	switch value {
	case parser.TriggerCardTypeArtifact:
		return TriggerCardTypeArtifact
	case parser.TriggerCardTypeBattle:
		return TriggerCardTypeBattle
	case parser.TriggerCardTypeCreature:
		return TriggerCardTypeCreature
	case parser.TriggerCardTypeEnchantment:
		return TriggerCardTypeEnchantment
	case parser.TriggerCardTypeInstant:
		return TriggerCardTypeInstant
	case parser.TriggerCardTypeLand:
		return TriggerCardTypeLand
	case parser.TriggerCardTypePlaneswalker:
		return TriggerCardTypePlaneswalker
	case parser.TriggerCardTypeSorcery:
		return TriggerCardTypeSorcery
	default:
		return TriggerCardTypeUnknown
	}
}

func compileTriggerSupertype(value parser.TriggerSupertype) TriggerSupertype {
	switch value {
	case parser.TriggerSupertypeLegendary:
		return TriggerSupertypeLegendary
	case parser.TriggerSupertypeSnow:
		return TriggerSupertypeSnow
	default:
		return TriggerSupertypeUnknown
	}
}

func compileTriggerColor(value parser.TriggerColor) TriggerColor {
	switch value {
	case parser.TriggerColorWhite:
		return TriggerColorWhite
	case parser.TriggerColorBlue:
		return TriggerColorBlue
	case parser.TriggerColorBlack:
		return TriggerColorBlack
	case parser.TriggerColorRed:
		return TriggerColorRed
	case parser.TriggerColorGreen:
		return TriggerColorGreen
	default:
		return TriggerColorUnknown
	}
}

func compileTriggerController(value parser.TriggerController) (ControllerKind, bool) {
	switch value {
	case parser.ControllerAny:
		return ControllerAny, true
	case parser.ControllerYou:
		return ControllerYou, true
	case parser.ControllerOpponent:
		return ControllerOpponent, true
	default:
		return ControllerAny, false
	}
}

func compileTriggerSelectionTapped(value parser.TriggerSelectionTappedState) (TriggerTriState, bool) {
	switch value {
	case parser.TriggerSelectionTappedAny:
		return TriggerTriAny, true
	case parser.TriggerSelectionTapped:
		return TriggerTriTrue, true
	case parser.TriggerSelectionUntapped:
		return TriggerTriFalse, true
	default:
		return TriggerTriAny, false
	}
}

func compileTriggerSelectionCombatState(value parser.TriggerSelectionCombatState) (TriggerCombatState, bool) {
	switch value {
	case parser.TriggerSelectionCombatAny:
		return TriggerCombatStateAny, true
	case parser.TriggerSelectionAttacking:
		return TriggerCombatStateAttacking, true
	case parser.TriggerSelectionBlocking:
		return TriggerCombatStateBlocking, true
	default:
		return TriggerCombatStateAny, false
	}
}

func compileTriggerSelectionKeyword(value parser.KeywordKind) (TriggerKeyword, bool) {
	switch value {
	case parser.KeywordUnknown:
		return TriggerKeywordUnknown, true
	case parser.KeywordDefender:
		return TriggerKeywordDefender, true
	case parser.KeywordFlash:
		return TriggerKeywordFlash, true
	case parser.KeywordFlying:
		return TriggerKeywordFlying, true
	case parser.KeywordHaste:
		return TriggerKeywordHaste, true
	case parser.KeywordShadow:
		return TriggerKeywordShadow, true
	default:
		return TriggerKeywordUnknown, false
	}
}

func compileTriggerSelectionNumber(value parser.TriggerSelectionNumber) (TriggerNumberFilter, bool) {
	switch value.Comparison {
	case parser.TriggerSelectionComparisonUnknown:
		return TriggerNumberFilter{}, value.Value == 0
	case parser.TriggerSelectionComparisonEqual:
		return TriggerNumberFilter{Comparison: TriggerComparisonEqual, Value: value.Value}, value.Value >= 0
	case parser.TriggerSelectionComparisonAtMost:
		return TriggerNumberFilter{Comparison: TriggerComparisonAtMost, Value: value.Value}, value.Value >= 0
	case parser.TriggerSelectionComparisonAtLeast:
		return TriggerNumberFilter{Comparison: TriggerComparisonAtLeast, Value: value.Value}, value.Value >= 0
	default:
		return TriggerNumberFilter{}, false
	}
}

func playerEventActionHasCard(action parser.PlayerEventActionKind) bool {
	switch action {
	case parser.PlayerEventActionDraw,
		parser.PlayerEventActionDiscard,
		parser.PlayerEventActionCycle,
		parser.PlayerEventActionCycleOrDiscard:
		return true
	default:
		return false
	}
}

func playerEventFirstEachTurnAllowed(action parser.PlayerEventActionKind, player parser.TriggerPlayerSelectorKind) bool {
	switch action {
	case parser.PlayerEventActionDraw,
		parser.PlayerEventActionScry,
		parser.PlayerEventActionSurveil:
		return true
	case parser.PlayerEventActionGainLife, parser.PlayerEventActionLoseLife:
		return player != parser.TriggerPlayerSelectorAny
	default:
		return false
	}
}

func phaseStepAttachedSelectionEmpty(selection TriggerSelection) bool {
	return len(selection.RequiredTypes) == 0 &&
		len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		len(selection.Supertypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		len(selection.ExcludedColors) == 0 &&
		!selection.Colorless &&
		!selection.Multicolored &&
		selection.Tapped == TriggerTriAny &&
		selection.CombatState == TriggerCombatStateAny &&
		selection.Keyword == TriggerKeywordUnknown &&
		selection.ExcludedKeyword == TriggerKeywordUnknown &&
		!selection.NonToken &&
		!selection.TokenOnly &&
		selection.ManaValueAtLeast == 0 &&
		!selection.MatchManaValue &&
		selection.ManaValue.Comparison == TriggerComparisonUnknown &&
		selection.Power.Comparison == TriggerComparisonUnknown &&
		selection.Toughness.Comparison == TriggerComparisonUnknown &&
		selection.Controller == ControllerAny
}

func compileTriggerEventPattern(
	clause *parser.TriggerEventClause,
	kind TriggerKind,
	condition *CompiledCondition,
) TriggerPattern {
	pattern := TriggerPattern{
		Span:                 clause.Span,
		Kind:                 kind,
		InterveningCondition: condition,
	}
	if kind != TriggerWhen && kind != TriggerWhenever {
		return pattern
	}
	compiled, ok := compileTriggerEventClause(clause)
	if !ok {
		return pattern
	}
	compiled.Span = clause.Span
	compiled.Kind = kind
	compiled.InterveningCondition = condition
	return compiled
}

func compileTriggerEventClause(clause *parser.TriggerEventClause) (TriggerPattern, bool) {
	pattern := TriggerPattern{
		ExcludeSelf:               clause.ExcludeSelf,
		OneOrMore:                 clause.OneOrMore,
		OneOrMorePerAttackTarget:  clause.OneOrMorePerAttackTarget,
		ExcludeManaAbility:        clause.ExcludeManaAbility,
		DamageSourceIsStackObject: clause.DamageSourceIsStackObject,
		MatchFaceDown:             clause.FaceDown,
		FaceDown:                  clause.FaceDown,
	}
	var ok bool
	pattern.Controller, ok = compileTriggerController(clause.Controller)
	if !ok {
		return TriggerPattern{}, false
	}
	pattern.Player, ok = compileOptionalTriggerPlayer(clause.Player)
	if !ok {
		return TriggerPattern{}, false
	}
	switch clause.Kind {
	case parser.TriggerEventKindZoneChange:
		ok = compileZoneChangeEvent(clause, &pattern)
	case parser.TriggerEventKindSpellCast:
		ok = compileSpellCastEvent(clause, &pattern)
	case parser.TriggerEventKindAbilityActivated:
		ok = compileAbilityActivatedEvent(clause, &pattern)
	case parser.TriggerEventKindAttack:
		ok = compileAttackEvent(clause, &pattern)
	case parser.TriggerEventKindBlock:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventBlockerDeclared)
	case parser.TriggerEventKindBecameBlocked:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventAttackerBecameBlocked)
	case parser.TriggerEventKindDamageDealt:
		ok = compileDamageEvent(clause, &pattern)
	case parser.TriggerEventKindCounterAdded:
		ok = compileCounterEvent(clause, &pattern)
	case parser.TriggerEventKindBecomesTapped:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventPermanentTapped)
	case parser.TriggerEventKindBecomesUntapped:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventPermanentUntapped)
	case parser.TriggerEventKindTurnedFaceUp:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventPermanentTurnedFaceUp)
	case parser.TriggerEventKindSacrificed:
		ok = compileSacrificeEvent(clause, &pattern)
	case parser.TriggerEventKindMutated:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventPermanentMutated)
	case parser.TriggerEventKindBecameTarget:
		ok = compileBecameTargetEvent(clause, &pattern)
	default:
		return TriggerPattern{}, false
	}
	if !ok {
		return TriggerPattern{}, false
	}
	return pattern, true
}

func compileZoneChangeEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if !compileEventSubject(clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	switch clause.ZoneChange.Kind {
	case parser.TriggerEventZoneChangeEnteredBattlefield:
		pattern.Event = TriggerEventPermanentEnteredBattlefield
	case parser.TriggerEventZoneChangeDied:
		pattern.Event = TriggerEventPermanentDied
	case parser.TriggerEventZoneChangeMoved:
		pattern.Event = TriggerEventZoneChanged
	default:
		return false
	}
	if !compileZoneChangeZones(clause, pattern) {
		return false
	}
	switch clause.Tapped.Kind {
	case parser.TriggerEventTappedStateAny:
	case parser.TriggerEventTappedStateTapped:
		pattern.SubjectSelection.Tapped = TriggerTriTrue
	case parser.TriggerEventTappedStateUntapped:
		pattern.SubjectSelection.Tapped = TriggerTriFalse
	default:
		return false
	}
	return true
}

func compileZoneChangeZones(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if pattern.Event == TriggerEventPermanentDied {
		return true
	}
	if clause.Zone.MatchFromZone {
		pattern.FromZone, _ = compileTriggerEventZone(clause.Zone.FromZone.Kind)
		if pattern.FromZone == TriggerZoneNone {
			return false
		}
		pattern.MatchFromZone = true
	}
	if pattern.Event != TriggerEventZoneChanged {
		return true
	}
	pattern.MatchToZone = clause.Zone.MatchToZone
	pattern.ExcludeToZone = clause.Zone.ExcludeToZone
	if !pattern.MatchToZone && !pattern.ExcludeToZone {
		return true
	}
	pattern.ToZone, _ = compileTriggerEventZone(clause.Zone.ToZone.Kind)
	return pattern.ToZone != TriggerZoneNone &&
		(!pattern.MatchToZone || !pattern.ExcludeToZone)
}

func compileSpellCastEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if clause.Actor.Kind == parser.TriggerEventActorUnknown {
		return false
	}
	controller, ok := compileTriggerActorController(clause.Actor.Kind)
	if !ok {
		return false
	}
	selection, ok := compileTriggerSpellSelection(clause.SpellSelection)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventSpellCast
	pattern.Controller = controller
	pattern.CardSelection = selection
	pattern.RequireKickerPaid = clause.SpellSelection.Kicker
	pattern.RequireHistoric = clause.SpellSelection.Historic
	if clause.SpellSelection.FromZone.Kind != parser.TriggerEventZoneNone {
		pattern.FromZone, ok = compileTriggerEventZone(clause.SpellSelection.FromZone.Kind)
		if !ok || controller != ControllerYou {
			return false
		}
		pattern.MatchFromZone = true
	}
	return true
}

func compileAbilityActivatedEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	player, ok := compileTriggerActorPlayer(clause.Actor.Kind)
	if !ok || !clause.ExcludeManaAbility {
		return false
	}
	selection, ok := compileTriggerSelection(clause.SourceSelection)
	if !ok || selection.Controller != ControllerAny {
		return false
	}
	pattern.Event = TriggerEventAbilityActivated
	pattern.Player = player
	pattern.SubjectSelection = selection
	return true
}

func compileAttackEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if clause.Actor.Kind != parser.TriggerEventActorUnknown {
		controller, ok := compileTriggerActorController(clause.Actor.Kind)
		if !ok {
			return false
		}
		pattern.Controller = controller
	}
	if clause.Subject.Kind != parser.TriggerEventSubjectUnknown &&
		!compileEventSubject(clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	recipient, ok := compileTriggerAttackRecipient(clause.AttackRecipient.Kind)
	if !ok {
		return false
	}
	selection, ok := compileTriggerSelection(clause.AttackRecipient.Selection)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventAttackerDeclared
	pattern.AttackRecipient = recipient
	pattern.AttackRecipientSelection = selection
	return true
}

func compilePermanentSubjectEvent(
	clause *parser.TriggerEventClause,
	pattern *TriggerPattern,
	event TriggerEvent,
) bool {
	if !compileEventSubject(clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	related, ok := compileTriggerSelection(clause.RelatedSelection)
	if !ok {
		return false
	}
	pattern.Event = event
	if event != TriggerEventAttackerBecameBlocked {
		pattern.RelatedSubjectSelection = related
	}
	return true
}

func compileDamageEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	qualifier, ok := compileTriggerCombatQualifier(clause.CombatQualifier.Kind)
	if !ok {
		return false
	}
	recipient, ok := compileTriggerDamageRecipient(clause.DamageRecipient.Kind)
	if !ok {
		return false
	}
	recipientSelection, ok := compileTriggerSelection(clause.DamageRecipient.Selection)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventDamageDealt
	pattern.CombatQualifier = qualifier
	pattern.DamageRecipient = recipient
	pattern.DamageRecipientSelection = recipientSelection
	pattern.DamageRecipientIsSource = clause.DamageRecipient.IsSource
	if clause.StackObject.Kind != parser.TriggerEventStackObjectAny {
		pattern.StackObject, ok = compileTriggerStackObject(clause.StackObject.Kind)
		if !ok {
			return false
		}
	}
	if clause.DamageSource.Kind != parser.TriggerEventSubjectUnknown {
		pattern.Subject = TriggerSubjectDamageSource
		return compileDamageSourceSubject(clause.DamageSource, pattern)
	}
	if clause.DamageSourceIsStackObject {
		if clause.DamageSourceSpellSelection.Kicker ||
			clause.DamageSourceSpellSelection.Historic ||
			clause.DamageSourceSpellSelection.FromZone.Kind != parser.TriggerEventZoneNone {
			return false
		}
		pattern.Subject = TriggerSubjectDamageSource
		selection, selectionOK := compileTriggerSpellSelection(clause.DamageSourceSpellSelection)
		if !selectionOK {
			return false
		}
		pattern.DamageSourceSelection = selection
		return pattern.StackObject == TriggerStackObjectSpell
	}
	if clause.Subject.Kind == parser.TriggerEventSubjectUnknown {
		return recipient == TriggerDamageRecipientPlayer
	}
	if clause.Subject.Kind == parser.TriggerEventSubjectSelf {
		pattern.Subject = TriggerSubjectPermanent
	}
	return compileEventSubject(clause.Subject, pattern, &pattern.SubjectSelection)
}

func compileCounterEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if clause.Subject.Kind != parser.TriggerEventSubjectSelf {
		return false
	}
	counterValue, ok := compileTriggerCounter(clause.Counter.Kind)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventCountersAdded
	pattern.Source = TriggerSourceSelf
	pattern.Counter = counterValue
	return true
}

func compileSacrificeEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	player, ok := compileTriggerActorPlayer(clause.Actor.Kind)
	if !ok || !compileEventSubject(clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	pattern.Event = TriggerEventPermanentSacrificed
	pattern.Player = player
	return true
}

func compileBecameTargetEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if !compileEventSubject(clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	stackObject, ok := compileTriggerStackObject(clause.StackObject.Kind)
	if !ok {
		return false
	}
	causeController, ok := compileTriggerActorController(clause.CauseController)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventObjectBecameTarget
	pattern.StackObject = stackObject
	pattern.CauseController = causeController
	return true
}

func compileEventSubject(
	subject parser.TriggerEventSubject,
	pattern *TriggerPattern,
	destination *TriggerSelection,
) bool {
	selection, ok := compileTriggerSelection(subject.Selection)
	if !ok {
		return false
	}
	switch subject.Kind {
	case parser.TriggerEventSubjectSelf:
		pattern.Source = TriggerSourceSelf
		*destination = selection
	case parser.TriggerEventSubjectAttached:
		pattern.Source = TriggerSourceAttachedPermanent
		*destination = selection
	case parser.TriggerEventSubjectSelection:
		*destination = selection
	case parser.TriggerEventSubjectDamageSource:
	default:
		return false
	}
	return true
}

func compileDamageSourceSubject(subject parser.TriggerEventSubject, pattern *TriggerPattern) bool {
	selection, ok := compileTriggerSelection(subject.Selection)
	if !ok {
		return false
	}
	switch subject.Kind {
	case parser.TriggerEventSubjectSelf:
		pattern.Source = TriggerSourceSelf
	case parser.TriggerEventSubjectAttached:
		pattern.Source = TriggerSourceAttachedPermanent
		pattern.DamageSourceSelection = selection
	case parser.TriggerEventSubjectSelection:
		pattern.DamageSourceSelection = selection
	case parser.TriggerEventSubjectDamageSource:
	default:
		return false
	}
	return true
}

func compileTriggerSpellSelection(syntax parser.TriggerEventSpellSelection) (TriggerSelection, bool) {
	selection := TriggerSelection{
		Colorless:        syntax.Colorless,
		Multicolored:     syntax.Multicolored,
		ManaValueAtLeast: syntax.ManaValueAtLeast,
		MatchManaValue:   syntax.MatchManaValue,
	}
	for _, value := range syntax.Types {
		compiled := compileTriggerCardType(value)
		if compiled == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		selection.RequiredTypes = append(selection.RequiredTypes, compiled)
	}
	for _, value := range syntax.TypesAny {
		compiled := compileTriggerCardType(value)
		if compiled == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		selection.RequiredTypesAny = append(selection.RequiredTypesAny, compiled)
	}
	for _, value := range syntax.ExcludedTypes {
		compiled := compileTriggerCardType(value)
		if compiled == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		selection.ExcludedTypes = append(selection.ExcludedTypes, compiled)
	}
	for _, value := range syntax.ColorsAny {
		compiled := compileTriggerColor(value)
		if compiled == TriggerColorUnknown {
			return TriggerSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, compiled)
	}
	if len(syntax.SubtypesAny) > 0 {
		selection.SubtypesAny = append(selection.SubtypesAny, syntax.SubtypesAny...)
	}
	return selection, true
}

func compileOptionalTriggerPlayer(player parser.TriggerPlayerSelector) (TriggerPlayerRelation, bool) {
	if player.Kind == parser.TriggerPlayerSelectorUnknown {
		return TriggerPlayerAny, true
	}
	return compilePlayerEventPlayer(player)
}

func compileTriggerActorPlayer(actor parser.TriggerEventActorKind) (TriggerPlayerRelation, bool) {
	switch actor {
	case parser.TriggerEventActorYou:
		return TriggerPlayerYou, true
	case parser.TriggerEventActorOpponent:
		return TriggerPlayerOpponent, true
	case parser.TriggerEventActorPlayer:
		return TriggerPlayerAny, true
	default:
		return TriggerPlayerAny, false
	}
}

func compileTriggerActorController(actor parser.TriggerEventActorKind) (ControllerKind, bool) {
	switch actor {
	case parser.TriggerEventActorUnknown, parser.TriggerEventActorPlayer:
		return ControllerAny, true
	case parser.TriggerEventActorYou:
		return ControllerYou, true
	case parser.TriggerEventActorOpponent:
		return ControllerOpponent, true
	default:
		return ControllerAny, false
	}
}

func compileTriggerEventZone(value parser.TriggerEventZoneKind) (TriggerZone, bool) {
	switch value {
	case parser.TriggerEventZoneNone:
		return TriggerZoneNone, true
	case parser.TriggerEventZoneGraveyard:
		return TriggerZoneGraveyard, true
	case parser.TriggerEventZoneBattlefield:
		return TriggerZoneBattlefield, true
	case parser.TriggerEventZoneHand:
		return TriggerZoneHand, true
	case parser.TriggerEventZoneExile:
		return TriggerZoneExile, true
	case parser.TriggerEventZoneLibrary:
		return TriggerZoneLibrary, true
	case parser.TriggerEventZoneStack:
		return TriggerZoneStack, true
	case parser.TriggerEventZoneCommand:
		return TriggerZoneCommand, true
	default:
		return TriggerZoneNone, false
	}
}

func compileTriggerCombatQualifier(value parser.TriggerEventCombatQualifierKind) (TriggerCombatQualifier, bool) {
	switch value {
	case parser.TriggerEventCombatQualifierAny:
		return TriggerCombatAny, true
	case parser.TriggerEventCombatQualifierCombat:
		return TriggerCombatDamage, true
	case parser.TriggerEventCombatQualifierNoncombat:
		return TriggerNonCombatDamage, true
	default:
		return TriggerCombatAny, false
	}
}

func compileTriggerDamageRecipient(value parser.TriggerEventDamageRecipientKind) (TriggerDamageRecipient, bool) {
	const known = parser.TriggerEventDamageRecipientPlayer | parser.TriggerEventDamageRecipientPermanent
	if value&^known != 0 {
		return TriggerDamageRecipientAny, false
	}
	return TriggerDamageRecipient(value), true
}

func compileTriggerAttackRecipient(value parser.TriggerEventAttackRecipientKind) (TriggerAttackRecipient, bool) {
	const known = parser.TriggerEventAttackRecipientPlayer |
		parser.TriggerEventAttackRecipientPlaneswalker |
		parser.TriggerEventAttackRecipientBattle
	if value&^known != 0 {
		return TriggerAttackRecipientAny, false
	}
	return TriggerAttackRecipient(value), true
}

func compileTriggerStackObject(value parser.TriggerEventStackObjectKind) (TriggerStackObject, bool) {
	switch value {
	case parser.TriggerEventStackObjectAny:
		return TriggerStackObjectAny, true
	case parser.TriggerEventStackObjectSpell:
		return TriggerStackObjectSpell, true
	default:
		return TriggerStackObjectAny, false
	}
}

func compileTriggerCounter(value parser.TriggerEventCounterKind) (TriggerCounter, bool) {
	switch value {
	case parser.TriggerEventCounterPlusOnePlusOne:
		return TriggerCounterPlusOnePlusOne, true
	case parser.TriggerEventCounterMinusOneMinusOne:
		return TriggerCounterMinusOneMinusOne, true
	default:
		return TriggerCounterAny, false
	}
}
