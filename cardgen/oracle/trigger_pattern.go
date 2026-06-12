package oracle

import (
	"slices"
	"strings"
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

// TriggerSubtype identifies a subtype used by a semantic trigger Selection.
// The runtime adapter validates values against its closed subtype vocabulary.
type TriggerSubtype string

// Trigger subtypes.
const (
	TriggerSubtypeUnknown TriggerSubtype = ""
	TriggerSubtypeSpirit  TriggerSubtype = "Spirit"
	TriggerSubtypeArcane  TriggerSubtype = "Arcane"
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
	Span Span
	Kind TriggerKind

	Event      TriggerEvent
	Source     TriggerSourceRelation
	Subject    TriggerSubject
	Controller ControllerKind
	Player     TriggerPlayerRelation

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
	AttackRecipient                   TriggerAttackRecipient
	StackObject                       TriggerStackObject
	Counter                           TriggerCounter

	ExcludeSelf              bool
	OneOrMore                bool
	OneOrMorePerAttackTarget bool
	RequireKickerPaid        bool
	RequireHistoric          bool

	InterveningCondition *CompiledCondition
}

type triggerPatternTemplate struct {
	kinds []TriggerKind
	bind  func(string, TriggerKind, string) (TriggerPattern, bool)
}

var triggerPatternTemplates = []triggerPatternTemplate{
	{kinds: []TriggerKind{TriggerAt}, bind: recognizePhaseStepTrigger},
	{kinds: []TriggerKind{TriggerWhen, TriggerWhenever}, bind: recognizePermanentZoneChangeTrigger},
	{kinds: []TriggerKind{TriggerWhenever}, bind: recognizeSpellAbilityTrigger},
	{kinds: []TriggerKind{TriggerWhen, TriggerWhenever}, bind: recognizeCombatTrigger},
	{kinds: []TriggerKind{TriggerWhenever}, bind: recognizePermanentStateTrigger},
	{kinds: []TriggerKind{TriggerWhenever}, bind: recognizePlayerEventTrigger},
}

func compileTriggerPattern(
	event string,
	kind TriggerKind,
	span Span,
	cardName string,
	condition *CompiledCondition,
) TriggerPattern {
	event = strings.ToLower(event)
	pattern := TriggerPattern{
		Span:                 span,
		Kind:                 kind,
		InterveningCondition: condition,
	}
	recognized, ok := bindTriggerPatternTemplates(event, kind, cardName, triggerPatternTemplates)
	if ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	return pattern
}

func bindTriggerPatternTemplates(
	event string,
	kind TriggerKind,
	cardName string,
	templates []triggerPatternTemplate,
) (TriggerPattern, bool) {
	var recognized TriggerPattern
	matched := false
	for _, template := range templates {
		if !slices.Contains(template.kinds, kind) {
			continue
		}
		candidate, ok := template.bind(event, kind, cardName)
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

type controllerPhraseSlot struct {
	text       string
	controller ControllerKind
}

type phaseStepTemplate struct {
	prefix    string
	suffix    string
	step      TriggerStep
	relations []controllerPhraseSlot
}

var standardStepControllerSlots = []controllerPhraseSlot{
	{text: "your", controller: ControllerYou},
	{text: "its controller's", controller: ControllerYou},
	{text: "the", controller: ControllerAny},
	{text: "each", controller: ControllerAny},
	{text: "each player's", controller: ControllerAny},
	{text: "each opponent's", controller: ControllerOpponent},
}

var phaseStepTemplates = []phaseStepTemplate{
	{suffix: " upkeep", step: TriggerStepUpkeep, relations: standardStepControllerSlots},
	{suffix: " draw step", step: TriggerStepDraw, relations: standardStepControllerSlots},
	{suffix: " end step", step: TriggerStepEnd, relations: standardStepControllerSlots},
	{
		prefix: "combat on ",
		suffix: " turn",
		step:   TriggerStepBeginningOfCombat,
		relations: []controllerPhraseSlot{
			{text: "your", controller: ControllerYou},
			{text: "each", controller: ControllerAny},
			{text: "each opponent's", controller: ControllerOpponent},
		},
	},
	{
		suffix:    " combat",
		step:      TriggerStepBeginningOfCombat,
		relations: []controllerPhraseSlot{{text: "each", controller: ControllerAny}},
	},
	{
		suffix:    " end of combat",
		step:      TriggerStepEndOfCombat,
		relations: []controllerPhraseSlot{{text: "the", controller: ControllerAny}},
	},
	{
		prefix:    "the end of combat on ",
		suffix:    " turn",
		step:      TriggerStepEndOfCombat,
		relations: []controllerPhraseSlot{{text: "your", controller: ControllerYou}},
	},
	{
		suffix:    " end of combat step",
		step:      TriggerStepEndOfCombat,
		relations: []controllerPhraseSlot{{text: "each", controller: ControllerAny}},
	},
	{
		suffix: " precombat main phase",
		step:   TriggerStepPrecombatMain,
		relations: []controllerPhraseSlot{
			{text: "your", controller: ControllerYou},
			{text: "each player's", controller: ControllerAny},
			{text: "each opponent's", controller: ControllerOpponent},
		},
	},
	{
		suffix: " postcombat main phase",
		step:   TriggerStepPostcombatMain,
		relations: []controllerPhraseSlot{
			{text: "your", controller: ControllerYou},
			{text: "each player's", controller: ControllerAny},
			{text: "each opponent's", controller: ControllerOpponent},
		},
	},
}

var phaseStepAliases = map[string]TriggerPattern{
	"your first main phase":               {Step: TriggerStepPrecombatMain, Controller: ControllerYou},
	"each of your first main phases":      {Step: TriggerStepPrecombatMain, Controller: ControllerYou},
	"each player's first main phase":      {Step: TriggerStepPrecombatMain, Controller: ControllerAny},
	"each opponent's first main phase":    {Step: TriggerStepPrecombatMain, Controller: ControllerOpponent},
	"your second main phase":              {Step: TriggerStepPostcombatMain, Controller: ControllerYou},
	"each player's second main phase":     {Step: TriggerStepPostcombatMain, Controller: ControllerAny},
	"each opponent's second main phase":   {Step: TriggerStepPostcombatMain, Controller: ControllerOpponent},
	"each of your postcombat main phases": {Step: TriggerStepPostcombatMain, Controller: ControllerYou},
	"your combat step":                    {Step: TriggerStepBeginningOfCombat, Controller: ControllerYou},
}

func recognizePhaseStepTrigger(event string, _ TriggerKind, _ string) (TriggerPattern, bool) {
	if event == "end of combat" || event == "the end of combat" {
		return TriggerPattern{
			Event:      TriggerEventBeginningOfStep,
			Step:       TriggerStepEndOfCombat,
			Controller: ControllerAny,
		}, true
	}
	event, ok := strings.CutPrefix(event, "the beginning of ")
	if !ok {
		return TriggerPattern{}, false
	}
	if pattern, ok := phaseStepAliases[event]; ok {
		pattern.Event = TriggerEventBeginningOfStep
		return pattern, true
	}
	if pattern, ok := recognizeAttachedControllerPhaseStep(event); ok {
		return pattern, true
	}
	for _, template := range phaseStepTemplates {
		slot, ok := strings.CutPrefix(event, template.prefix)
		if !ok {
			continue
		}

		slot, ok = strings.CutSuffix(slot, template.suffix)
		if !ok {
			continue
		}
		for _, relation := range template.relations {
			if slot == relation.text {
				return TriggerPattern{
					Event:      TriggerEventBeginningOfStep,
					Step:       template.step,
					Controller: relation.controller,
				}, true
			}
		}
	}
	return TriggerPattern{}, false
}

func recognizeAttachedControllerPhaseStep(event string) (TriggerPattern, bool) {
	for _, template := range []struct {
		prefix string
		step   TriggerStep
	}{
		{prefix: "the upkeep of enchanted ", step: TriggerStepUpkeep},
		{prefix: "the draw step of enchanted ", step: TriggerStepDraw},
		{prefix: "the end step of enchanted ", step: TriggerStepEnd},
	} {
		subject, ok := strings.CutPrefix(event, template.prefix)
		if !ok {
			continue
		}
		subject, ok = strings.CutSuffix(subject, "'s controller")
		if !ok {
			return TriggerPattern{}, false
		}
		parsed := parseCombatPermanentSelection("a "+subject, false)
		if !parsed.ok {
			return TriggerPattern{}, false
		}
		return TriggerPattern{
			Event:                             TriggerEventBeginningOfStep,
			Step:                              template.step,
			StepPlayerSourceAttachedSelection: parsed.selection,
		}, true
	}
	return TriggerPattern{}, false
}

func recognizePermanentZoneChangeTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	return recognizeZoneChangeTrigger(event, kind, cardName)
}

func recognizeSpellAbilityTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	if pattern, ok := recognizeCastTrigger(event, kind); ok {
		return pattern, true
	}
	return recognizeBecameTargetTrigger(event, cardName)
}

func recognizeCombatTrigger(event string, kind TriggerKind, sourceName string) (TriggerPattern, bool) {
	if pattern, ok := recognizeAttackBlockTrigger(event, sourceName); ok {
		return pattern, true
	}
	if pattern, ok := recognizeParameterizedDamageTrigger(event, sourceName); ok {
		return pattern, true
	}
	return recognizePermanentActionTrigger(event, kind, combatPermanentActions)
}

func recognizeAttackBlockTrigger(event, sourceName string) (TriggerPattern, bool) {
	if pattern, ok := recognizePlayerAttackTrigger(event); ok {
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
		subject, remainder, ok := strings.Cut(event, template.marker)
		if !ok {
			continue
		}
		pattern, ok := combatSubjectPattern(subject, sourceName, template.event, template.plural)
		if !ok {
			return TriggerPattern{}, false
		}
		if template.related {
			related, ok := parseRelatedCombatSelection(remainder)
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
		recipient := parseAttackRecipient(remainder)
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
		if !strings.HasSuffix(event, template.suffix) {
			continue
		}
		return combatSubjectPattern(strings.TrimSuffix(event, template.suffix), sourceName, template.event, template.plural)
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

func combatSubjectPattern(subject, sourceName string, event TriggerEvent, plural bool) (TriggerPattern, bool) {
	oneOrMore := false
	if rest, ok := strings.CutPrefix(subject, "one or more "); ok {
		subject = rest
		plural = true
		oneOrMore = true
	}
	if matchesSelfSubjectSlot(subject, sourceName, selfCombatSubjectSlots, true) {
		return TriggerPattern{
			Event:     event,
			Source:    TriggerSourceSelf,
			OneOrMore: oneOrMore,
		}, true
	}
	if subject == "enchanted creature" || subject == "equipped creature" {
		return TriggerPattern{
			Event:     event,
			Source:    TriggerSourceAttachedPermanent,
			OneOrMore: oneOrMore,
			SubjectSelection: TriggerSelection{
				RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
			},
		}, true
	}
	parsed := parseCombatPermanentSelection(subject, plural)
	if !parsed.ok {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Event:            event,
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

func parseCombatPermanentSelection(subject string, plural bool) combatPermanentSelection {
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
	selection, ok := parsePermanentTriggerSelection(subject, plural)
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

func parseRelatedCombatSelection(subject string) (TriggerSelection, bool) {
	parsed := parseCombatPermanentSelection(subject, false)
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

func parseAttackRecipient(recipient string) attackRecipientPattern {
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
	parsed := parseCombatPermanentSelection(recipient, false)
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

func recognizePermanentStateTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	if pattern, ok := recognizeSelfPermanentStateTrigger(event, kind, cardName); ok {
		return pattern, true
	}
	return recognizePermanentActionTrigger(event, kind, statePermanentActions)
}

func recognizePlayerEventTrigger(event string, kind TriggerKind, _ string) (TriggerPattern, bool) {
	return recognizeSimpleTrigger(event, kind)
}

type playerEventTemplate struct {
	suffix      string
	event       TriggerEvent
	relations   []TriggerPlayerRelation
	oneOrMore   bool
	excludeSelf bool
}

var playerRelationSlots = []struct {
	text     string
	relation TriggerPlayerRelation
}{
	{text: "you", relation: TriggerPlayerYou},
	{text: "an opponent", relation: TriggerPlayerOpponent},
	{text: "a player", relation: TriggerPlayerAny},
}

var playerEventTemplates = []playerEventTemplate{
	{suffix: " draw a card", event: TriggerEventCardDrawn, relations: []TriggerPlayerRelation{TriggerPlayerYou}},
	{suffix: " draws a card", event: TriggerEventCardDrawn, relations: []TriggerPlayerRelation{TriggerPlayerOpponent, TriggerPlayerAny}},
	{suffix: " discard a card", event: TriggerEventCardDiscarded, relations: []TriggerPlayerRelation{TriggerPlayerYou}},
	{suffix: " discards a card", event: TriggerEventCardDiscarded, relations: []TriggerPlayerRelation{TriggerPlayerOpponent, TriggerPlayerAny}},
	{suffix: " discard one or more cards", event: TriggerEventCardDiscarded, relations: []TriggerPlayerRelation{TriggerPlayerYou}, oneOrMore: true},
	{suffix: " cycle a card", event: TriggerEventCycled, relations: []TriggerPlayerRelation{TriggerPlayerYou}},
	{suffix: " cycle another card", event: TriggerEventCycled, relations: []TriggerPlayerRelation{TriggerPlayerYou}, excludeSelf: true},
	{suffix: " cycle or discard a card", event: TriggerEventCardDiscarded, relations: []TriggerPlayerRelation{TriggerPlayerYou}},
	{suffix: " cycle or discard another card", event: TriggerEventCardDiscarded, relations: []TriggerPlayerRelation{TriggerPlayerYou}, excludeSelf: true},
	{suffix: " gain life", event: TriggerEventLifeGained, relations: []TriggerPlayerRelation{TriggerPlayerYou}},
	{suffix: " gains life", event: TriggerEventLifeGained, relations: []TriggerPlayerRelation{TriggerPlayerOpponent}},
	{suffix: " lose life", event: TriggerEventLifeLost, relations: []TriggerPlayerRelation{TriggerPlayerYou}},
	{suffix: " loses life", event: TriggerEventLifeLost, relations: []TriggerPlayerRelation{TriggerPlayerOpponent}},
}

func recognizeSimpleTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	for _, template := range playerEventTemplates {
		relationText, ok := strings.CutSuffix(event, template.suffix)
		if !ok {
			continue
		}
		for _, slot := range playerRelationSlots {
			if relationText != slot.text || !slices.Contains(template.relations, slot.relation) {
				continue
			}
			return TriggerPattern{
				Event:       template.event,
				Player:      slot.relation,
				OneOrMore:   template.oneOrMore,
				ExcludeSelf: template.excludeSelf,
			}, true
		}
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
	"this land",
	"this artifact",
	"this enchantment",
	"this vehicle",
}

func recognizeSelfCombatTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	for _, template := range []permanentActionTemplate{
		{suffix: " attacks", event: TriggerEventAttackerDeclared},
		{suffix: " blocks", event: TriggerEventBlockerDeclared},
		{suffix: " becomes blocked", event: TriggerEventAttackerBecameBlocked},
	} {
		subject, ok := strings.CutSuffix(event, template.suffix)
		if ok && subject == "this creature" {
			return TriggerPattern{Event: template.event, Source: TriggerSourceSelf}, true
		}
	}
	return TriggerPattern{}, false
}

func recognizeSelfPermanentStateTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	if event == "this creature mutates" {
		return TriggerPattern{Event: TriggerEventPermanentMutated, Source: TriggerSourceSelf}, true
	}
	for _, template := range []struct {
		suffix string
		event  TriggerEvent
	}{
		{suffix: " becomes tapped", event: TriggerEventPermanentTapped},
		{suffix: " becomes untapped", event: TriggerEventPermanentUntapped},
	} {
		subject, ok := strings.CutSuffix(event, template.suffix)
		if ok && matchesSelfSubjectSlot(subject, cardName, selfStateSubjectSlots, true) {
			return TriggerPattern{Event: template.event, Source: TriggerSourceSelf}, true
		}
	}
	return recognizeSelfCounterTrigger(event)
}

func recognizeSelfCounterTrigger(event string) (TriggerPattern, bool) {
	oneOrMore := false
	counterText, subject, ok := strings.Cut(event, " counter is put on ")
	if !ok {
		counterText, subject, ok = strings.Cut(event, " counters are put on ")
		if !ok {
			return TriggerPattern{}, false
		}
		counterText, oneOrMore = strings.CutPrefix(counterText, "one or more ")
		if !oneOrMore {
			return TriggerPattern{}, false
		}
	} else {
		counterText, ok = strings.CutPrefix(counterText, "a ")
		if !ok {
			return TriggerPattern{}, false
		}
	}
	if subject != "this creature" && subject != "this permanent" {
		return TriggerPattern{}, false
	}
	var counter TriggerCounter
	switch counterText {
	case "+1/+1":
		counter = TriggerCounterPlusOnePlusOne
	case "-1/-1":
		counter = TriggerCounterMinusOneMinusOne
	default:
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Event:     TriggerEventCountersAdded,
		Source:    TriggerSourceSelf,
		Counter:   counter,
		OneOrMore: oneOrMore,
	}, true
}

func matchesSelfSubjectSlot(subject, cardName string, slots []string, allowCardName bool) bool {
	if slices.Contains(slots, subject) {
		return true
	}
	return allowCardName && matchesCardNameSubject(subject, cardName)
}

func matchesCardNameSubject(subject, cardName string) bool {
	if cardName == "" {
		return false
	}
	if strings.EqualFold(subject, cardName) {
		return true
	}
	shortName, _, hasComma := strings.Cut(cardName, ",")
	if hasComma && strings.EqualFold(subject, shortName) {
		return true
	}
	frontName, _, hasBackFace := strings.Cut(cardName, " // ")
	if hasBackFace && strings.EqualFold(subject, frontName) {
		return true
	}
	firstWord, _, hasMore := strings.Cut(cardName, " ")
	return hasMore &&
		!slices.Contains([]string{"a", "an", "the"}, strings.ToLower(firstWord)) &&
		strings.EqualFold(subject, firstWord)
}

func recognizeBecameTargetTrigger(event, cardName string) (TriggerPattern, bool) {
	subjects := []string{
		"this creature",
		"this permanent",
		"this artifact",
		"this enchantment",
		"this land",
		"this planeswalker",
	}
	if cardName != "" {
		subjects = append(subjects, cardName)
	}
	for _, template := range []struct {
		suffix      string
		stackObject TriggerStackObject
	}{
		{suffix: " becomes the target of a spell", stackObject: TriggerStackObjectSpell},
		{suffix: " becomes the target of a spell or ability"},
	} {
		subject, ok := strings.CutSuffix(event, template.suffix)
		if !ok || !slices.ContainsFunc(subjects, func(candidate string) bool {
			return strings.EqualFold(subject, candidate)
		}) {
			continue
		}
		return TriggerPattern{
			Event:       TriggerEventObjectBecameTarget,
			Source:      TriggerSourceSelf,
			StackObject: template.stackObject,
		}, true
	}
	return TriggerPattern{}, false
}

func recognizeParameterizedDamageTrigger(event, sourceName string) (TriggerPattern, bool) {
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
		if event == template.text {
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
		subject, ok := strings.CutSuffix(event, template.suffix)
		if !ok {
			continue
		}
		pattern, ok := damageRecipientSubjectPattern(subject, sourceName, template.plural)
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
		source, remainder, ok := strings.Cut(event, template.marker)
		if !ok {
			continue
		}
		pattern, ok := damageSourcePattern(source, sourceName, template.plural)
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
		recipient := parseDamageRecipient(target, sourceName)
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

func damageSourcePattern(subject, sourceName string, plural bool) (TriggerPattern, bool) {
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
	if matchesSelfSubjectSlot(subject, sourceName, selfCombatSubjectSlots, true) {
		return TriggerPattern{
			Event:     TriggerEventDamageDealt,
			Source:    TriggerSourceSelf,
			Subject:   TriggerSubjectDamageSource,
			OneOrMore: oneOrMore,
		}, true
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
	parsed := parseCombatPermanentSelection(subject, plural)
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

func damageRecipientSubjectPattern(subject, sourceName string, plural bool) (TriggerPattern, bool) {
	oneOrMore := false
	if rest, ok := strings.CutPrefix(subject, "one or more "); ok {
		subject = rest
		plural = true
		oneOrMore = true
	}
	if matchesSelfSubjectSlot(subject, sourceName, selfCombatSubjectSlots, true) {
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
	parsed := parseCombatPermanentSelection(subject, plural)
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

func parseDamageRecipient(recipient, sourceName string) damageRecipientPattern {
	if matchesSelfSubjectSlot(recipient, sourceName, selfCombatSubjectSlots, true) {
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
	parsed := parseCombatPermanentSelection(recipient, false)
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

func recognizeCastTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
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
		if rest, ok := strings.CutPrefix(event, relation.text); ok {
			pattern.Controller = relation.controller
			phrase = rest
			break
		}
	}
	if phrase == "" {
		return TriggerPattern{}, false
	}
	if !compileCastSelection(phrase, &pattern) {
		return TriggerPattern{}, false
	}
	if pattern.MatchFromZone && pattern.Controller != ControllerYou {
		return TriggerPattern{}, false
	}
	return pattern, true
}

func compileCastSelection(phrase string, pattern *TriggerPattern) bool {
	type predicate struct {
		selection     TriggerSelection
		kicker        bool
		historic      bool
		fromGraveyard bool
	}
	predicates := map[string]predicate{
		"a spell":                     {},
		"a noncreature spell":         {selection: TriggerSelection{ExcludedTypes: []TriggerCardType{TriggerCardTypeCreature}}},
		"a creature spell":            {selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}}},
		"an instant or sorcery spell": {selection: TriggerSelection{RequiredTypesAny: []TriggerCardType{TriggerCardTypeInstant, TriggerCardTypeSorcery}}},
		"an instant spell":            {selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeInstant}}},
		"an instant":                  {selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeInstant}}},
		"a sorcery spell":             {selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeSorcery}}},
		"an artifact spell":           {selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact}}},
		"an enchantment spell":        {selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeEnchantment}}},
		"a land spell":                {selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeLand}}},
		"a planeswalker spell":        {selection: TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker}}},
		"a noncreature, nonland spell": {
			selection: TriggerSelection{ExcludedTypes: []TriggerCardType{TriggerCardTypeCreature, TriggerCardTypeLand}},
		},
		"a white spell":               {selection: TriggerSelection{ColorsAny: []TriggerColor{TriggerColorWhite}}},
		"a blue spell":                {selection: TriggerSelection{ColorsAny: []TriggerColor{TriggerColorBlue}}},
		"a black spell":               {selection: TriggerSelection{ColorsAny: []TriggerColor{TriggerColorBlack}}},
		"a red spell":                 {selection: TriggerSelection{ColorsAny: []TriggerColor{TriggerColorRed}}},
		"a green spell":               {selection: TriggerSelection{ColorsAny: []TriggerColor{TriggerColorGreen}}},
		"a colorless spell":           {selection: TriggerSelection{Colorless: true}},
		"a multicolored spell":        {selection: TriggerSelection{Multicolored: true}},
		"a kicked spell":              {kicker: true},
		"a spell from your graveyard": {fromGraveyard: true},
		"a spirit or arcane spell": {
			selection: TriggerSelection{SubtypesAny: []TriggerSubtype{TriggerSubtypeSpirit, TriggerSubtypeArcane}},
		},
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

func recognizeZoneChangeTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	if kind != TriggerWhen && kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	for _, template := range permanentZoneChangeTemplates {
		if subject, ok := strings.CutSuffix(event, template.singularSuffix); ok {
			pattern, ok := bindSinglePermanentZoneChangeSubject(subject, cardName)
			if !ok {
				continue
			}
			if !completePermanentZoneChangePattern(&pattern, template) {
				return TriggerPattern{}, false
			}
			return pattern, true
		}
		if subject, ok := strings.CutSuffix(event, template.pluralSuffix); ok {
			pattern, ok := bindPluralPermanentZoneChangeSubject(subject, cardName)
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

func bindSinglePermanentZoneChangeSubject(subject, cardName string) (TriggerPattern, bool) {
	if matchesSelfSubjectSlot(subject, cardName, selfEnterSubjectSlots, true) {
		return TriggerPattern{Source: TriggerSourceSelf}, true
	}
	if selection, ok := parseAttachedPermanentZoneChangeSubject(subject); ok {
		return TriggerPattern{Source: TriggerSourceAttachedPermanent, SubjectSelection: selection}, true
	}

	subject, otherThanSelf := stripOtherThanSelfSubject(subject, cardName)
	parsed, ok := parseZoneChangePermanentSubject(subject, false)
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

func stripOtherThanSelfSubject(subject, cardName string) (string, bool) {
	before, excluded, ok := strings.Cut(subject, " other than ")
	if !ok {
		return subject, false
	}
	if matchesSelfSubjectSlot(excluded, cardName, selfEnterSubjectSlots, true) {
		return before, true
	}
	return subject, false
}

func parseAttachedPermanentZoneChangeSubject(subject string) (TriggerSelection, bool) {
	if selection, ok := attachedPermanentSubjects[subject]; ok {
		return selection, true
	}
	for _, prefix := range []string{"enchanted ", "equipped ", "fortified "} {
		rest, ok := strings.CutPrefix(subject, prefix)
		if !ok {
			continue
		}
		selection, ok := parsePermanentTriggerSelection(rest, false)
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

func bindPluralPermanentZoneChangeSubject(subject, cardName string) (TriggerPattern, bool) {
	if matchesCardNameSubject(subject, cardName) {
		return TriggerPattern{Source: TriggerSourceSelf}, true
	}
	subject, oneOrMore := strings.CutPrefix(subject, "one or more ")
	if !oneOrMore {
		return TriggerPattern{}, false
	}
	parsed, ok := parseZoneChangePermanentSubject(subject, true)
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

func parseZoneChangePermanentSubject(subject string, plural bool) (zoneChangePermanentSubject, bool) {
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
	result.selection, ok = parsePermanentTriggerSelection(subject, plural)
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

func parsePermanentTriggerSelection(subject string, plural bool) (TriggerSelection, bool) {
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
		rest, matched := parsePermanentTriggerSelectionAdjective(subject, &selection)
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
		leftType, leftOK := parseSingleTriggerPermanentType(left, plural)
		rightType, rightOK := parseSingleTriggerPermanentType(right, plural)
		if leftOK && rightOK && leftType != TriggerCardTypeUnknown && rightType != TriggerCardTypeUnknown {
			selection.RequiredTypesAny = []TriggerCardType{leftType, rightType}
			return selection, true
		}
		if leftOK || rightOK {
			return TriggerSelection{}, false
		}
		left = singularTriggerSubtype(left, plural)
		right = singularTriggerSubtype(right, plural)
		if !looksLikeTriggerSubtype(left) || !looksLikeTriggerSubtype(right) {
			return TriggerSelection{}, false
		}
		selection.SubtypesAny = []TriggerSubtype{
			TriggerSubtype(left),
			TriggerSubtype(right),
		}
		return selection, true
	}
	if strings.HasPrefix(subject, "non") {
		excluded, rest, ok := parseExcludedTriggerPermanentType(subject, plural)
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
		cardType, ok := parseSingleTriggerPermanentType(word, plural)
		if !ok {
			subtypeWords = append(subtypeWords, word)
			continue
		}
		if cardType != TriggerCardTypeUnknown {
			addRequiredTriggerType(&selection, cardType)
		}
	}
	if len(subtypeWords) > 0 {
		subtype := singularTriggerSubtype(strings.Join(subtypeWords, " "), plural)
		if subtype == "outlaw" {
			selection.SubtypesAny = []TriggerSubtype{"assassin", "mercenary", "pirate", "rogue", "warlock"}
			return selection, true
		}
		if !looksLikeTriggerSubtype(subtype) {
			return TriggerSelection{}, false
		}
		selection.SubtypesAny = []TriggerSubtype{TriggerSubtype(subtype)}
	}
	return selection, true
}

func looksLikeTriggerSubtype(subject string) bool {
	fields := strings.Fields(subject)
	if len(fields) == 0 || len(fields) > 2 {
		return false
	}
	for _, word := range fields {
		if _, ok := triggerPermanentType(word); ok ||
			strings.HasPrefix(word, "non") ||
			slices.Contains([]string{
				"base", "chosen", "commander", "double-faced", "face-down", "historic", "modified",
				"an", "a", "the", "you", "your", "opponent", "or", "and", "but",
			}, word) {
			return false
		}
	}
	for _, r := range subject {
		if (r >= 'a' && r <= 'z') ||
			r == ' ' ||
			r == '-' ||
			r == '\'' {
			continue
		}
		return false
	}
	return true
}

func parsePermanentTriggerSelectionAdjective(subject string, selection *TriggerSelection) (string, bool) {
	adjectives := []struct {
		prefix string
		bind   func()
	}{
		{prefix: "nontoken ", bind: func() { selection.NonToken = true }},
		{prefix: "token ", bind: func() { selection.TokenOnly = true }},
		{prefix: "legendary ", bind: func() {
			selection.Supertypes = append(selection.Supertypes, TriggerSupertypeLegendary)
		}},
		{prefix: "snow ", bind: func() {
			selection.Supertypes = append(selection.Supertypes, TriggerSupertypeSnow)
		}},
		{prefix: "white ", bind: func() { selection.ColorsAny = append(selection.ColorsAny, TriggerColorWhite) }},
		{prefix: "blue ", bind: func() { selection.ColorsAny = append(selection.ColorsAny, TriggerColorBlue) }},
		{prefix: "black ", bind: func() { selection.ColorsAny = append(selection.ColorsAny, TriggerColorBlack) }},
		{prefix: "red ", bind: func() { selection.ColorsAny = append(selection.ColorsAny, TriggerColorRed) }},
		{prefix: "green ", bind: func() { selection.ColorsAny = append(selection.ColorsAny, TriggerColorGreen) }},
		{prefix: "nonwhite ", bind: func() {
			selection.ExcludedColors = append(selection.ExcludedColors, TriggerColorWhite)
		}},
		{prefix: "nonblue ", bind: func() {
			selection.ExcludedColors = append(selection.ExcludedColors, TriggerColorBlue)
		}},
		{prefix: "nonblack ", bind: func() {
			selection.ExcludedColors = append(selection.ExcludedColors, TriggerColorBlack)
		}},
		{prefix: "nonred ", bind: func() {
			selection.ExcludedColors = append(selection.ExcludedColors, TriggerColorRed)
		}},
		{prefix: "nongreen ", bind: func() {
			selection.ExcludedColors = append(selection.ExcludedColors, TriggerColorGreen)
		}},
		{prefix: "colorless ", bind: func() { selection.Colorless = true }},
		{prefix: "multicolored ", bind: func() { selection.Multicolored = true }},
		{prefix: "attacking ", bind: func() { selection.CombatState = TriggerCombatStateAttacking }},
		{prefix: "blocking ", bind: func() { selection.CombatState = TriggerCombatStateBlocking }},
	}
	for _, adjective := range adjectives {
		if rest, ok := strings.CutPrefix(subject, adjective.prefix); ok {
			adjective.bind()
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

func singularTriggerSubtype(subject string, plural bool) string {
	if !plural {
		return subject
	}
	switch subject {
	case "children":
		return "child"
	case "dwarves":
		return "dwarf"
	case "elves":
		return "elf"
	case "faeries":
		return "faerie"
	case "mice":
		return "mouse"
	case "oxen":
		return "ox"
	case "wolves":
		return "wolf"
	}
	switch {
	case strings.HasSuffix(subject, "ies"):
		return strings.TrimSuffix(subject, "ies") + "y"
	case strings.HasSuffix(subject, "ses"),
		strings.HasSuffix(subject, "xes"),
		strings.HasSuffix(subject, "zes"),
		strings.HasSuffix(subject, "ches"),
		strings.HasSuffix(subject, "shes"):
		return strings.TrimSuffix(subject, "es")
	case strings.HasSuffix(subject, "s"):
		return strings.TrimSuffix(subject, "s")
	default:
		return subject
	}
}

func parseExcludedTriggerPermanentType(subject string, plural bool) (TriggerCardType, string, bool) {
	for _, word := range []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker"} {
		prefix := "non" + word + " "
		if strings.HasPrefix(subject, prefix) {
			cardType, _ := parseSingleTriggerPermanentType(word, plural)
			return cardType, strings.TrimPrefix(subject, prefix), true
		}
	}
	return TriggerCardTypeUnknown, "", false
}

func parseSingleTriggerPermanentType(word string, plural bool) (TriggerCardType, bool) {
	word = strings.TrimPrefix(strings.TrimPrefix(word, "a "), "an ")
	if plural {
		if cardType, ok := triggerPermanentPluralType(word); ok {
			return cardType, true
		}
	}
	return triggerPermanentType(word)
}

type permanentActionTemplate struct {
	suffix string
	event  TriggerEvent
}

var combatPermanentActions = []permanentActionTemplate{
	{suffix: " attacks", event: TriggerEventAttackerDeclared},
	{suffix: " blocks", event: TriggerEventBlockerDeclared},
}

var statePermanentActions = []permanentActionTemplate{
	{suffix: " becomes tapped", event: TriggerEventPermanentTapped},
	{suffix: " becomes untapped", event: TriggerEventPermanentUntapped},
}

func recognizePermanentActionTrigger(event string, kind TriggerKind, actions []permanentActionTemplate) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	for _, action := range actions {
		if !strings.HasSuffix(event, action.suffix) {
			continue
		}
		subject := strings.TrimSuffix(event, action.suffix)
		if subject == "enchanted creature" || subject == "equipped creature" ||
			(subject == "enchanted permanent" &&
				(action.event == TriggerEventPermanentTapped || action.event == TriggerEventPermanentUntapped)) {
			selection := TriggerSelection{}
			if subject != "enchanted permanent" {
				selection.RequiredTypes = []TriggerCardType{TriggerCardTypeCreature}
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

func parseSinglePermanentEventSubject(event, suffix string) (permanentEventSubject, bool) {
	if !strings.HasSuffix(event, suffix) {
		return permanentEventSubject{}, false
	}
	subject := strings.TrimSuffix(event, suffix)
	controller := ControllerAny
	switch {
	case strings.HasSuffix(subject, " you control"):
		subject = strings.TrimSuffix(subject, " you control")
		controller = ControllerYou
	case strings.HasSuffix(subject, " an opponent controls"):
		subject = strings.TrimSuffix(subject, " an opponent controls")
		controller = ControllerOpponent
	default:
	}
	excludeSelf := false
	switch {
	case strings.HasPrefix(subject, "another "):
		excludeSelf = true
		subject = strings.TrimPrefix(subject, "another ")
	case strings.HasPrefix(subject, "a "):
		subject = strings.TrimPrefix(subject, "a ")
	case strings.HasPrefix(subject, "an "):
		subject = strings.TrimPrefix(subject, "an ")
	default:
		return permanentEventSubject{}, false
	}
	selection := TriggerSelection{}
	if strings.HasPrefix(subject, "nontoken ") {
		selection.NonToken = true
		subject = strings.TrimPrefix(subject, "nontoken ")
	}
	cardType, ok := triggerPermanentType(subject)
	if !ok {
		return permanentEventSubject{}, false
	}
	if cardType != TriggerCardTypeUnknown {
		selection.RequiredTypes = []TriggerCardType{cardType}
	}
	return permanentEventSubject{
		selection:   selection,
		controller:  controller,
		excludeSelf: excludeSelf,
	}, true
}

func parsePluralPermanentEventSubject(
	event string,
	suffix string,
) (TriggerSelection, ControllerKind, bool) {
	if !strings.HasSuffix(event, suffix) {
		return TriggerSelection{}, ControllerAny, false
	}
	subject := strings.TrimSuffix(event, suffix)
	controller := ControllerAny
	switch {
	case strings.HasSuffix(subject, " you control"):
		subject = strings.TrimSuffix(subject, " you control")
		controller = ControllerYou
	case strings.HasSuffix(subject, " an opponent controls"):
		subject = strings.TrimSuffix(subject, " an opponent controls")
		controller = ControllerOpponent
	default:
	}
	cardType, ok := triggerPermanentPluralType(subject)
	if !ok {
		return TriggerSelection{}, ControllerAny, false
	}
	selection := TriggerSelection{}
	if cardType != TriggerCardTypeUnknown {
		selection.RequiredTypes = []TriggerCardType{cardType}
	}
	return selection, controller, true
}

func triggerPermanentType(word string) (TriggerCardType, bool) {
	switch word {
	case "artifact":
		return TriggerCardTypeArtifact, true
	case "battle":
		return TriggerCardTypeBattle, true
	case "creature":
		return TriggerCardTypeCreature, true
	case "enchantment":
		return TriggerCardTypeEnchantment, true
	case "land":
		return TriggerCardTypeLand, true
	case "permanent":
		return TriggerCardTypeUnknown, true
	case "planeswalker":
		return TriggerCardTypePlaneswalker, true
	default:
		return TriggerCardTypeUnknown, false
	}
}

func triggerPermanentPluralType(word string) (TriggerCardType, bool) {
	switch word {
	case "artifacts":
		return TriggerCardTypeArtifact, true
	case "battles":
		return TriggerCardTypeBattle, true
	case "creatures":
		return TriggerCardTypeCreature, true
	case "enchantments":
		return TriggerCardTypeEnchantment, true
	case "lands":
		return TriggerCardTypeLand, true
	case "permanents":
		return TriggerCardTypeUnknown, true
	case "planeswalkers":
		return TriggerCardTypePlaneswalker, true
	default:
		return TriggerCardTypeUnknown, false
	}
}
