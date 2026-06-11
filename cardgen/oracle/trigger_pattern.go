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

type stepTriggerPattern struct {
	step       TriggerStep
	controller ControllerKind
}

var stepTriggerPhrases = map[string]stepTriggerPattern{
	"the beginning of your upkeep":                           {TriggerStepUpkeep, ControllerYou},
	"the beginning of each upkeep":                           {TriggerStepUpkeep, ControllerAny},
	"the beginning of each player's upkeep":                  {TriggerStepUpkeep, ControllerAny},
	"the beginning of each opponent's upkeep":                {TriggerStepUpkeep, ControllerOpponent},
	"the beginning of your draw step":                        {TriggerStepDraw, ControllerYou},
	"the beginning of each draw step":                        {TriggerStepDraw, ControllerAny},
	"the beginning of each player's draw step":               {TriggerStepDraw, ControllerAny},
	"the beginning of each opponent's draw step":             {TriggerStepDraw, ControllerOpponent},
	"the beginning of your end step":                         {TriggerStepEnd, ControllerYou},
	"the beginning of each end step":                         {TriggerStepEnd, ControllerAny},
	"the beginning of each player's end step":                {TriggerStepEnd, ControllerAny},
	"the beginning of each opponent's end step":              {TriggerStepEnd, ControllerOpponent},
	"the beginning of combat on your turn":                   {TriggerStepBeginningOfCombat, ControllerYou},
	"the beginning of combat on each turn":                   {TriggerStepBeginningOfCombat, ControllerAny},
	"the beginning of combat on each opponent's turn":        {TriggerStepBeginningOfCombat, ControllerOpponent},
	"the beginning of each combat":                           {TriggerStepBeginningOfCombat, ControllerAny},
	"the beginning of the end of combat":                     {TriggerStepEndOfCombat, ControllerAny},
	"the beginning of the end of combat on your turn":        {TriggerStepEndOfCombat, ControllerYou},
	"the beginning of each end of combat step":               {TriggerStepEndOfCombat, ControllerAny},
	"the beginning of your first main phase":                 {TriggerStepPrecombatMain, ControllerYou},
	"the beginning of your precombat main phase":             {TriggerStepPrecombatMain, ControllerYou},
	"the beginning of each of your first main phases":        {TriggerStepPrecombatMain, ControllerYou},
	"the beginning of each player's precombat main phase":    {TriggerStepPrecombatMain, ControllerAny},
	"the beginning of your second main phase":                {TriggerStepPostcombatMain, ControllerYou},
	"the beginning of your postcombat main phase":            {TriggerStepPostcombatMain, ControllerYou},
	"the beginning of each of your postcombat main phases":   {TriggerStepPostcombatMain, ControllerYou},
	"the beginning of each player's postcombat main phase":   {TriggerStepPostcombatMain, ControllerAny},
	"the beginning of each opponent's postcombat main phase": {TriggerStepPostcombatMain, ControllerOpponent},
	"the beginning of each opponent's precombat main phase":  {TriggerStepPrecombatMain, ControllerOpponent},
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
	if kind == TriggerAt {
		if step, ok := stepTriggerPhrases[event]; ok {
			pattern.Event = TriggerEventBeginningOfStep
			pattern.Step = step.step
			pattern.Controller = step.controller
		}
		return pattern
	}
	if kind != TriggerWhen && kind != TriggerWhenever {
		return pattern
	}
	if recognized, ok := recognizeSimpleTrigger(event, kind); ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	if recognized, ok := recognizeSelfTrigger(event, kind, cardName); ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	if recognized, ok := recognizeDamageTrigger(event, kind); ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	if recognized, ok := recognizeCastTrigger(event, kind); ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	if recognized, ok := recognizeEnterTrigger(event, kind); ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	if recognized, ok := recognizeDiesTrigger(event, kind); ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	if recognized, ok := recognizePermanentActionTrigger(event, kind); ok {
		return completeTriggerPattern(&recognized, &pattern)
	}
	return pattern
}

func completeTriggerPattern(recognized, source *TriggerPattern) TriggerPattern {
	recognized.Span = source.Span
	recognized.Kind = source.Kind
	recognized.InterveningCondition = source.InterveningCondition
	return *recognized
}

func recognizeSimpleTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	patterns := map[string]TriggerPattern{
		"you draw a card": {
			Event:  TriggerEventCardDrawn,
			Player: TriggerPlayerYou,
		},
		"an opponent draws a card": {
			Event:  TriggerEventCardDrawn,
			Player: TriggerPlayerOpponent,
		},
		"a player draws a card": {
			Event: TriggerEventCardDrawn,
		},
		"you discard a card": {
			Event:  TriggerEventCardDiscarded,
			Player: TriggerPlayerYou,
		},
		"you discard one or more cards": {
			Event:     TriggerEventCardDiscarded,
			Player:    TriggerPlayerYou,
			OneOrMore: true,
		},
		"an opponent discards a card": {
			Event:  TriggerEventCardDiscarded,
			Player: TriggerPlayerOpponent,
		},
		"a player discards a card": {
			Event: TriggerEventCardDiscarded,
		},
		"you cycle a card": {
			Event:  TriggerEventCycled,
			Player: TriggerPlayerYou,
		},
		"you cycle another card": {
			Event:       TriggerEventCycled,
			Player:      TriggerPlayerYou,
			ExcludeSelf: true,
		},
		"you cycle or discard a card": {
			Event:  TriggerEventCardDiscarded,
			Player: TriggerPlayerYou,
		},
		"you cycle or discard another card": {
			Event:       TriggerEventCardDiscarded,
			Player:      TriggerPlayerYou,
			ExcludeSelf: true,
		},
		"you gain life": {
			Event:  TriggerEventLifeGained,
			Player: TriggerPlayerYou,
		},
		"an opponent gains life": {
			Event:  TriggerEventLifeGained,
			Player: TriggerPlayerOpponent,
		},
		"you lose life": {
			Event:  TriggerEventLifeLost,
			Player: TriggerPlayerYou,
		},
		"an opponent loses life": {
			Event:  TriggerEventLifeLost,
			Player: TriggerPlayerOpponent,
		},
	}
	pattern, ok := patterns[event]
	return pattern, ok
}

func recognizeSelfTrigger(event string, kind TriggerKind, cardName string) (TriggerPattern, bool) {
	whenPatterns := map[string]TriggerEvent{
		"this creature enters":     TriggerEventPermanentEnteredBattlefield,
		"this permanent enters":    TriggerEventPermanentEnteredBattlefield,
		"this aura enters":         TriggerEventPermanentEnteredBattlefield,
		"this artifact enters":     TriggerEventPermanentEnteredBattlefield,
		"this equipment enters":    TriggerEventPermanentEnteredBattlefield,
		"this land enters":         TriggerEventPermanentEnteredBattlefield,
		"this vehicle enters":      TriggerEventPermanentEnteredBattlefield,
		"this enchantment enters":  TriggerEventPermanentEnteredBattlefield,
		"this battle enters":       TriggerEventPermanentEnteredBattlefield,
		"this case enters":         TriggerEventPermanentEnteredBattlefield,
		"this class enters":        TriggerEventPermanentEnteredBattlefield,
		"this planeswalker enters": TriggerEventPermanentEnteredBattlefield,
		"this spacecraft enters":   TriggerEventPermanentEnteredBattlefield,
		"this creature dies":       TriggerEventPermanentDied,
		"this permanent dies":      TriggerEventPermanentDied,
		"this aura is put into a graveyard from the battlefield":        TriggerEventPermanentDied,
		"this artifact is put into a graveyard from the battlefield":    TriggerEventPermanentDied,
		"this enchantment is put into a graveyard from the battlefield": TriggerEventPermanentDied,
		"this vehicle is put into a graveyard from the battlefield":     TriggerEventPermanentDied,
		"this equipment is put into a graveyard from the battlefield":   TriggerEventPermanentDied,
	}
	if kind == TriggerWhen {
		if family, ok := whenPatterns[event]; ok {
			return TriggerPattern{Event: family, Source: TriggerSourceSelf}, true
		}
		if cardName != "" {
			switch {
			case strings.EqualFold(event, cardName+" enters"):
				return TriggerPattern{Event: TriggerEventPermanentEnteredBattlefield, Source: TriggerSourceSelf}, true
			case strings.EqualFold(event, cardName+" dies"):
				return TriggerPattern{Event: TriggerEventPermanentDied, Source: TriggerSourceSelf}, true
			}
		}
	}
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	wheneverPatterns := map[string]TriggerPattern{
		"this creature mutates": {
			Event:  TriggerEventPermanentMutated,
			Source: TriggerSourceSelf,
		},
		"this creature attacks": {
			Event:  TriggerEventAttackerDeclared,
			Source: TriggerSourceSelf,
		},
		"this creature blocks": {
			Event:  TriggerEventBlockerDeclared,
			Source: TriggerSourceSelf,
		},
		"this creature becomes blocked": {
			Event:  TriggerEventAttackerBecameBlocked,
			Source: TriggerSourceSelf,
		},
		"this creature becomes tapped": {
			Event:  TriggerEventPermanentTapped,
			Source: TriggerSourceSelf,
		},
		"this permanent becomes tapped": {
			Event:  TriggerEventPermanentTapped,
			Source: TriggerSourceSelf,
		},
		"this land becomes tapped": {
			Event:  TriggerEventPermanentTapped,
			Source: TriggerSourceSelf,
		},
		"this artifact becomes tapped": {
			Event:  TriggerEventPermanentTapped,
			Source: TriggerSourceSelf,
		},
		"this enchantment becomes tapped": {
			Event:  TriggerEventPermanentTapped,
			Source: TriggerSourceSelf,
		},
		"this vehicle becomes tapped": {
			Event:  TriggerEventPermanentTapped,
			Source: TriggerSourceSelf,
		},
		"this creature becomes untapped": {
			Event:  TriggerEventPermanentUntapped,
			Source: TriggerSourceSelf,
		},
		"this permanent becomes untapped": {
			Event:  TriggerEventPermanentUntapped,
			Source: TriggerSourceSelf,
		},
		"this land becomes untapped": {
			Event:  TriggerEventPermanentUntapped,
			Source: TriggerSourceSelf,
		},
		"this artifact becomes untapped": {
			Event:  TriggerEventPermanentUntapped,
			Source: TriggerSourceSelf,
		},
		"this enchantment becomes untapped": {
			Event:  TriggerEventPermanentUntapped,
			Source: TriggerSourceSelf,
		},
		"this vehicle becomes untapped": {
			Event:  TriggerEventPermanentUntapped,
			Source: TriggerSourceSelf,
		},
	}
	if pattern, ok := wheneverPatterns[event]; ok {
		return pattern, true
	}
	if cardName != "" {
		switch {
		case strings.EqualFold(event, cardName+" becomes tapped"):
			return TriggerPattern{Event: TriggerEventPermanentTapped, Source: TriggerSourceSelf}, true
		case strings.EqualFold(event, cardName+" becomes untapped"):
			return TriggerPattern{Event: TriggerEventPermanentUntapped, Source: TriggerSourceSelf}, true
		}
	}
	counterPatterns := map[string]TriggerPattern{
		"one or more +1/+1 counters are put on this creature": {
			Event: TriggerEventCountersAdded, Source: TriggerSourceSelf, Counter: TriggerCounterPlusOnePlusOne, OneOrMore: true,
		},
		"one or more +1/+1 counters are put on this permanent": {
			Event: TriggerEventCountersAdded, Source: TriggerSourceSelf, Counter: TriggerCounterPlusOnePlusOne, OneOrMore: true,
		},
		"a +1/+1 counter is put on this creature": {
			Event: TriggerEventCountersAdded, Source: TriggerSourceSelf, Counter: TriggerCounterPlusOnePlusOne,
		},
		"a +1/+1 counter is put on this permanent": {
			Event: TriggerEventCountersAdded, Source: TriggerSourceSelf, Counter: TriggerCounterPlusOnePlusOne,
		},
		"one or more -1/-1 counters are put on this creature": {
			Event: TriggerEventCountersAdded, Source: TriggerSourceSelf, Counter: TriggerCounterMinusOneMinusOne, OneOrMore: true,
		},
		"one or more -1/-1 counters are put on this permanent": {
			Event: TriggerEventCountersAdded, Source: TriggerSourceSelf, Counter: TriggerCounterMinusOneMinusOne, OneOrMore: true,
		},
		"a -1/-1 counter is put on this creature": {
			Event: TriggerEventCountersAdded, Source: TriggerSourceSelf, Counter: TriggerCounterMinusOneMinusOne,
		},
		"a -1/-1 counter is put on this permanent": {
			Event: TriggerEventCountersAdded, Source: TriggerSourceSelf, Counter: TriggerCounterMinusOneMinusOne,
		},
	}
	if pattern, ok := counterPatterns[event]; ok {
		return pattern, true
	}
	return recognizeBecameTargetTrigger(event, cardName)
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
	for _, subject := range subjects {
		if !strings.EqualFold(event, subject+" becomes the target of a spell") &&
			!strings.EqualFold(event, subject+" becomes the target of a spell or ability") {
			continue
		}
		pattern := TriggerPattern{Event: TriggerEventObjectBecameTarget, Source: TriggerSourceSelf}
		if strings.HasSuffix(strings.ToLower(event), " of a spell") {
			pattern.StackObject = TriggerStackObjectSpell
		}
		return pattern, true
	}
	return TriggerPattern{}, false
}

func recognizeDamageTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	patterns := map[string]TriggerPattern{
		"this creature is dealt damage": {
			Event: TriggerEventDamageDealt, Source: TriggerSourceSelf, Subject: TriggerSubjectPermanent, DamageRecipient: TriggerDamageRecipientPermanent,
		},
		"this permanent is dealt damage": {
			Event: TriggerEventDamageDealt, Source: TriggerSourceSelf, Subject: TriggerSubjectPermanent, DamageRecipient: TriggerDamageRecipientPermanent,
		},
		"enchanted creature is dealt damage": {
			Event: TriggerEventDamageDealt, Source: TriggerSourceAttachedPermanent, DamageRecipient: TriggerDamageRecipientPermanent,
		},
		"enchanted permanent is dealt damage": {
			Event: TriggerEventDamageDealt, Source: TriggerSourceAttachedPermanent, DamageRecipient: TriggerDamageRecipientPermanent,
		},
		"equipped creature is dealt damage": {
			Event: TriggerEventDamageDealt, Source: TriggerSourceAttachedPermanent, DamageRecipient: TriggerDamageRecipientPermanent,
		},
		"you're dealt damage": {
			Event: TriggerEventDamageDealt, Player: TriggerPlayerYou, DamageRecipient: TriggerDamageRecipientPlayer,
		},
		"you are dealt damage": {
			Event: TriggerEventDamageDealt, Player: TriggerPlayerYou, DamageRecipient: TriggerDamageRecipientPlayer,
		},
	}
	if pattern, ok := patterns[event]; ok {
		return pattern, true
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
	switch {
	case strings.HasPrefix(event, "you cast "):
		pattern.Controller = ControllerYou
		phrase = strings.TrimPrefix(event, "you cast ")
	case strings.HasPrefix(event, "a player casts "):
		phrase = strings.TrimPrefix(event, "a player casts ")
	case strings.HasPrefix(event, "an opponent casts "):
		pattern.Controller = ControllerOpponent
		phrase = strings.TrimPrefix(event, "an opponent casts ")
	default:
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
		"enchanted creature dies":  {RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}},
		"equipped creature dies":   {RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}},
		"enchanted land dies":      {RequiredTypes: []TriggerCardType{TriggerCardTypeLand}},
		"enchanted permanent dies": {},
	}
	if kind == TriggerWhen || kind == TriggerWhenever {
		if selection, ok := attached[event]; ok {
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

func recognizePermanentActionTrigger(event string, kind TriggerKind) (TriggerPattern, bool) {
	if kind != TriggerWhenever {
		return TriggerPattern{}, false
	}
	actions := []struct {
		suffix string
		event  TriggerEvent
	}{
		{" attacks", TriggerEventAttackerDeclared},
		{" blocks", TriggerEventBlockerDeclared},
		{" becomes tapped", TriggerEventPermanentTapped},
		{" becomes untapped", TriggerEventPermanentUntapped},
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
