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
)

// TriggerDamageRecipient identifies what received damage.
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
type TriggerSubtype uint8

// Trigger subtypes.
const (
	TriggerSubtypeUnknown TriggerSubtype = iota
	TriggerSubtypeSpirit
	TriggerSubtypeArcane
)

// TriggerSelection is the closed semantic Selection vocabulary currently used
// by representable event subjects and cast spells. Its zero value is a
// wildcard.
type TriggerSelection struct {
	RequiredTypes    []TriggerCardType
	RequiredTypesAny []TriggerCardType
	ExcludedTypes    []TriggerCardType
	SubtypesAny      []TriggerSubtype
	ColorsAny        []TriggerColor
	Colorless        bool
	Multicolored     bool
	NonToken         bool
	ManaValueAtLeast int
	MatchManaValue   bool
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
	CardSelection            TriggerSelection
	DamageRecipientSelection TriggerSelection

	MatchFromZone bool
	FromZone      TriggerZone
	MatchToZone   bool
	ToZone        TriggerZone

	Step            TriggerStep
	CombatQualifier TriggerCombatQualifier
	DamageRecipient TriggerDamageRecipient
	StackObject     TriggerStackObject
	Counter         TriggerCounter

	ExcludeSelf       bool
	OneOrMore         bool
	RequireKickerPaid bool
	RequireHistoric   bool

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
	{kinds: []TriggerKind{TriggerWhenever}, bind: recognizeCombatTrigger},
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
	"your second main phase":              {Step: TriggerStepPostcombatMain, Controller: ControllerYou},
	"each of your postcombat main phases": {Step: TriggerStepPostcombatMain, Controller: ControllerYou},
}

func recognizePhaseStepTrigger(event string, _ TriggerKind, _ string) (TriggerPattern, bool) {
	event, ok := strings.CutPrefix(event, "the beginning of ")
	if !ok {
		return TriggerPattern{}, false
	}
	if pattern, ok := phaseStepAliases[event]; ok {
		pattern.Event = TriggerEventBeginningOfStep
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

func recognizePermanentZoneChangeTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	if pattern, ok := recognizeSelfZoneChangeTrigger(event, kind, cardName); ok {
		return pattern, true
	}
	if pattern, ok := recognizeEnterTrigger(event, kind); ok {
		return pattern, true
	}
	return recognizeDiesTrigger(event, kind)
}

func recognizeSpellAbilityTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	if pattern, ok := recognizeCastTrigger(event, kind); ok {
		return pattern, true
	}
	return recognizeBecameTargetTrigger(event, cardName)
}

func recognizeCombatTrigger(event string, kind TriggerKind, _ string) (TriggerPattern, bool) {
	if pattern, ok := recognizeSelfCombatTrigger(event, kind); ok {
		return pattern, true
	}
	if pattern, ok := recognizeDamageTrigger(event, kind); ok {
		return pattern, true
	}
	return recognizePermanentActionTrigger(event, kind, combatPermanentActions)
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
	"this aura",
	"this artifact",
	"this equipment",
	"this land",
	"this vehicle",
	"this enchantment",
	"this battle",
	"this case",
	"this class",
	"this planeswalker",
	"this spacecraft",
}

var selfGraveyardSubjectSlots = []string{
	"this aura",
	"this artifact",
	"this enchantment",
	"this vehicle",
	"this equipment",
}

var selfStateSubjectSlots = []string{
	"this creature",
	"this permanent",
	"this land",
	"this artifact",
	"this enchantment",
	"this vehicle",
}

func recognizeSelfZoneChangeTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	if kind != TriggerWhen {
		return TriggerPattern{}, false
	}
	templates := []struct {
		suffix        string
		event         TriggerEvent
		subjects      []string
		allowCardName bool
	}{
		{suffix: " enters", event: TriggerEventPermanentEnteredBattlefield, subjects: selfEnterSubjectSlots, allowCardName: true},
		{suffix: " dies", event: TriggerEventPermanentDied, subjects: []string{"this creature", "this permanent"}, allowCardName: true},
		{suffix: " is put into a graveyard from the battlefield", event: TriggerEventPermanentDied, subjects: selfGraveyardSubjectSlots},
	}
	for _, template := range templates {
		subject, ok := strings.CutSuffix(event, template.suffix)
		if !ok || !matchesSelfSubjectSlot(subject, cardName, template.subjects, template.allowCardName) {
			continue
		}
		return TriggerPattern{Event: template.event, Source: TriggerSourceSelf}, true
	}
	return TriggerPattern{}, false
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
	return allowCardName && cardName != "" && strings.EqualFold(subject, cardName)
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

func recognizeEnterTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	pattern := TriggerPattern{Event: TriggerEventPermanentEnteredBattlefield}
	if rest, ok := strings.CutPrefix(event, "one or more "); ok {
		pattern.OneOrMore = true
		selection, controller, ok := parsePluralPermanentEventSubject(rest, " enter")
		if !ok {
			return TriggerPattern{}, false
		}
		pattern.SubjectSelection = selection
		pattern.Controller = controller
		return pattern, true
	}
	subject, ok := parseSinglePermanentEventSubject(event, " enters")
	if !ok {
		return TriggerPattern{}, false
	}
	pattern.SubjectSelection = subject.selection
	pattern.Controller = subject.controller
	pattern.ExcludeSelf = subject.excludeSelf
	return pattern, true
}

func recognizeDiesTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
	attached := map[string]TriggerSelection{
		"enchanted creature":  {RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}},
		"equipped creature":   {RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}},
		"enchanted land":      {RequiredTypes: []TriggerCardType{TriggerCardTypeLand}},
		"enchanted permanent": {},
	}
	if kind == TriggerWhen || kind == TriggerWhenever {
		subject, matchesDiesTemplate := strings.CutSuffix(event, " dies")
		if selection, ok := attached[subject]; matchesDiesTemplate && ok {
			return TriggerPattern{
				Event:            TriggerEventPermanentDied,
				Source:           TriggerSourceAttachedPermanent,
				SubjectSelection: selection,
			}, true
		}
	}
	if kind == TriggerWhen {
		return TriggerPattern{}, false
	}
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	subject, ok := parseSinglePermanentEventSubject(event, " dies")
	if !ok || !slices.Contains(subject.selection.RequiredTypes, TriggerCardTypeCreature) {
		return TriggerPattern{}, false
	}
	return TriggerPattern{
		Event:            TriggerEventPermanentDied,
		Controller:       subject.controller,
		ExcludeSelf:      subject.excludeSelf,
		SubjectSelection: subject.selection,
	}, true
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
