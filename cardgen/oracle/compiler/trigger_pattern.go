package compiler

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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

type triggerPatternTemplate struct {
	kinds []TriggerKind
	bind  func(triggerEventSyntax, TriggerKind) (TriggerPattern, bool)
}

type controllerPhraseSlot struct {
	text       string
	controller ControllerKind
}

var triggerPatternTemplates = []triggerPatternTemplate{
	{kinds: []TriggerKind{TriggerWhen, TriggerWhenever}, bind: recognizePermanentZoneChangeTrigger},
	{kinds: []TriggerKind{TriggerWhen, TriggerWhenever}, bind: recognizeSpellAbilityTrigger},
	{kinds: []TriggerKind{TriggerWhen, TriggerWhenever}, bind: recognizeCombatTrigger},
	{kinds: []TriggerKind{TriggerWhen, TriggerWhenever}, bind: recognizePermanentStateTrigger},
	{kinds: []TriggerKind{TriggerWhen, TriggerWhenever}, bind: recognizeSacrificeTrigger},
}

type triggerEventSyntax struct {
	text   string
	tokens []shared.Token
	atoms  parser.Atoms
}

func newTriggerEventSyntax(event string, tokens []shared.Token, atoms parser.Atoms) triggerEventSyntax {
	return triggerEventSyntax{text: strings.ToLower(event), tokens: tokens, atoms: atoms}
}

func (e triggerEventSyntax) tokensFor(text string) []shared.Token {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return nil
	}
	for start := range e.tokens {
		for end := start + 1; end <= len(e.tokens); end++ {
			if strings.ToLower(joinedSourceText(e.tokens[start:end])) == text {
				return e.tokens[start:end]
			}
		}
	}
	return nil
}

func (e triggerEventSyntax) selfSubject(text string, slots []string, allowName bool) bool {
	if slices.ContainsFunc(slots, func(slot string) bool { return strings.EqualFold(slot, text) }) {
		tokens := e.tokensFor(text)
		if len(tokens) == 0 {
			return false
		}
		span := shared.SpanOf(tokens)
		if len(e.atoms.ReferencesIn(span)) > 0 {
			return true
		}
		markerSpan, ok := e.atoms.SourceMarkerSpanStartingAt(tokens[0].Span)
		return ok && markerSpan == span
	}
	if !allowName {
		return false
	}
	tokens := e.tokensFor(text)
	if len(tokens) == 0 {
		return false
	}
	span := shared.SpanOf(tokens)
	nameSpan, ok := e.atoms.SourceNameSpanStartingAt(tokens[0].Span)
	return ok && nameSpan == span
}

func compileTriggerPattern(
	event string,
	kind TriggerKind,
	span shared.Span,
	cardName string,
	condition *CompiledCondition,
) TriggerPattern {
	source := triggerIntroForKind(kind) + " " + event + ", draw a card."
	document, _ := parser.Parse(source, parser.Context{CardName: cardName})
	if len(document.Abilities) == 1 && document.Abilities[0].Trigger != nil {
		ability := document.Abilities[0]
		return compileTriggerPatternForSyntax(event, kind, span, ability.Trigger.Event.Tokens, ability.Atoms, condition)
	}
	return compileTriggerPatternForSyntax(event, kind, span, nil, parser.Atoms{}, condition)
}

func triggerIntroForKind(kind TriggerKind) string {
	switch kind {
	case TriggerAt:
		return "At"
	case TriggerWhenever:
		return "Whenever"
	default:
		return "When"
	}
}
func compileTriggerPatternForSyntax(
	event string,
	kind TriggerKind,
	span shared.Span,
	eventTokens []shared.Token,
	atoms parser.Atoms,
	condition *CompiledCondition,
) TriggerPattern {
	eventSyntax := newTriggerEventSyntax(event, eventTokens, atoms)
	pattern := TriggerPattern{
		Span:                 span,
		Kind:                 kind,
		InterveningCondition: condition,
	}
	recognized, ok := bindTriggerPatternTemplates(eventSyntax, kind, triggerPatternTemplates)
	if ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	return pattern
}

func bindTriggerPatternTemplates(
	event triggerEventSyntax,
	kind TriggerKind,
	templates []triggerPatternTemplate,
) (TriggerPattern, bool) {
	var recognized TriggerPattern
	matched := false
	for _, template := range templates {
		if !slices.Contains(template.kinds, kind) {
			continue
		}
		candidate, ok := template.bind(event, kind)
		if !ok {
			continue
		}
		if matched {
			// Overlapping templates make the event clause ambiguous.
			return TriggerPattern{}, false
		}
		recognized = candidate
		matched = true
	}
	return recognized, matched
}

func completeTriggerPattern(recognized, source *TriggerPattern) TriggerPattern {
	recognized.Span = source.Span
	recognized.Kind = source.Kind
	recognized.InterveningCondition = source.InterveningCondition
	return *recognized
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

func recognizePermanentZoneChangeTrigger(event triggerEventSyntax, kind TriggerKind) (TriggerPattern, bool) {
	return recognizeZoneChangeTrigger(event, kind)
}

func recognizeSpellAbilityTrigger(event triggerEventSyntax, kind TriggerKind) (TriggerPattern, bool) {
	if pattern, ok := recognizeCastTrigger(event, kind); ok {
		return pattern, true
	}
	if pattern, ok := recognizeAbilityActivatedTrigger(event, kind); ok {
		return pattern, true
	}
	return recognizeBecameTargetTrigger(event)
}

func recognizeAbilityActivatedTrigger(event triggerEventSyntax, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	for _, actor := range []struct {
		prefix   string
		relation TriggerPlayerRelation
	}{
		{prefix: "you activate ", relation: TriggerPlayerYou},
		{prefix: "an opponent activates ", relation: TriggerPlayerOpponent},
		{prefix: "a player activates ", relation: TriggerPlayerAny},
	} {
		ability, ok := strings.CutPrefix(event.text, actor.prefix)
		if !ok {
			continue
		}
		pattern := TriggerPattern{
			Event:  TriggerEventAbilityActivated,
			Player: actor.relation,
		}
		ability, pattern.ExcludeManaAbility = strings.CutSuffix(ability, " that isn't a mana ability")
		if !pattern.ExcludeManaAbility {
			return TriggerPattern{}, false
		}
		if ability == "an ability" {
			return pattern, true
		}
		source, ok := strings.CutPrefix(ability, "an ability of ")
		if !ok {
			return TriggerPattern{}, false
		}
		parsed := parseCombatPermanentSelection(event, source, false)
		if !parsed.ok || parsed.controller != ControllerAny || parsed.excludeSelf {
			return TriggerPattern{}, false
		}
		pattern.SubjectSelection = parsed.selection
		return pattern, true
	}
	return TriggerPattern{}, false
}

func recognizeCombatTrigger(event triggerEventSyntax, kind TriggerKind) (TriggerPattern, bool) {
	if pattern, ok := recognizeAttackBlockTrigger(event); ok {
		return pattern, true
	}
	if pattern, ok := recognizeParameterizedDamageTrigger(event); ok {
		return pattern, true
	}
	return recognizePermanentActionTrigger(event, kind, combatPermanentActions)
}

func recognizeAttackBlockTrigger(event triggerEventSyntax) (TriggerPattern, bool) {
	if pattern, ok := recognizePlayerAttackTrigger(event.text); ok {
		return pattern, true
	}
	for _, template := range []struct {
		marker  string
		event   TriggerEvent
		plural  bool
		related bool
	}{
		{marker: " becomes blocked by ", event: TriggerEventAttackerBecameBlocked, related: true},
		{marker: " blocks ", event: TriggerEventBlockerDeclared, related: true},
		{marker: " block ", event: TriggerEventBlockerDeclared, plural: true, related: true},
		{marker: " attacks ", event: TriggerEventAttackerDeclared},
		{marker: " attack ", event: TriggerEventAttackerDeclared, plural: true},
	} {
		subject, remainder, ok := strings.Cut(event.text, template.marker)
		if !ok {
			continue
		}
		pattern, ok := combatSubjectPattern(event, subject, template.event, template.plural)
		if !ok {
			return TriggerPattern{}, false
		}
		if template.related {
			related, ok := parseRelatedCombatSelection(event, remainder)
			if !ok {
				return TriggerPattern{}, false
			}
			if template.event == TriggerEventAttackerBecameBlocked {
				if !basicCreatureSelection(related) {
					return TriggerPattern{}, false
				}
			} else {
				pattern.RelatedSubjectSelection = related
			}
			return pattern, true
		}
		recipient := parseAttackRecipient(event, remainder)
		if !recipient.ok {
			return TriggerPattern{}, false
		}
		pattern.AttackRecipient = recipient.recipient
		pattern.Player = recipient.player
		pattern.AttackRecipientSelection = recipient.selection
		return pattern, true
	}
	for _, template := range []struct {
		suffix string
		event  TriggerEvent
		plural bool
	}{
		{suffix: " becomes blocked", event: TriggerEventAttackerBecameBlocked},
		{suffix: " attacks", event: TriggerEventAttackerDeclared},
		{suffix: " attack", event: TriggerEventAttackerDeclared, plural: true},
		{suffix: " blocks", event: TriggerEventBlockerDeclared},
		{suffix: " block", event: TriggerEventBlockerDeclared, plural: true},
	} {
		if !strings.HasSuffix(event.text, template.suffix) {
			continue
		}
		return combatSubjectPattern(event, strings.TrimSuffix(event.text, template.suffix), template.event, template.plural)
	}
	return TriggerPattern{}, false
}

func recognizePlayerAttackTrigger(event string) (TriggerPattern, bool) {
	for _, template := range []struct {
		text       string
		controller ControllerKind
		player     TriggerPlayerRelation
		recipient  TriggerAttackRecipient
		perTarget  bool
	}{
		{text: "you attack", controller: ControllerYou},
		{text: "you attack with one or more creatures", controller: ControllerYou},
		{text: "an opponent attacks", controller: ControllerOpponent},
		{text: "a player attacks", controller: ControllerAny},
		{text: "you attack a player", controller: ControllerYou, recipient: TriggerAttackRecipientPlayer, perTarget: true},
		{text: "an opponent attacks you", controller: ControllerOpponent, player: TriggerPlayerYou, recipient: TriggerAttackRecipientPlayer, perTarget: true},
		{text: "a player attacks you", controller: ControllerAny, player: TriggerPlayerYou, recipient: TriggerAttackRecipientPlayer, perTarget: true},
		{text: "a player attacks one of your opponents", controller: ControllerAny, player: TriggerPlayerOpponent, recipient: TriggerAttackRecipientPlayer, perTarget: true},
	} {
		if event != template.text {
			continue
		}
		return TriggerPattern{
			Event:                    TriggerEventAttackerDeclared,
			Controller:               template.controller,
			Player:                   template.player,
			AttackRecipient:          template.recipient,
			OneOrMore:                true,
			OneOrMorePerAttackTarget: template.perTarget,
		}, true
	}
	return TriggerPattern{}, false
}

var selfCombatSubjectSlots = []string{
	"this creature",
	"this permanent",
	"this artifact",
	"this battle",
	"this enchantment",
	"this land",
	"this planeswalker",
	"this vehicle",
}

func combatSubjectPattern(syntax triggerEventSyntax, subject string, eventKind TriggerEvent, plural bool) (TriggerPattern, bool) {
	oneOrMore := false
	if rest, ok := strings.CutPrefix(subject, "one or more "); ok {
		subject = rest
		plural = true
		oneOrMore = true
	}
	if syntax.selfSubject(subject, selfCombatSubjectSlots, true) {
		return TriggerPattern{
			Event:     eventKind,
			Source:    TriggerSourceSelf,
			OneOrMore: oneOrMore,
		}, true
	}
	if subject == "enchanted creature" || subject == "equipped creature" {
		return TriggerPattern{
			Event:     eventKind,
			Source:    TriggerSourceAttachedPermanent,
			OneOrMore: oneOrMore,
			SubjectSelection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			},
		}, true
	}
	parsed := parseCombatPermanentSelection(syntax, subject, plural)
	if !parsed.ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Event:            eventKind,
		Controller:       parsed.controller,
		ExcludeSelf:      parsed.excludeSelf,
		OneOrMore:        oneOrMore,
		SubjectSelection: parsed.selection,
	}, true
}

type combatPermanentSelection struct {
	selection   TriggerSelection
	controller  ControllerKind
	excludeSelf bool
	ok          bool
}

func parseCombatPermanentSelection(event triggerEventSyntax, subject string, plural bool) combatPermanentSelection {
	relations := parsePermanentSubjectRelations(subject)
	if !relations.ok || relations.player != TriggerPlayerAny {
		return combatPermanentSelection{}
	}
	subject = relations.subject
	excludeSelf := false
	if plural {
		subject, excludeSelf = strings.CutPrefix(subject, "other ")
	} else {
		switch {
		case strings.HasPrefix(subject, "another "):
			subject = strings.TrimPrefix(subject, "another ")
			excludeSelf = true
		case strings.HasPrefix(subject, "a "):
			subject = strings.TrimPrefix(subject, "a ")
		case strings.HasPrefix(subject, "an "):
			subject = strings.TrimPrefix(subject, "an ")
		default:
			return combatPermanentSelection{}
		}
	}
	if subject == "source" || subject == "sources" || subject == "player" || subject == "players" {
		return combatPermanentSelection{}
	}
	selection, ok := parsePermanentTriggerSelection(event, subject, plural)
	if !ok {
		return combatPermanentSelection{}
	}
	return combatPermanentSelection{
		selection:   selection,
		controller:  relations.controller,
		excludeSelf: excludeSelf,
		ok:          true,
	}
}

func parseRelatedCombatSelection(event triggerEventSyntax, subject string) (TriggerSelection, bool) {
	parsed := parseCombatPermanentSelection(event, subject, false)
	if !parsed.ok {
		return TriggerSelection{}, false
	}
	parsed.selection.Controller = parsed.controller
	return parsed.selection, true
}

func basicCreatureSelection(selection TriggerSelection) bool {
	return len(selection.RequiredTypes) == 1 &&
		selection.RequiredTypes[0] == TriggerCardTypeCreature &&
		selection.Controller == ControllerAny &&
		len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		len(selection.ExcludedColors) == 0 &&
		!selection.Colorless &&
		!selection.Multicolored &&
		selection.Tapped == TriggerTriAny &&
		selection.CombatState == TriggerCombatStateAny &&
		selection.Keyword == TriggerKeywordUnknown &&
		selection.ExcludedKeyword == TriggerKeywordUnknown &&
		selection.ManaValueAtLeast == 0 &&
		!selection.MatchManaValue &&
		selection.ManaValue.Comparison == TriggerComparisonUnknown &&
		selection.Power.Comparison == TriggerComparisonUnknown &&
		selection.Toughness.Comparison == TriggerComparisonUnknown &&
		!selection.NonToken &&
		!selection.TokenOnly
}

type attackRecipientPattern struct {
	recipient TriggerAttackRecipient
	player    TriggerPlayerRelation
	selection TriggerSelection
	ok        bool
}

func parseAttackRecipient(event triggerEventSyntax, recipient string) attackRecipientPattern {
	switch recipient {
	case "you":
		return attackRecipientPattern{recipient: TriggerAttackRecipientPlayer, player: TriggerPlayerYou, ok: true}
	case "an opponent", "one of your opponents":
		return attackRecipientPattern{recipient: TriggerAttackRecipientPlayer, player: TriggerPlayerOpponent, ok: true}
	case "a player":
		return attackRecipientPattern{recipient: TriggerAttackRecipientPlayer, ok: true}
	case "a player or planeswalker":
		return attackRecipientPattern{
			recipient: TriggerAttackRecipientPlayer | TriggerAttackRecipientPlaneswalker,
			selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker}},
			ok:        true,
		}
	case "a player or battle":
		return attackRecipientPattern{
			recipient: TriggerAttackRecipientPlayer | TriggerAttackRecipientBattle,
			selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeBattle}},
			ok:        true,
		}
	case "you or a planeswalker you control":
		return attackRecipientPattern{
			recipient: TriggerAttackRecipientPlayer | TriggerAttackRecipientPlaneswalker,
			player:    TriggerPlayerYou,
			selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker}, Controller: ControllerYou},
			ok:        true,
		}
	case "you or a battle you protect":
		return attackRecipientPattern{
			recipient: TriggerAttackRecipientPlayer | TriggerAttackRecipientBattle,
			player:    TriggerPlayerYou,
			selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeBattle}},
			ok:        true,
		}
	}
	parsed := parseCombatPermanentSelection(event, recipient, false)
	if !parsed.ok || len(parsed.selection.RequiredTypes) != 1 {
		return attackRecipientPattern{}
	}
	var result TriggerAttackRecipient
	switch parsed.selection.RequiredTypes[0] {
	case TriggerCardTypePlaneswalker:
		result = TriggerAttackRecipientPlaneswalker
	case TriggerCardTypeBattle:
		result = TriggerAttackRecipientBattle
	default:
		return attackRecipientPattern{}
	}
	parsed.selection.Controller = parsed.controller
	player := TriggerPlayerAny
	switch parsed.controller {
	case ControllerYou:
		player = TriggerPlayerYou
	case ControllerOpponent:
		player = TriggerPlayerOpponent
	default:
	}
	return attackRecipientPattern{recipient: result, player: player, selection: parsed.selection, ok: true}
}

func recognizePermanentStateTrigger(event triggerEventSyntax, kind TriggerKind) (TriggerPattern, bool) {
	if pattern, ok := recognizeSelfPermanentStateTrigger(event, kind); ok {
		return pattern, true
	}
	return recognizePermanentActionTrigger(event, kind, statePermanentActions)
}

func recognizeSacrificeTrigger(event triggerEventSyntax, _ TriggerKind) (TriggerPattern, bool) {
	for _, actor := range []struct {
		prefix   string
		relation TriggerPlayerRelation
	}{
		{prefix: "you sacrifice ", relation: TriggerPlayerYou},
		{prefix: "an opponent sacrifices ", relation: TriggerPlayerOpponent},
		{prefix: "a player sacrifices ", relation: TriggerPlayerAny},
	} {
		subject, ok := strings.CutPrefix(event.text, actor.prefix)
		if !ok {
			continue
		}
		pattern := TriggerPattern{
			Event:  TriggerEventPermanentSacrificed,
			Player: actor.relation,
		}
		if event.selfSubject(subject, []string{
			"this creature",
			"this permanent",
			"this artifact",
			"this enchantment",
			"this land",
		}, true) {
			pattern.Source = TriggerSourceSelf
			return pattern, true
		}
		if subject, pattern.OneOrMore = strings.CutPrefix(subject, "one or more "); pattern.OneOrMore {
			parsed := parseCombatPermanentSelection(event, subject, true)
			if !parsed.ok {
				return TriggerPattern{}, false
			}
			pattern.Controller = parsed.controller
			pattern.ExcludeSelf = parsed.excludeSelf
			pattern.SubjectSelection = parsed.selection
			return pattern, true
		}
		parsed := parseCombatPermanentSelection(event, subject, false)
		if !parsed.ok {
			return TriggerPattern{}, false
		}
		pattern.Controller = parsed.controller
		pattern.ExcludeSelf = parsed.excludeSelf
		pattern.SubjectSelection = parsed.selection
		return pattern, true
	}
	return TriggerPattern{}, false
}

var selfEnterSubjectSlots = []string{
	"this creature",
	"this permanent",
	"this token",
	"this aura",
	"this artifact",
	"this equipment",
	"this land",
	"this vehicle",
	"this enchantment",
	"this battle",
	"this siege",
	"this case",
	"this class",
	"this planeswalker",
	"this spacecraft",
}

var selfStateSubjectSlots = []string{
	"this creature",
	"this permanent",
	"this token",
	"this aura",
	"this land",
	"this artifact",
	"this equipment",
	"this enchantment",
	"this vehicle",
	"this battle",
	"this siege",
	"this case",
	"this class",
	"this planeswalker",
	"this spacecraft",
}

func recognizeSelfPermanentStateTrigger(event triggerEventSyntax, kind TriggerKind) (TriggerPattern, bool) {
	if kind == TriggerWhenever && event.text == "this creature mutates" {
		return TriggerPattern{Event: TriggerEventPermanentMutated, Source: TriggerSourceSelf}, true
	}
	for _, template := range []struct {
		suffix    string
		event     TriggerEvent
		allowWhen bool
	}{
		{suffix: " becomes tapped", event: TriggerEventPermanentTapped},
		{suffix: " becomes untapped", event: TriggerEventPermanentUntapped},
		{suffix: " is turned face up", event: TriggerEventPermanentTurnedFaceUp, allowWhen: true},
	} {
		if kind != TriggerWhenever && (kind != TriggerWhen || !template.allowWhen) {
			continue
		}
		subject, ok := strings.CutSuffix(event.text, template.suffix)
		if ok && event.selfSubject(subject, selfStateSubjectSlots, true) {
			return TriggerPattern{Event: template.event, Source: TriggerSourceSelf}, true
		}
	}
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	return recognizeSelfCounterTrigger(event)
}

func recognizeSelfCounterTrigger(event triggerEventSyntax) (TriggerPattern, bool) {
	oneOrMore := false
	_, subject, ok := strings.Cut(event.text, " counter is put on ")
	if !ok {
		_, subject, ok = strings.Cut(event.text, " counters are put on ")
		if !ok {
			return TriggerPattern{}, false
		}
		_, oneOrMore = strings.CutPrefix(event.text, "one or more ")
		if !oneOrMore {
			return TriggerPattern{}, false
		}
	} else if !strings.HasPrefix(event.text, "a ") {
		return TriggerPattern{}, false
	}
	if subject != "this creature" && subject != "this permanent" {
		return TriggerPattern{}, false
	}
	counterValue, ok := triggerCounterAtom(shared.SpanOf(event.tokens), event.atoms)
	if !ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Event:     TriggerEventCountersAdded,
		Source:    TriggerSourceSelf,
		Counter:   counterValue,
		OneOrMore: oneOrMore,
	}, true
}

func triggerCounterAtom(span shared.Span, atoms parser.Atoms) (TriggerCounter, bool) {
	kind, _, ok := atoms.CounterIn(span)
	if !ok {
		return TriggerCounterAny, false
	}
	switch kind {
	case counter.PlusOnePlusOne:
		return TriggerCounterPlusOnePlusOne, true
	case counter.MinusOneMinusOne:
		return TriggerCounterMinusOneMinusOne, true
	default:
		return TriggerCounterAny, false
	}
}

func recognizeBecameTargetTrigger(event triggerEventSyntax) (TriggerPattern, bool) {
	subject, cause, ok := strings.Cut(event.text, " becomes the target of ")
	if !ok {
		return TriggerPattern{}, false
	}
	stackObject, causeController, ok := parseBecameTargetCause(cause)
	if !ok {
		return TriggerPattern{}, false
	}
	pattern, ok := becameTargetSubjectPattern(event, subject)
	if !ok {
		return TriggerPattern{}, false
	}
	pattern.Event = TriggerEventObjectBecameTarget
	pattern.StackObject = stackObject
	pattern.CauseController = causeController
	return pattern, true
}

func parseBecameTargetCause(cause string) (TriggerStackObject, ControllerKind, bool) {
	controller := ControllerAny
	switch {
	case strings.HasSuffix(cause, " you control"):
		cause = strings.TrimSuffix(cause, " you control")
		controller = ControllerYou
	case strings.HasSuffix(cause, " an opponent controls"):
		cause = strings.TrimSuffix(cause, " an opponent controls")
		controller = ControllerOpponent
	default:
	}
	switch cause {
	case "a spell":
		return TriggerStackObjectSpell, controller, true
	case "a spell or ability":
		return TriggerStackObjectAny, controller, true
	default:
		return TriggerStackObjectAny, ControllerAny, false
	}
}

func becameTargetSubjectPattern(event triggerEventSyntax, subject string) (TriggerPattern, bool) {
	if event.selfSubject(subject, []string{
		"this creature",
		"this permanent",
		"this artifact",
		"this enchantment",
		"this land",
		"this planeswalker",
	}, true) {
		return TriggerPattern{Source: TriggerSourceSelf}, true
	}
	if subject == "enchanted creature" {
		return TriggerPattern{
			Source: TriggerSourceAttachedPermanent,
			SubjectSelection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			},
		}, true
	}
	parsed := parseCombatPermanentSelection(event, subject, false)
	if !parsed.ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Controller:       parsed.controller,
		ExcludeSelf:      parsed.excludeSelf,
		SubjectSelection: parsed.selection,
	}, true
}

func recognizeParameterizedDamageTrigger(event triggerEventSyntax) (TriggerPattern, bool) {
	for _, template := range []struct {
		text      string
		qualifier TriggerCombatQualifier
	}{
		{text: "you're dealt combat damage", qualifier: TriggerCombatDamage},
		{text: "you are dealt combat damage", qualifier: TriggerCombatDamage},
		{text: "you're dealt noncombat damage", qualifier: TriggerNonCombatDamage},
		{text: "you are dealt noncombat damage", qualifier: TriggerNonCombatDamage},
		{text: "you're dealt damage"},
		{text: "you are dealt damage"},
	} {
		if event.text == template.text {
			return TriggerPattern{
				Event:           TriggerEventDamageDealt,
				Player:          TriggerPlayerYou,
				CombatQualifier: template.qualifier,
				DamageRecipient: TriggerDamageRecipientPlayer,
			}, true
		}
	}
	for _, template := range []struct {
		suffix    string
		qualifier TriggerCombatQualifier
		plural    bool
	}{
		{suffix: " is dealt combat damage", qualifier: TriggerCombatDamage},
		{suffix: " is dealt noncombat damage", qualifier: TriggerNonCombatDamage},
		{suffix: " is dealt damage"},
		{suffix: " are dealt combat damage", qualifier: TriggerCombatDamage, plural: true},
		{suffix: " are dealt noncombat damage", qualifier: TriggerNonCombatDamage, plural: true},
		{suffix: " are dealt damage", plural: true},
	} {
		subject, ok := strings.CutSuffix(event.text, template.suffix)
		if !ok {
			continue
		}
		pattern, ok := damageRecipientSubjectPattern(event, subject, template.plural)
		if !ok {
			return TriggerPattern{}, false
		}
		pattern.CombatQualifier = template.qualifier
		return pattern, true
	}
	for _, template := range []struct {
		marker    string
		qualifier TriggerCombatQualifier
		plural    bool
	}{
		{marker: " deals combat damage", qualifier: TriggerCombatDamage},
		{marker: " deals noncombat damage", qualifier: TriggerNonCombatDamage},
		{marker: " deals damage"},
		{marker: " deal combat damage", qualifier: TriggerCombatDamage, plural: true},
		{marker: " deal noncombat damage", qualifier: TriggerNonCombatDamage, plural: true},
		{marker: " deal damage", plural: true},
	} {
		source, remainder, ok := strings.Cut(event.text, template.marker)
		if !ok {
			continue
		}
		pattern, ok := damageSourcePattern(event, source, template.plural)
		if !ok {
			return TriggerPattern{}, false
		}
		pattern.CombatQualifier = template.qualifier
		if remainder == "" {
			return pattern, true
		}
		target, ok := strings.CutPrefix(remainder, " to ")
		if !ok {
			return TriggerPattern{}, false
		}
		recipient := parseDamageRecipient(event, target)
		if !recipient.ok {
			return TriggerPattern{}, false
		}
		pattern.DamageRecipient = recipient.recipient
		pattern.Player = recipient.player
		pattern.DamageRecipientSelection = recipient.selection
		pattern.DamageRecipientIsSource = recipient.isSource
		return pattern, true
	}
	return TriggerPattern{}, false
}

func damageSourcePattern(event triggerEventSyntax, subject string, plural bool) (TriggerPattern, bool) {
	oneOrMore := false
	if rest, ok := strings.CutPrefix(subject, "one or more "); ok {
		subject = rest
		plural = true
		oneOrMore = true
	}
	if subject == "a source" {
		return TriggerPattern{
			Event:   TriggerEventDamageDealt,
			Subject: TriggerSubjectDamageSource,
		}, true
	}
	if event.selfSubject(subject, selfCombatSubjectSlots, true) {
		return TriggerPattern{
			Event:     TriggerEventDamageDealt,
			Source:    TriggerSourceSelf,
			Subject:   TriggerSubjectDamageSource,
			OneOrMore: oneOrMore,
		}, true
	}
	if pattern, ok := damageStackObjectSourcePattern(event, subject, plural); ok {
		pattern.OneOrMore = pattern.OneOrMore || oneOrMore
		return pattern, true
	}
	if subject == "enchanted creature" || subject == "equipped creature" {
		return TriggerPattern{
			Event:     TriggerEventDamageDealt,
			Source:    TriggerSourceAttachedPermanent,
			Subject:   TriggerSubjectDamageSource,
			OneOrMore: oneOrMore,
			DamageSourceSelection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			},
		}, true
	}

	parsed := parseCombatPermanentSelection(event, subject, plural)
	if !parsed.ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Event:                 TriggerEventDamageDealt,
		Subject:               TriggerSubjectDamageSource,
		Controller:            parsed.controller,
		ExcludeSelf:           parsed.excludeSelf,
		OneOrMore:             oneOrMore,
		DamageSourceSelection: parsed.selection,
	}, true
}

func damageStackObjectSourcePattern(event triggerEventSyntax, subject string, plural bool) (TriggerPattern, bool) {
	relations := parsePermanentSubjectRelations(subject)
	if !relations.ok || relations.player != TriggerPlayerAny {
		return TriggerPattern{}, false
	}
	subject = relations.subject
	if plural {
		if rest, ok := strings.CutPrefix(subject, "other "); ok {
			subject = rest
		}
	} else {
		switch {
		case strings.HasPrefix(subject, "a "):
			subject = strings.TrimPrefix(subject, "a ")
		case strings.HasPrefix(subject, "an "):
			subject = strings.TrimPrefix(subject, "an ")
		default:
			return TriggerPattern{}, false
		}
	}
	suffix := " spell"
	if plural {
		suffix = " spells"
	}
	if !strings.HasSuffix(subject, suffix) {
		return TriggerPattern{}, false
	}
	subject = strings.TrimSpace(strings.TrimSuffix(subject, suffix))
	selection, ok := parseSpellTriggerSelection(event, subject)
	if !ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Event:                     TriggerEventDamageDealt,
		Subject:                   TriggerSubjectDamageSource,
		Controller:                relations.controller,
		StackObject:               TriggerStackObjectSpell,
		DamageSourceIsStackObject: true,
		DamageSourceSelection:     selection,
	}, true
}

func parseSpellTriggerSelection(event triggerEventSyntax, subject string) (TriggerSelection, bool) {
	switch subject {
	case "":
		return TriggerSelection{}, true
	case "noncreature":
		return TriggerSelection{ExcludedTypes: []TriggerCardType{TriggerCardTypeCreature}}, true
	case "instant or sorcery":
		tokens := event.tokensFor(subject)
		if len(tokens) != 3 || !equalWord(tokens[1], "or") {
			return TriggerSelection{}, false
		}
		left, leftOK := parseSingleTriggerPermanentType(event, tokens[0].Text, false)
		right, rightOK := parseSingleTriggerPermanentType(event, tokens[2].Text, false)
		if !leftOK || !rightOK || left == TriggerCardTypeUnknown || right == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		return TriggerSelection{RequiredTypesAny: []TriggerCardType{left, right}}, true
	default:
		tokens := event.tokensFor(subject)
		if len(tokens) != 1 {
			return TriggerSelection{}, false
		}
		cardType, ok := parseSingleTriggerPermanentType(event, tokens[0].Text, false)
		if !ok || cardType == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		return TriggerSelection{RequiredTypes: []TriggerCardType{cardType}}, true
	}
}

func damageRecipientSubjectPattern(event triggerEventSyntax, subject string, plural bool) (TriggerPattern, bool) {
	oneOrMore := false
	if rest, ok := strings.CutPrefix(subject, "one or more "); ok {
		subject = rest
		plural = true
		oneOrMore = true
	}
	if event.selfSubject(subject, selfCombatSubjectSlots, true) {
		return TriggerPattern{
			Event:           TriggerEventDamageDealt,
			Source:          TriggerSourceSelf,
			Subject:         TriggerSubjectPermanent,
			OneOrMore:       oneOrMore,
			DamageRecipient: TriggerDamageRecipientPermanent,
		}, true
	}
	if subject == "enchanted creature" || subject == "equipped creature" || subject == "enchanted permanent" {
		selection := TriggerSelection{}
		if subject != "enchanted permanent" {
			selection.RequiredTypes = []TriggerCardType{TriggerCardTypeCreature}
		}
		return TriggerPattern{
			Event:            TriggerEventDamageDealt,
			Source:           TriggerSourceAttachedPermanent,
			OneOrMore:        oneOrMore,
			DamageRecipient:  TriggerDamageRecipientPermanent,
			SubjectSelection: selection,
		}, true
	}
	parsed := parseCombatPermanentSelection(event, subject, plural)
	if !parsed.ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Event:            TriggerEventDamageDealt,
		Controller:       parsed.controller,
		ExcludeSelf:      parsed.excludeSelf,
		OneOrMore:        oneOrMore,
		DamageRecipient:  TriggerDamageRecipientPermanent,
		SubjectSelection: parsed.selection,
	}, true
}

type damageRecipientPattern struct {
	recipient TriggerDamageRecipient
	player    TriggerPlayerRelation
	selection TriggerSelection
	isSource  bool
	ok        bool
}

func parseDamageRecipient(event triggerEventSyntax, recipient string) damageRecipientPattern {
	if event.selfSubject(recipient, selfCombatSubjectSlots, true) {
		return damageRecipientPattern{recipient: TriggerDamageRecipientPermanent, isSource: true, ok: true}
	}
	switch recipient {
	case "you":
		return damageRecipientPattern{recipient: TriggerDamageRecipientPlayer, player: TriggerPlayerYou, ok: true}
	case "an opponent", "one of your opponents":
		return damageRecipientPattern{recipient: TriggerDamageRecipientPlayer, player: TriggerPlayerOpponent, ok: true}
	case "a player":
		return damageRecipientPattern{recipient: TriggerDamageRecipientPlayer, ok: true}
	case "a player or planeswalker":
		return damageRecipientPattern{
			recipient: TriggerDamageRecipientPlayer | TriggerDamageRecipientPermanent,
			selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker}},
			ok:        true,
		}
	case "a player or battle":
		return damageRecipientPattern{
			recipient: TriggerDamageRecipientPlayer | TriggerDamageRecipientPermanent,
			selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeBattle}},
			ok:        true,
		}
	case "any target":
		return damageRecipientPattern{
			recipient: TriggerDamageRecipientPlayer | TriggerDamageRecipientPermanent,
			selection: TriggerSelection{RequiredTypesAny: []TriggerCardType{
				TriggerCardTypeCreature,
				TriggerCardTypePlaneswalker,
				TriggerCardTypeBattle,
			}},
			ok: true,
		}
	}
	parsed := parseCombatPermanentSelection(event, recipient, false)
	if !parsed.ok {
		return damageRecipientPattern{}
	}
	parsed.selection.Controller = parsed.controller
	player := TriggerPlayerAny
	switch parsed.controller {
	case ControllerYou:
		player = TriggerPlayerYou
	case ControllerOpponent:
		player = TriggerPlayerOpponent
	default:
	}
	return damageRecipientPattern{
		recipient: TriggerDamageRecipientPermanent,
		player:    player,
		selection: parsed.selection,
		ok:        true,
	}
}

func recognizeDamageTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	if event == "you're dealt damage" || event == "you are dealt damage" {
		return TriggerPattern{
			Event:           TriggerEventDamageDealt,
			Player:          TriggerPlayerYou,
			DamageRecipient: TriggerDamageRecipientPlayer,
		}, true
	}
	if subject, ok := strings.CutSuffix(event, " is dealt damage"); ok {
		switch subject {
		case "this creature", "this permanent":
			return TriggerPattern{
				Event:           TriggerEventDamageDealt,
				Source:          TriggerSourceSelf,
				Subject:         TriggerSubjectPermanent,
				DamageRecipient: TriggerDamageRecipientPermanent,
			}, true
		case "enchanted creature", "enchanted permanent", "equipped creature":
			return TriggerPattern{
				Event:           TriggerEventDamageDealt,
				Source:          TriggerSourceAttachedPermanent,
				DamageRecipient: TriggerDamageRecipientPermanent,
			}, true
		}
	}
	for _, source := range []struct {
		text     string
		relation TriggerSourceRelation
	}{
		{"this creature", TriggerSourceSelf},
		{"equipped creature", TriggerSourceAttachedPermanent},
		{"enchanted creature", TriggerSourceAttachedPermanent},
	} {
		prefix := source.text + " deals "
		if !strings.HasPrefix(event, prefix) {
			continue
		}
		rest := strings.TrimPrefix(event, prefix)
		pattern := TriggerPattern{
			Event:   TriggerEventDamageDealt,
			Source:  source.relation,
			Subject: TriggerSubjectDamageSource,
		}
		if strings.HasPrefix(rest, "combat damage") {
			pattern.CombatQualifier = TriggerCombatDamage
			rest = strings.TrimPrefix(rest, "combat ")
		}
		switch rest {
		case "damage":
			return pattern, true
		case "damage to a player":
			pattern.DamageRecipient = TriggerDamageRecipientPlayer
			return pattern, true
		case "damage to an opponent":
			pattern.DamageRecipient = TriggerDamageRecipientPlayer
			pattern.Player = TriggerPlayerOpponent
			return pattern, true
		case "damage to a creature":
			pattern.DamageRecipient = TriggerDamageRecipientPermanent
			pattern.DamageRecipientSelection.RequiredTypes = []TriggerCardType{TriggerCardTypeCreature}
			return pattern, true
		}
	}
	return TriggerPattern{}, false
}

func recognizeCastTrigger(event triggerEventSyntax, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	pattern := TriggerPattern{Event: TriggerEventSpellCast}
	var phrase string
	for _, relation := range []controllerPhraseSlot{
		{text: "you cast ", controller: ControllerYou},
		{text: "a player casts ", controller: ControllerAny},
		{text: "an opponent casts ", controller: ControllerOpponent},
	} {
		if rest, ok := strings.CutPrefix(event.text, relation.text); ok {
			pattern.Controller = relation.controller
			phrase = rest
			break
		}
	}
	if phrase == "" {
		return TriggerPattern{}, false
	}
	if !compileCastSelection(event, phrase, &pattern) {
		return TriggerPattern{}, false
	}
	if pattern.MatchFromZone && pattern.Controller != ControllerYou {
		return TriggerPattern{}, false
	}
	return pattern, true
}

func compileCastSelection(event triggerEventSyntax, phrase string, pattern *TriggerPattern) bool {
	type predicate struct {
		selection     TriggerSelection
		kicker        bool
		historic      bool
		fromGraveyard bool
	}
	predicates := map[string]predicate{
		"a spell":          {},
		"a kicked spell":   {kicker: true},
		"a historic spell": {historic: true},
	}
	if predicate, ok := predicates[phrase]; ok {
		pattern.CardSelection = predicate.selection
		pattern.RequireKickerPaid = predicate.kicker
		pattern.RequireHistoric = predicate.historic
		if predicate.fromGraveyard {
			pattern.MatchFromZone = true
			pattern.FromZone = TriggerZoneGraveyard
		}
		return true
	}
	if compileCastAtomSelection(event, phrase, pattern) {
		return true
	}
	const (
		prefix = "a spell with mana value "
		suffix = " or greater"
	)
	valueText, ok := strings.CutPrefix(phrase, prefix)
	if !ok {
		return false
	}
	valueText, ok = strings.CutSuffix(valueText, suffix)
	if !ok {
		return false
	}
	value := 0
	for _, r := range valueText {
		if r < '0' || r > '9' {
			return false
		}
		value = value*10 + int(r-'0')
	}
	if valueText == "" {
		return false
	}
	pattern.CardSelection.MatchManaValue = true
	pattern.CardSelection.ManaValueAtLeast = value
	return true
}

func compileCastAtomSelection(event triggerEventSyntax, phrase string, pattern *TriggerPattern) bool {
	if strings.Contains(phrase, "copy") {
		return false
	}
	tokens := event.tokensFor(phrase)
	if len(tokens) == 0 {
		return false
	}
	switch phrase {
	case "a noncreature spell":
		return compileCastSingleTypeSelection(event, tokens, pattern, 1, true)
	case "a creature spell", "an instant spell", "an instant", "a sorcery spell",
		"an artifact spell", "an enchantment spell", "a land spell", "a planeswalker spell":
		return compileCastSingleTypeSelection(event, tokens, pattern, 1, false)
	case "an instant or sorcery spell":
		if len(tokens) != 5 || !equalWord(tokens[2], "or") {
			return false
		}
		left, ok := triggerCardTypeAt(event, tokens[1])
		if !ok {
			return false
		}
		right, ok := triggerCardTypeAt(event, tokens[3])
		if !ok {
			return false
		}
		if (left != TriggerCardTypeInstant || right != TriggerCardTypeSorcery) &&
			(left != TriggerCardTypeSorcery || right != TriggerCardTypeInstant) {
			return false
		}
		pattern.CardSelection.RequiredTypesAny = []TriggerCardType{left, right}
		return true
	case "a noncreature, nonland spell":
		if len(tokens) != 5 || tokens[2].Kind != shared.Comma {
			return false
		}
		left, ok := triggerExcludedCardTypeAt(event, tokens[1])
		if !ok {
			return false
		}
		right, ok := triggerExcludedCardTypeAt(event, tokens[3])
		if !ok {
			return false
		}
		pattern.CardSelection.ExcludedTypes = []TriggerCardType{left, right}
		return true
	case "a white spell", "a blue spell", "a black spell", "a red spell", "a green spell":
		color, ok := triggerColorAt(event, tokens, 1)
		if !ok {
			return false
		}
		pattern.CardSelection.ColorsAny = []TriggerColor{color}
		return true
	case "a colorless spell", "a multicolored spell":
		qualifier, ok := event.atoms.ColorQualifierAt(tokens[1].Span)
		if !ok {
			return false
		}
		switch qualifier {
		case parser.ColorQualifierColorless:
			pattern.CardSelection.Colorless = true
		case parser.ColorQualifierMulticolored:
			pattern.CardSelection.Multicolored = true
		default:
			return false
		}
		return true
	case "a spell from your graveyard":
		if len(tokens) != 5 {
			return false
		}
		z, ok := event.atoms.ZoneIn(shared.SpanOf(tokens[2:]), parser.ZoneRoleFrom)
		if !ok || z != zone.Graveyard {
			return false
		}
		pattern.MatchFromZone = true
		pattern.FromZone = TriggerZoneGraveyard
		return true
	case "a spirit or arcane spell":
		if len(tokens) != 5 || !equalWord(tokens[2], "or") {
			return false
		}
		left, ok := triggerSubtypeFromPhrase(event, tokens[1].Text)
		if !ok {
			return false
		}
		right, ok := triggerSubtypeFromPhrase(event, tokens[3].Text)
		if !ok {
			return false
		}
		pattern.CardSelection.SubtypesAny = []TriggerSubtype{left, right}
		return true
	default:
		return false
	}
}

func compileCastSingleTypeSelection(event triggerEventSyntax, tokens []shared.Token, pattern *TriggerPattern, index int, excluded bool) bool {
	if index >= len(tokens) {
		return false
	}
	cardType, ok := triggerCardTypeAt(event, tokens[index])
	if excluded {
		cardType, ok = triggerExcludedCardTypeAt(event, tokens[index])
	}
	if !ok {
		return false
	}
	if excluded {
		pattern.CardSelection.ExcludedTypes = []TriggerCardType{cardType}
	} else {
		pattern.CardSelection.RequiredTypes = []TriggerCardType{cardType}
	}
	return true
}

func triggerCardTypeAt(event triggerEventSyntax, token shared.Token) (TriggerCardType, bool) {
	cardType, ok := event.atoms.CardTypeAt(token.Span)
	if !ok {
		return TriggerCardTypeUnknown, false
	}
	return triggerCardTypeFromParser(cardType)
}

func triggerExcludedCardTypeAt(event triggerEventSyntax, token shared.Token) (TriggerCardType, bool) {
	cardType, ok := event.atoms.ExcludedCardTypeAt(token.Span)
	if !ok {
		return TriggerCardTypeUnknown, false
	}
	return triggerCardTypeFromParser(cardType)
}

func triggerColorAt(event triggerEventSyntax, tokens []shared.Token, index int) (TriggerColor, bool) {
	if index >= len(tokens) {
		return TriggerColorUnknown, false
	}
	color, ok := event.atoms.ColorAt(tokens[index].Span)
	if !ok {
		return TriggerColorUnknown, false
	}
	return triggerColorFromParser(color)
}

func triggerColorFromParser(color parser.Color) (TriggerColor, bool) {
	switch color {
	case parser.ColorWhite:
		return TriggerColorWhite, true
	case parser.ColorBlue:
		return TriggerColorBlue, true
	case parser.ColorBlack:
		return TriggerColorBlack, true
	case parser.ColorRed:
		return TriggerColorRed, true
	case parser.ColorGreen:
		return TriggerColorGreen, true
	default:
		return TriggerColorUnknown, false
	}
}

type permanentZoneChangeTemplate struct {
	singularSuffix string
	pluralSuffix   string
	event          TriggerEvent
	matchFromZone  bool
	fromZone       TriggerZone
	matchToZone    bool
	toZone         TriggerZone
	excludeToZone  bool
	controller     ControllerKind
	player         TriggerPlayerRelation
	tapped         TriggerTriState
}

var permanentZoneChangeTemplates = []permanentZoneChangeTemplate{
	{singularSuffix: " enters tapped", pluralSuffix: " enter tapped", event: TriggerEventPermanentEnteredBattlefield, tapped: TriggerTriTrue},
	{singularSuffix: " enters untapped", pluralSuffix: " enter untapped", event: TriggerEventPermanentEnteredBattlefield, tapped: TriggerTriFalse},
	{singularSuffix: " enters", pluralSuffix: " enter", event: TriggerEventPermanentEnteredBattlefield},
	{singularSuffix: " enters the battlefield", pluralSuffix: " enter the battlefield", event: TriggerEventPermanentEnteredBattlefield},
	{singularSuffix: " enters under your control", pluralSuffix: " enter under your control", event: TriggerEventPermanentEnteredBattlefield, controller: ControllerYou},
	{singularSuffix: " enters the battlefield under your control", pluralSuffix: " enter the battlefield under your control", event: TriggerEventPermanentEnteredBattlefield, controller: ControllerYou},
	{singularSuffix: " enters under an opponent's control", pluralSuffix: " enter under an opponent's control", event: TriggerEventPermanentEnteredBattlefield, controller: ControllerOpponent},
	{singularSuffix: " enters the battlefield under an opponent's control", pluralSuffix: " enter the battlefield under an opponent's control", event: TriggerEventPermanentEnteredBattlefield, controller: ControllerOpponent},
	{singularSuffix: " enters from a graveyard", pluralSuffix: " enter from a graveyard", event: TriggerEventPermanentEnteredBattlefield, matchFromZone: true, fromZone: TriggerZoneGraveyard},
	{singularSuffix: " enters from your graveyard", pluralSuffix: " enter from your graveyard", event: TriggerEventPermanentEnteredBattlefield, matchFromZone: true, fromZone: TriggerZoneGraveyard, player: TriggerPlayerYou},
	{singularSuffix: " enters from an opponent's graveyard", pluralSuffix: " enter from an opponent's graveyard", event: TriggerEventPermanentEnteredBattlefield, matchFromZone: true, fromZone: TriggerZoneGraveyard, player: TriggerPlayerOpponent},
	{singularSuffix: " enters from exile", pluralSuffix: " enter from exile", event: TriggerEventPermanentEnteredBattlefield, matchFromZone: true, fromZone: TriggerZoneExile},
	{singularSuffix: " enters from your hand", pluralSuffix: " enter from your hand", event: TriggerEventPermanentEnteredBattlefield, matchFromZone: true, fromZone: TriggerZoneHand, player: TriggerPlayerYou},
	{singularSuffix: " dies", pluralSuffix: " die", event: TriggerEventPermanentDied},
	{
		singularSuffix: " is put into a graveyard from the battlefield",
		pluralSuffix:   " are put into a graveyard from the battlefield",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneGraveyard,
	},
	{
		singularSuffix: " is put into a graveyard",
		pluralSuffix:   " are put into a graveyard",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneGraveyard,
	},
	{
		singularSuffix: " is put into your graveyard from the battlefield",
		pluralSuffix:   " are put into your graveyard from the battlefield",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneGraveyard,
		player:         TriggerPlayerYou,
	},
	{
		singularSuffix: " is put into an opponent's graveyard from the battlefield",
		pluralSuffix:   " are put into an opponent's graveyard from the battlefield",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneGraveyard,
		player:         TriggerPlayerOpponent,
	},
	{
		singularSuffix: " is put into its owner's graveyard from the battlefield",
		pluralSuffix:   " are put into their owners' graveyards from the battlefield",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneGraveyard,
	},
	{
		singularSuffix: " leaves the battlefield without dying",
		pluralSuffix:   " leave the battlefield without dying",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		excludeToZone:  true,
		toZone:         TriggerZoneGraveyard,
	},
	{
		singularSuffix: " leaves the battlefield",
		pluralSuffix:   " leave the battlefield",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
	},
	{
		singularSuffix: " is exiled from the battlefield",
		pluralSuffix:   " are exiled from the battlefield",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneExile,
	},
	{
		singularSuffix: " is put into exile from the battlefield",
		pluralSuffix:   " are put into exile from the battlefield",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneExile,
	},
	{
		singularSuffix: " is put into exile",
		pluralSuffix:   " are put into exile",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneExile,
	},
	{
		singularSuffix: " is returned to its owner's hand",
		pluralSuffix:   " are returned to their owners' hands",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneHand,
	},
	{
		singularSuffix: " is returned to your hand",
		pluralSuffix:   " are returned to your hand",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneHand,
		player:         TriggerPlayerYou,
	},
	{
		singularSuffix: " is returned to a player's hand",
		pluralSuffix:   " are returned to a player's hand",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneHand,
	},
	{
		singularSuffix: " is returned to hand",
		pluralSuffix:   " are returned to hand",
		event:          TriggerEventZoneChanged,
		matchFromZone:  true,
		fromZone:       TriggerZoneBattlefield,
		matchToZone:    true,
		toZone:         TriggerZoneHand,
	},
}

var attachedPermanentSubjects = map[string]TriggerSelection{
	"enchanted artifact":    {RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact}},
	"enchanted creature":    {RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}},
	"enchanted enchantment": {RequiredTypes: []TriggerCardType{TriggerCardTypeEnchantment}},
	"enchanted land":        {RequiredTypes: []TriggerCardType{TriggerCardTypeLand}},
	"enchanted permanent":   {},
	"equipped creature":     {RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}},
	"fortified land":        {RequiredTypes: []TriggerCardType{TriggerCardTypeLand}},
}

func recognizeZoneChangeTrigger(event triggerEventSyntax, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhen && kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	for _, template := range permanentZoneChangeTemplates {
		if subject, ok := strings.CutSuffix(event.text, template.singularSuffix); ok {
			pattern, ok := bindSinglePermanentZoneChangeSubject(event, subject)
			if !ok {
				continue
			}
			if !completePermanentZoneChangePattern(&pattern, template) {
				return TriggerPattern{}, false
			}
			return pattern, true
		}
		if subject, ok := strings.CutSuffix(event.text, template.pluralSuffix); ok {
			pattern, ok := bindPluralPermanentZoneChangeSubject(event, subject)
			if !ok {
				continue
			}
			if !completePermanentZoneChangePattern(&pattern, template) {
				return TriggerPattern{}, false
			}
			return pattern, true
		}
	}
	return TriggerPattern{}, false
}

func bindSinglePermanentZoneChangeSubject(event triggerEventSyntax, subject string) (TriggerPattern, bool) {
	if event.selfSubject(subject, selfEnterSubjectSlots, true) {
		return TriggerPattern{Source: TriggerSourceSelf}, true
	}
	if selection, ok := parseAttachedPermanentZoneChangeSubject(event, subject); ok {
		return TriggerPattern{Source: TriggerSourceAttachedPermanent, SubjectSelection: selection}, true
	}

	subject, otherThanSelf := stripOtherThanSelfSubject(event, subject)
	parsed, ok := parseZoneChangePermanentSubject(event, subject, false)
	if !ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Controller:       parsed.controller,
		Player:           parsed.player,
		ExcludeSelf:      parsed.excludeSelf || otherThanSelf,
		SubjectSelection: parsed.selection,
		MatchFaceDown:    parsed.faceDown,
		FaceDown:         parsed.faceDown,
	}, true
}

func stripOtherThanSelfSubject(event triggerEventSyntax, subject string) (string, bool) {
	before, excluded, ok := strings.Cut(subject, " other than ")
	if !ok {
		return subject, false
	}
	if event.selfSubject(excluded, selfEnterSubjectSlots, true) {
		return before, true
	}
	return subject, false
}

func parseAttachedPermanentZoneChangeSubject(event triggerEventSyntax, subject string) (TriggerSelection, bool) {
	if selection, ok := attachedPermanentSubjects[subject]; ok {
		return selection, true
	}
	for _, prefix := range []string{"enchanted ", "equipped ", "fortified "} {
		rest, ok := strings.CutPrefix(subject, prefix)
		if !ok {
			continue
		}
		selection, ok := parsePermanentTriggerSelection(event, rest, false)
		if !ok {
			return TriggerSelection{}, false
		}
		switch prefix {
		case "equipped ":
			addRequiredTriggerType(&selection, TriggerCardTypeCreature)
		case "fortified ":
			addRequiredTriggerType(&selection, TriggerCardTypeLand)
		default:
		}
		return selection, true
	}
	return TriggerSelection{}, false
}

func bindPluralPermanentZoneChangeSubject(event triggerEventSyntax, subject string) (TriggerPattern, bool) {
	if event.selfSubject(subject, nil, true) {
		return TriggerPattern{Source: TriggerSourceSelf}, true
	}
	subject, oneOrMore := strings.CutPrefix(subject, "one or more ")
	if !oneOrMore {
		return TriggerPattern{}, false
	}
	parsed, ok := parseZoneChangePermanentSubject(event, subject, true)
	if !ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Controller:       parsed.controller,
		Player:           parsed.player,
		ExcludeSelf:      parsed.excludeSelf,
		SubjectSelection: parsed.selection,
		MatchFaceDown:    parsed.faceDown,
		FaceDown:         parsed.faceDown,
		OneOrMore:        true,
	}, true
}

func completePermanentZoneChangePattern(pattern *TriggerPattern, template permanentZoneChangeTemplate) bool {
	if !mergeControllerKind(&pattern.Controller, template.controller) ||
		!mergeTriggerPlayerRelation(&pattern.Player, template.player) {
		return false
	}
	pattern.Event = template.event
	pattern.MatchFromZone = template.matchFromZone
	pattern.FromZone = template.fromZone
	pattern.MatchToZone = template.matchToZone
	pattern.ToZone = template.toZone
	pattern.ExcludeToZone = template.excludeToZone
	if template.tapped != TriggerTriAny {
		if pattern.SubjectSelection.Tapped != TriggerTriAny &&
			pattern.SubjectSelection.Tapped != template.tapped {
			return false
		}
		pattern.SubjectSelection.Tapped = template.tapped
	}
	if template.event == TriggerEventPermanentDied {
		addRequiredTriggerType(&pattern.SubjectSelection, TriggerCardTypeCreature)
	}
	return true
}

func mergeControllerKind(current *ControllerKind, additional ControllerKind) bool {
	if additional == ControllerAny {
		return true
	}
	if *current != ControllerAny && *current != additional {
		return false
	}
	*current = additional
	return true
}

func mergeTriggerPlayerRelation(current *TriggerPlayerRelation, additional TriggerPlayerRelation) bool {
	if additional == TriggerPlayerAny {
		return true
	}
	if *current != TriggerPlayerAny && *current != additional {
		return false
	}
	*current = additional
	return true
}

func addRequiredTriggerType(selection *TriggerSelection, cardType TriggerCardType) {
	if !slices.Contains(selection.RequiredTypes, cardType) {
		selection.RequiredTypes = append(selection.RequiredTypes, cardType)
	}
}

type zoneChangePermanentSubject struct {
	selection   TriggerSelection
	controller  ControllerKind
	player      TriggerPlayerRelation
	excludeSelf bool
	faceDown    bool
}

func parseZoneChangePermanentSubject(event triggerEventSyntax, subject string, plural bool) (zoneChangePermanentSubject, bool) {
	result := zoneChangePermanentSubject{}
	relations := parsePermanentSubjectRelations(subject)
	if !relations.ok {
		return zoneChangePermanentSubject{}, false
	}
	subject = relations.subject
	result.controller = relations.controller
	result.player = relations.player
	if plural {
		subject, result.excludeSelf = strings.CutPrefix(subject, "other ")
	} else {
		switch {
		case strings.HasPrefix(subject, "another "):
			result.excludeSelf = true
			subject = strings.TrimPrefix(subject, "another ")
		case strings.HasPrefix(subject, "a "):
			subject = strings.TrimPrefix(subject, "a ")
		case strings.HasPrefix(subject, "an "):
			subject = strings.TrimPrefix(subject, "an ")
		default:
			return zoneChangePermanentSubject{}, false
		}
	}
	if rest, ok := strings.CutPrefix(subject, "face-down "); ok {
		subject = rest
		result.faceDown = true
	}
	var ok bool
	result.selection, ok = parsePermanentTriggerSelection(event, subject, plural)
	if !ok {
		return zoneChangePermanentSubject{}, false
	}
	return result, true
}

type permanentSubjectRelations struct {
	subject    string
	controller ControllerKind
	player     TriggerPlayerRelation
	ok         bool
}

func parsePermanentSubjectRelations(subject string) permanentSubjectRelations {
	controller := ControllerAny
	player := TriggerPlayerAny
	for _, relation := range []struct {
		phrase     string
		controller ControllerKind
	}{
		{phrase: " you control", controller: ControllerYou},
		{phrase: " an opponent controls", controller: ControllerOpponent},
		{phrase: " your opponents control", controller: ControllerOpponent},
		{phrase: " you don't control", controller: ControllerOpponent},
	} {
		for _, qualifier := range []string{" with ", " without "} {
			if strings.Contains(subject, relation.phrase+qualifier) {
				subject = strings.Replace(subject, relation.phrase+qualifier, qualifier, 1)
				controller = relation.controller
				break
			}
		}
	}
	switch {
	case strings.HasSuffix(subject, " you control but don't own"):
		subject = strings.TrimSuffix(subject, " you control but don't own")
		controller = ControllerYou
		player = TriggerPlayerOpponent
	case strings.HasSuffix(subject, " you control"):
		subject = strings.TrimSuffix(subject, " you control")
		controller = ControllerYou
	case strings.HasSuffix(subject, " an opponent controls"):
		subject = strings.TrimSuffix(subject, " an opponent controls")
		controller = ControllerOpponent
	case strings.HasSuffix(subject, " your opponents control"):
		subject = strings.TrimSuffix(subject, " your opponents control")
		controller = ControllerOpponent
	case strings.HasSuffix(subject, " you don't control"):
		subject = strings.TrimSuffix(subject, " you don't control")
		controller = ControllerOpponent
	case strings.HasSuffix(subject, " you own"):
		subject = strings.TrimSuffix(subject, " you own")
		player = TriggerPlayerYou
	case strings.HasSuffix(subject, " an opponent owns"):
		subject = strings.TrimSuffix(subject, " an opponent owns")
		player = TriggerPlayerOpponent
	case strings.HasSuffix(subject, " owned by another player"):
		subject = strings.TrimSuffix(subject, " owned by another player")
		player = TriggerPlayerOpponent
	default:
	}
	return permanentSubjectRelations{subject: subject, controller: controller, player: player, ok: subject != ""}
}

func parsePermanentTriggerSelection(event triggerEventSyntax, subject string, plural bool) (TriggerSelection, bool) {
	selection := TriggerSelection{}
	if plural && strings.HasSuffix(subject, " tokens") {
		subject = strings.TrimSuffix(subject, " tokens")
		selection.TokenOnly = true
	} else if !plural && strings.HasSuffix(subject, " token") {
		subject = strings.TrimSuffix(subject, " token")
		selection.TokenOnly = true
	}
	if plural && strings.HasSuffix(subject, " cards") {
		subject = strings.TrimSuffix(subject, " cards")
	} else if !plural && strings.HasSuffix(subject, " card") {
		subject = strings.TrimSuffix(subject, " card")
	}
	var ok bool
	subject, ok = parsePermanentTriggerSelectionSuffix(subject, &selection)
	if !ok {
		return TriggerSelection{}, false
	}
	for {
		rest, matched := parsePermanentTriggerSelectionAdjective(event, subject, &selection)
		if !matched {
			break
		}
		subject = rest
	}
	for _, separator := range []string{" and/or ", " or "} {
		left, right, ok := strings.Cut(subject, separator)
		if !ok {
			continue
		}
		leftType, leftOK := parseSingleTriggerPermanentType(event, left, plural)
		rightType, rightOK := parseSingleTriggerPermanentType(event, right, plural)
		if leftOK && rightOK && leftType != TriggerCardTypeUnknown && rightType != TriggerCardTypeUnknown {
			selection.RequiredTypesAny = []TriggerCardType{leftType, rightType}
			return selection, true
		}
		if leftOK || rightOK {
			return TriggerSelection{}, false
		}
		leftSubtype, leftSubtypeOK := triggerSubtypeFromPhrase(event, left)
		rightSubtype, rightSubtypeOK := triggerSubtypeFromPhrase(event, right)
		if !leftSubtypeOK || !rightSubtypeOK {
			return TriggerSelection{}, false
		}
		selection.SubtypesAny = []TriggerSubtype{
			leftSubtype,
			rightSubtype,
		}
		return selection, true
	}
	if strings.HasPrefix(subject, "non") {
		excluded, rest, ok := parseExcludedTriggerPermanentType(event, subject, plural)
		if !ok {
			return TriggerSelection{}, false
		}
		selection.ExcludedTypes = []TriggerCardType{excluded}
		subject = rest
	}
	words := strings.Fields(subject)
	if len(words) == 0 {
		return TriggerSelection{}, false
	}
	if words[len(words)-1] == "token" || words[len(words)-1] == "tokens" {
		selection.TokenOnly = true
		words = words[:len(words)-1]
	}
	if len(words) == 0 {
		return selection, true
	}
	var subtypeWords []string
	for _, word := range words {
		cardType, ok := parseSingleTriggerPermanentType(event, word, plural)
		if !ok {
			subtypeWords = append(subtypeWords, word)
			continue
		}
		if cardType != TriggerCardTypeUnknown {
			addRequiredTriggerType(&selection, cardType)
		}
	}
	if len(subtypeWords) > 0 {
		subtypeText := strings.Join(subtypeWords, " ")
		if subtypeText == "outlaw" {
			selection.SubtypesAny = []TriggerSubtype{types.Assassin, types.Mercenary, types.Pirate, types.Rogue, types.Warlock}
			return selection, true
		}
		subtype, ok := triggerSubtypeFromPhrase(event, subtypeText)
		if !ok {
			return TriggerSelection{}, false
		}
		selection.SubtypesAny = []TriggerSubtype{subtype}
	}
	return selection, true
}

func parsePermanentTriggerSelectionAdjective(event triggerEventSyntax, subject string, selection *TriggerSelection) (string, bool) {
	adjectives := []string{
		"nontoken ",
		"token ",
		"legendary ",
		"snow ",
		"white ",
		"blue ",
		"black ",
		"red ",
		"green ",
		"nonwhite ",
		"nonblue ",
		"nonblack ",
		"nonred ",
		"nongreen ",
		"colorless ",
		"multicolored ",
		"attacking ",
		"blocking ",
	}
	for _, adjective := range adjectives {
		if rest, ok := strings.CutPrefix(subject, adjective); ok {
			word := strings.TrimSpace(adjective)
			if !bindTriggerSelectionAdjective(event, word, selection) {
				return "", false
			}
			return rest, true
		}
	}
	return subject, false
}

func parsePermanentTriggerSelectionSuffix(subject string, selection *TriggerSelection) (string, bool) {
	switch {
	case strings.HasSuffix(subject, " tapped"):
		selection.Tapped = TriggerTriTrue
		return strings.TrimSuffix(subject, " tapped"), true
	case strings.HasSuffix(subject, " untapped"):
		selection.Tapped = TriggerTriFalse
		return strings.TrimSuffix(subject, " untapped"), true
	}
	if before, qualifier, ok := strings.Cut(subject, " with "); ok {
		if !parsePermanentTriggerSelectionQualifier(qualifier, selection) {
			return "", false
		}
		return before, true
	}
	if before, qualifier, ok := strings.Cut(subject, " without "); ok {
		keyword, ok := parseTriggerKeyword(qualifier)
		if !ok {
			return "", false
		}
		selection.ExcludedKeyword = keyword
		return before, true
	}
	fields := strings.Fields(subject)
	if len(fields) > 1 {
		powerText, toughnessText, hasSlash := strings.Cut(fields[0], "/")
		power, powerOK := parseTriggerNonNegativeInt(powerText)
		toughness, toughnessOK := parseTriggerNonNegativeInt(toughnessText)
		if hasSlash && powerOK && toughnessOK {
			selection.Power = TriggerNumberFilter{Comparison: TriggerComparisonEqual, Value: power}
			selection.Toughness = TriggerNumberFilter{Comparison: TriggerComparisonEqual, Value: toughness}
			return strings.Join(fields[1:], " "), true
		}
	}
	return subject, true
}

func parsePermanentTriggerSelectionQualifier(qualifier string, selection *TriggerSelection) bool {
	if keyword, ok := parseTriggerKeyword(qualifier); ok {
		selection.Keyword = keyword
		return true
	}
	for _, characteristic := range []struct {
		prefix string
		bind   func(TriggerNumberFilter)
	}{
		{prefix: "mana value ", bind: func(filter TriggerNumberFilter) { selection.ManaValue = filter }},
		{prefix: "power ", bind: func(filter TriggerNumberFilter) { selection.Power = filter }},
		{prefix: "toughness ", bind: func(filter TriggerNumberFilter) { selection.Toughness = filter }},
	} {
		if rest, ok := strings.CutPrefix(qualifier, characteristic.prefix); ok {
			filter, ok := parseTriggerNumberFilter(rest)
			if !ok {
				return false
			}
			characteristic.bind(filter)
			return true
		}
	}
	return false
}

func parseTriggerKeyword(word string) (TriggerKeyword, bool) {
	switch word {
	case "defender":
		return TriggerKeywordDefender, true
	case "flash":
		return TriggerKeywordFlash, true
	case "flying":
		return TriggerKeywordFlying, true
	case "haste":
		return TriggerKeywordHaste, true
	case "shadow":
		return TriggerKeywordShadow, true
	default:
		return TriggerKeywordUnknown, false
	}
}

func parseTriggerNumberFilter(phrase string) (TriggerNumberFilter, bool) {
	comparison := TriggerComparisonEqual
	switch {
	case strings.HasSuffix(phrase, " or less"):
		comparison = TriggerComparisonAtMost
		phrase = strings.TrimSuffix(phrase, " or less")
	case strings.HasSuffix(phrase, " or greater"):
		comparison = TriggerComparisonAtLeast
		phrase = strings.TrimSuffix(phrase, " or greater")
	default:
	}
	value, ok := parseTriggerNonNegativeInt(phrase)
	return TriggerNumberFilter{Comparison: comparison, Value: value}, ok
}

func parseTriggerNonNegativeInt(phrase string) (int, bool) {
	if phrase == "" {
		return 0, false
	}
	value := 0
	for _, r := range phrase {
		if r < '0' || r > '9' {
			return 0, false
		}
		value = value*10 + int(r-'0')
	}
	return value, true
}

func parseExcludedTriggerPermanentType(event triggerEventSyntax, subject string, plural bool) (TriggerCardType, string, bool) {
	for _, word := range []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker"} {
		prefix := "non" + word + " "
		if strings.HasPrefix(subject, prefix) {
			cardType, ok := parseSingleTriggerPermanentType(event, "non"+word, plural)
			return cardType, strings.TrimPrefix(subject, prefix), ok
		}
	}
	return TriggerCardTypeUnknown, "", false
}

func parseSingleTriggerPermanentType(event triggerEventSyntax, word string, _ bool) (TriggerCardType, bool) {
	word = strings.TrimPrefix(strings.TrimPrefix(word, "a "), "an ")
	excluded := strings.HasPrefix(word, "non")
	for _, token := range event.tokens {
		if !strings.EqualFold(token.Text, word) {
			continue
		}
		if excluded {
			if cardType, ok := event.atoms.ExcludedCardTypeAt(token.Span); ok {
				return triggerCardTypeFromParser(cardType)
			}
			return TriggerCardTypeUnknown, false
		}
		if cardType, ok := event.atoms.CardTypeAt(token.Span); ok {
			return triggerCardTypeFromParser(cardType)
		}
		if noun, ok := event.atoms.ObjectNounAt(token.Span); ok && noun == parser.ObjectNounPermanent {
			return TriggerCardTypeUnknown, true
		}
	}
	return TriggerCardTypeUnknown, false
}

func triggerCardTypeFromParser(cardType parser.CardType) (TriggerCardType, bool) {
	switch cardType {
	case parser.CardTypeArtifact:
		return TriggerCardTypeArtifact, true
	case parser.CardTypeBattle:
		return TriggerCardTypeBattle, true
	case parser.CardTypeCreature:
		return TriggerCardTypeCreature, true
	case parser.CardTypeEnchantment:
		return TriggerCardTypeEnchantment, true
	case parser.CardTypeInstant:
		return TriggerCardTypeInstant, true
	case parser.CardTypeLand:
		return TriggerCardTypeLand, true
	case parser.CardTypePlaneswalker:
		return TriggerCardTypePlaneswalker, true
	case parser.CardTypeSorcery:
		return TriggerCardTypeSorcery, true
	default:
		return TriggerCardTypeUnknown, false
	}
}

func triggerSubtypeFromPhrase(event triggerEventSyntax, phrase string) (types.Sub, bool) {
	tokens := event.tokensFor(phrase)
	if len(tokens) == 0 {
		return "", false
	}
	span := shared.SpanOf(tokens)
	if sub, ok := event.atoms.SubtypeAt(span); ok {
		return sub, true
	}
	return "", false
}

func bindTriggerSelectionAdjective(event triggerEventSyntax, word string, selection *TriggerSelection) bool {
	for _, token := range event.tokens {
		if !strings.EqualFold(token.Text, word) {
			continue
		}
		if color, ok := event.atoms.ExcludedColorAt(token.Span); ok {
			compiled, ok := triggerColorFromParser(color)
			if !ok {
				return false
			}
			selection.ExcludedColors = append(selection.ExcludedColors, compiled)
			return true
		}
		if color, ok := event.atoms.ColorAt(token.Span); ok {
			compiled, ok := triggerColorFromParser(color)
			if !ok {
				return false
			}
			selection.ColorsAny = append(selection.ColorsAny, compiled)
			return true
		}
		if qualifier, ok := event.atoms.ColorQualifierAt(token.Span); ok {
			switch qualifier {
			case parser.ColorQualifierColorless:
				selection.Colorless = true
			case parser.ColorQualifierMulticolored:
				selection.Multicolored = true
			default:
				return false
			}
			return true
		}
		if supertype, ok := event.atoms.SupertypeAt(token.Span); ok {
			compiled, ok := triggerSupertypeFromParser(supertype)
			if !ok {
				return false
			}
			selection.Supertypes = append(selection.Supertypes, compiled)
			return true
		}
		if event.atoms.SelectionFlagIn(token.Span, parser.SelectionFlagAttacking) {
			selection.CombatState = TriggerCombatStateAttacking
			return true
		}
		if event.atoms.SelectionFlagIn(token.Span, parser.SelectionFlagBlocking) {
			selection.CombatState = TriggerCombatStateBlocking
			return true
		}
		if event.atoms.SelectionFlagIn(token.Span, parser.SelectionFlagToken) {
			selection.TokenOnly = true
			return true
		}
		if event.atoms.SelectionFlagIn(token.Span, parser.SelectionFlagNonToken) {
			selection.NonToken = true
			return true
		}
	}
	return false
}

func triggerSupertypeFromParser(supertype parser.Supertype) (TriggerSupertype, bool) {
	switch supertype {
	case parser.SupertypeLegendary:
		return TriggerSupertypeLegendary, true
	case parser.SupertypeSnow:
		return TriggerSupertypeSnow, true
	default:
		return TriggerSupertypeUnknown, false
	}
}

type permanentActionTemplate struct {
	suffix    string
	event     TriggerEvent
	allowWhen bool
}

var combatPermanentActions = []permanentActionTemplate{
	{suffix: " attacks", event: TriggerEventAttackerDeclared},
	{suffix: " blocks", event: TriggerEventBlockerDeclared},
}

var statePermanentActions = []permanentActionTemplate{
	{suffix: " becomes tapped", event: TriggerEventPermanentTapped},
	{suffix: " becomes untapped", event: TriggerEventPermanentUntapped},
	{suffix: " is turned face up", event: TriggerEventPermanentTurnedFaceUp, allowWhen: true},
}

func recognizePermanentActionTrigger(event triggerEventSyntax, kind TriggerKind, actions []permanentActionTemplate) (TriggerPattern, bool) {
	for _, action := range actions {
		if kind != TriggerWhenever && (kind != TriggerWhen || !action.allowWhen) {
			continue
		}
		if !strings.HasSuffix(event.text, action.suffix) {
			continue
		}
		subject := strings.TrimSuffix(event.text, action.suffix)
		attachedSubject, attached := strings.CutPrefix(subject, "enchanted ")
		if subject == "equipped creature" {
			attachedSubject = "creature"
			attached = true
		}
		if attached {
			selection, ok := parsePermanentTriggerSelection(event, attachedSubject, false)
			if !ok {
				return TriggerPattern{}, false
			}
			return TriggerPattern{
				Event:            action.event,
				Source:           TriggerSourceAttachedPermanent,
				SubjectSelection: selection,
			}, true
		}
		parsed, ok := parseSinglePermanentEventSubject(event, action.suffix)
		if !ok {
			return TriggerPattern{}, false
		}
		if (action.event == TriggerEventAttackerDeclared || action.event == TriggerEventBlockerDeclared) &&
			!slices.Contains(parsed.selection.RequiredTypes, TriggerCardTypeCreature) {
			return TriggerPattern{}, false
		}
		return TriggerPattern{
			Event:            action.event,
			Controller:       parsed.controller,
			ExcludeSelf:      parsed.excludeSelf,
			SubjectSelection: parsed.selection,
		}, true
	}
	return TriggerPattern{}, false
}

type permanentEventSubject struct {
	selection   TriggerSelection
	controller  ControllerKind
	excludeSelf bool
}

func parseSinglePermanentEventSubject(event triggerEventSyntax, suffix string) (permanentEventSubject, bool) {
	if !strings.HasSuffix(event.text, suffix) {
		return permanentEventSubject{}, false
	}
	parsed := parseCombatPermanentSelection(event, strings.TrimSuffix(event.text, suffix), false)
	if !parsed.ok {
		return permanentEventSubject{}, false
	}
	return permanentEventSubject{
		selection:   parsed.selection,
		controller:  parsed.controller,
		excludeSelf: parsed.excludeSelf,
	}, true
}
