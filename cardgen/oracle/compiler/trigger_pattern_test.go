package compiler

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
)

func compileParsedTriggerPattern(
	event string,
	kind TriggerKind,
	span shared.Span,
	cardName string,
	condition *CompiledCondition,
) TriggerPattern {
	introduction := "When"
	switch kind {
	case TriggerAt:
		introduction = "At"
	case TriggerWhenever:
		introduction = "Whenever"
	default:
	}
	document, _ := parser.Parse(introduction+" "+event+", draw a card.", parser.Context{CardName: cardName})
	if len(document.Abilities) != 1 || document.Abilities[0].Trigger == nil ||
		document.Abilities[0].Trigger.TriggerEvent == nil {
		return TriggerPattern{Span: span, Kind: kind, InterveningCondition: condition}
	}
	clause := *document.Abilities[0].Trigger.TriggerEvent
	clause.Span = span
	return compileTriggerEventPattern(&clause, kind, condition)
}

func TestTypedTriggerEventsBindClosedSlots(t *testing.T) {
	t.Parallel()
	condition := &CompiledCondition{
		Kind:      ConditionIf,
		Predicate: ConditionPredicateControllerControls,
	}
	tests := []struct {
		name      string
		event     string
		kind      TriggerKind
		cardName  string
		condition *CompiledCondition
		want      TriggerPattern
	}{
		{
			name:  "permanent zone change binds Selection relation and batching",
			event: "one or more artifacts you control enter",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:       TriggerWhenever,
				Event:      TriggerEventPermanentEnteredBattlefield,
				Controller: ControllerYou,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Artifact},
				},
				OneOrMore: true,
			},
		},
		{
			name:      "named permanent zone change binds self and condition",
			event:     "Example Card dies",
			kind:      TriggerWhen,
			cardName:  "Example Card",
			condition: condition,
			want: TriggerPattern{
				Kind:                 TriggerWhen,
				Event:                TriggerEventPermanentDied,
				Source:               TriggerSourceSelf,
				SubjectSelection:     TriggerSelection{RequiredTypes: []types.Card{types.Creature}},
				InterveningCondition: condition,
			},
		},
		{
			name:  "spell event binds controller Selection and zone",
			event: "you cast a spell from your graveyard",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:          TriggerWhenever,
				Event:         TriggerEventSpellCast,
				Controller:    ControllerYou,
				MatchFromZone: true,
				FromZone:      TriggerZoneGraveyard,
			},
		},
		{
			name:  "spell event binds per-turn ordinal",
			event: "you cast your second spell each turn",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:                       TriggerWhenever,
				Event:                      TriggerEventSpellCast,
				Controller:                 ControllerYou,
				PlayerEventOrdinalThisTurn: 2,
			},
		},
		{
			name:  "spell event binds chosen-type card Selection",
			event: "you cast a creature spell of the chosen type",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:       TriggerWhenever,
				Event:      TriggerEventSpellCast,
				Controller: ControllerYou,
				CardSelection: TriggerSelection{
					RequiredTypes:          []types.Card{types.Creature},
					SubtypeFromEntryChoice: true,
				},
			},
		},
		{
			name:  "spell or ability target event shares self template",
			event: "this creature becomes the target of a spell or ability",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventObjectBecameTarget,
				Source: TriggerSourceSelf,
			},
		},
		{
			name:  "activated ability binds actor source Selection and mana exclusion",
			event: "an opponent activates an ability of a creature or land that isn't a mana ability",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:               TriggerWhenever,
				Event:              TriggerEventAbilityActivated,
				Player:             TriggerPlayerOpponent,
				ExcludeManaAbility: true,
				SubjectSelection: TriggerSelection{
					RequiredTypesAny: []types.Card{types.Creature, types.Land},
				},
			},
		},
		{
			name:  "target event separates subject and cause controllers",
			event: "a creature you control becomes the target of a spell or ability an opponent controls",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:            TriggerWhenever,
				Event:           TriggerEventObjectBecameTarget,
				Controller:      ControllerYou,
				CauseController: ControllerOpponent,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "combat event binds relation qualifier and recipient",
			event: "equipped creature deals combat damage to an opponent",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:            TriggerWhenever,
				Event:           TriggerEventDamageDealt,
				Source:          TriggerSourceAttachedPermanent,
				Subject:         TriggerSubjectDamageSource,
				Player:          TriggerPlayerOpponent,
				CombatQualifier: TriggerCombatDamage,
				DamageRecipient: TriggerDamageRecipientPlayer,
				DamageSourceSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "state event binds counter and batching",
			event: "one or more -1/-1 counters are put on this permanent",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:      TriggerWhenever,
				Event:     TriggerEventCountersAdded,
				Source:    TriggerSourceSelf,
				Counter:   TriggerCounterMinusOneMinusOne,
				OneOrMore: true,
			},
		},
		{
			name:  "counter event binds controller-scoped subject",
			event: "one or more +1/+1 counters are put on another creature you control",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:        TriggerWhenever,
				Event:       TriggerEventCountersAdded,
				Controller:  ControllerYou,
				ExcludeSelf: true,
				OneOrMore:   true,
				Counter:     TriggerCounterPlusOnePlusOne,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "face-up event binds self with when",
			event: "this creature is turned face up",
			kind:  TriggerWhen,
			want: TriggerPattern{
				Kind:   TriggerWhen,
				Event:  TriggerEventPermanentTurnedFaceUp,
				Source: TriggerSourceSelf,
			},
		},
		{
			name:  "face-up event binds selected permanent",
			event: "a permanent you control is turned face up",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:       TriggerWhenever,
				Event:      TriggerEventPermanentTurnedFaceUp,
				Controller: ControllerYou,
			},
		},
		{
			name:  "face-up event binds attached creature",
			event: "enchanted creature is turned face up",
			kind:  TriggerWhen,
			want: TriggerPattern{
				Kind:   TriggerWhen,
				Event:  TriggerEventPermanentTurnedFaceUp,
				Source: TriggerSourceAttachedPermanent,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "tap event binds subtype and opponent controller",
			event: "a forest an opponent controls becomes tapped",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:       TriggerWhenever,
				Event:      TriggerEventPermanentTapped,
				Controller: ControllerOpponent,
				SubjectSelection: TriggerSelection{
					SubtypesAny: []types.Sub{types.Forest},
				},
			},
		},
		{
			name:  "sacrifice event binds actor and selected subject",
			event: "you sacrifice a clue",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventPermanentSacrificed,
				Player: TriggerPlayerYou,
				SubjectSelection: TriggerSelection{
					SubtypesAny: []types.Sub{types.Clue},
				},
			},
		},
	}
	for i := range tests {
		test := &tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := compileParsedTriggerPattern(test.event, test.kind, shared.Span{}, test.cardName, test.condition)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("pattern = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestTypedTriggerEventsFailClosedOnUnsupportedSlots(t *testing.T) {
	t.Parallel()
	condition := &CompiledCondition{
		Kind:      ConditionIf,
		Predicate: ConditionPredicateUnsupported,
	}
	tests := []struct {
		event     string
		kind      TriggerKind
		condition *CompiledCondition
	}{
		{event: "two or more artifacts you control enter", kind: TriggerWhenever},
		{event: "a creature you or an opponent controls enters", kind: TriggerWhenever},
		{event: "you cast an instant or Wizard spell", kind: TriggerWhenever},
		{event: "you activate a boast ability", kind: TriggerWhenever},
		{event: "you turn a permanent face up", kind: TriggerWhenever},
		{event: "you scry or surveil", kind: TriggerWhenever},
		{event: "this creature becomes the target of an ability", kind: TriggerWhenever},
		{event: "this creature becomes the target of a spell or ability for the first time each turn", kind: TriggerWhenever},
		{event: "the beginning of a player's upkeep", kind: TriggerAt},
		{event: "the beginning of your next upkeep", kind: TriggerAt, condition: condition},
		{event: "you cast a spell", kind: TriggerWhen},
	}
	for _, test := range tests {
		t.Run(test.event, func(t *testing.T) {
			t.Parallel()
			got := compileParsedTriggerPattern(test.event, test.kind, shared.Span{}, "", test.condition)
			want := TriggerPattern{Kind: test.kind, InterveningCondition: test.condition}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("near-miss pattern = %#v, want %#v", got, want)
			}
		})
	}
}

func TestCombatPhaseAndStepTriggerPatternsSaturateRepresentableSlots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		event    string
		kind     TriggerKind
		cardName string
		want     TriggerPattern
	}{
		{
			name:     "named source attacks with when",
			event:    "Example attacks",
			kind:     TriggerWhen,
			cardName: "Example",
			want:     TriggerPattern{Kind: TriggerWhen, Event: TriggerEventAttackerDeclared, Source: TriggerSourceSelf},
		},
		{
			name:  "one or more selected attackers",
			event: "one or more artifact creatures you control attack",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:       TriggerWhenever,
				Event:      TriggerEventAttackerDeclared,
				Controller: ControllerYou,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Artifact, types.Creature},
				},
				OneOrMore: true,
			},
		},
		{
			name:  "attacks exact player or planeswalker recipient",
			event: "a creature attacks you or a planeswalker you control",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:            TriggerWhenever,
				Event:           TriggerEventAttackerDeclared,
				Player:          TriggerPlayerYou,
				AttackRecipient: TriggerAttackRecipientPlayer | TriggerAttackRecipientPlaneswalker,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
				AttackRecipientSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Planeswalker},
					Controller:    ControllerYou,
				},
			},
		},
		{
			name:  "player attack batches per recipient",
			event: "you attack a player",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:                     TriggerWhenever,
				Event:                    TriggerEventAttackerDeclared,
				Controller:               ControllerYou,
				AttackRecipient:          TriggerAttackRecipientPlayer,
				OneOrMore:                true,
				OneOrMorePerAttackTarget: true,
			},
		},
		{
			name:  "self source attacks alone",
			event: "this creature attacks alone",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:        TriggerWhenever,
				Event:       TriggerEventAttackerDeclared,
				Source:      TriggerSourceSelf,
				AttackAlone: true,
			},
		},
		{
			name:  "selected source attacks alone",
			event: "a creature you control attacks alone",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:        TriggerWhenever,
				Event:       TriggerEventAttackerDeclared,
				Controller:  ControllerYou,
				AttackAlone: true,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "controller attacks with two or more creatures",
			event: "you attack with two or more creatures",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:                 TriggerWhenever,
				Event:                TriggerEventAttackerDeclared,
				Controller:           ControllerYou,
				OneOrMore:            true,
				AttackerCountAtLeast: 2,
			},
		},
		{
			name:  "block related Selection",
			event: "this creature blocks a creature with flying",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventBlockerDeclared,
				Source: TriggerSourceSelf,
				RelatedSubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					Keyword:       parser.KeywordFlying,
				},
			},
		},
		{
			name:  "selected combat damage sources batch",
			event: "one or more artifact creatures you control deal combat damage to a player",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:            TriggerWhenever,
				Event:           TriggerEventDamageDealt,
				Subject:         TriggerSubjectDamageSource,
				Controller:      ControllerYou,
				CombatQualifier: TriggerCombatDamage,
				DamageRecipient: TriggerDamageRecipientPlayer,
				DamageSourceSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Artifact, types.Creature},
				},
				OneOrMore: true,
			},
		},
		{
			name:  "instant or sorcery damage source",
			event: "an instant or sorcery spell you control deals damage to an opponent",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:                      TriggerWhenever,
				Event:                     TriggerEventDamageDealt,
				Subject:                   TriggerSubjectDamageSource,
				Controller:                ControllerYou,
				Player:                    TriggerPlayerOpponent,
				StackObject:               TriggerStackObjectSpell,
				DamageSourceIsStackObject: true,
				DamageSourceSelection: TriggerSelection{
					RequiredTypesAny: []types.Card{types.Instant, types.Sorcery},
				},
				DamageRecipient: TriggerDamageRecipientPlayer,
			},
		},
		{
			name:  "noncombat damage exact opponent",
			event: "this creature deals noncombat damage to an opponent",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:            TriggerWhenever,
				Event:           TriggerEventDamageDealt,
				Source:          TriggerSourceSelf,
				Subject:         TriggerSubjectDamageSource,
				Player:          TriggerPlayerOpponent,
				CombatQualifier: TriggerNonCombatDamage,
				DamageRecipient: TriggerDamageRecipientPlayer,
			},
		},
		{
			name:  "damage player or planeswalker union",
			event: "this creature deals damage to a player or planeswalker",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:            TriggerWhenever,
				Event:           TriggerEventDamageDealt,
				Source:          TriggerSourceSelf,
				Subject:         TriggerSubjectDamageSource,
				DamageRecipient: TriggerDamageRecipientPlayer | TriggerDamageRecipientPermanent,
				DamageRecipientSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Planeswalker},
				},
			},
		},
		{
			name:  "any damage source to ability source",
			event: "a source deals damage to this creature",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:                    TriggerWhenever,
				Event:                   TriggerEventDamageDealt,
				Subject:                 TriggerSubjectDamageSource,
				DamageRecipient:         TriggerDamageRecipientPermanent,
				DamageRecipientIsSource: true,
			},
		},
		{
			name:  "selected recipient is dealt combat damage",
			event: "a creature an opponent controls is dealt combat damage",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:            TriggerWhenever,
				Event:           TriggerEventDamageDealt,
				Controller:      ControllerOpponent,
				CombatQualifier: TriggerCombatDamage,
				DamageRecipient: TriggerDamageRecipientPermanent,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
	}
	for i := range tests {
		test := &tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			pattern := compileParsedTriggerPattern(test.event, test.kind, shared.Span{}, test.cardName, nil)
			if !reflect.DeepEqual(pattern, test.want) {
				t.Fatalf("pattern = %#v, want %#v", pattern, test.want)
			}
		})
	}
}

func TestCombatPhaseAndStepTriggerPatternsFailClosedOnMissingCapabilities(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"this creature becomes blocked by a nonblack creature",
		"the beginning of your declare attackers step",
		"the beginning of your next upkeep",
	} {
		kind := TriggerWhenever
		if strings.HasPrefix(event, "the beginning") {
			kind = TriggerAt
		}
		pattern := compileParsedTriggerPattern(event, kind, shared.Span{}, "", nil)
		if pattern.Event != TriggerEventUnknown {
			t.Fatalf("%q pattern = %#v, want unknown event", event, pattern)
		}
	}
}

func TestPermanentZoneChangeTriggerPatternsBindRepresentableSlots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		event    string
		kind     TriggerKind
		cardName string
		want     TriggerPattern
	}{
		{
			name:     "short printed name is self",
			event:    "Sharuum enters",
			kind:     TriggerWhen,
			cardName: "Sharuum the Hegemon",
			want: TriggerPattern{
				Kind:   TriggerWhen,
				Event:  TriggerEventPermanentEnteredBattlefield,
				Source: TriggerSourceSelf,
			},
		},
		{
			name:  "attached subtype leaves for graveyard",
			event: "enchanted Plains is put into a graveyard from the battlefield",
			kind:  TriggerWhen,
			want: TriggerPattern{
				Kind:          TriggerWhen,
				Event:         TriggerEventZoneChanged,
				Source:        TriggerSourceAttachedPermanent,
				MatchFromZone: true,
				FromZone:      TriggerZoneBattlefield,
				MatchToZone:   true,
				ToZone:        TriggerZoneGraveyard,
				SubjectSelection: TriggerSelection{
					SubtypesAny: []types.Sub{types.Plains},
				},
			},
		},
		{
			name:  "batched other creatures die",
			event: "one or more other creatures an opponent controls die",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:        TriggerWhenever,
				Event:       TriggerEventPermanentDied,
				Controller:  ControllerOpponent,
				ExcludeSelf: true,
				OneOrMore:   true,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:     "chosen-type creature enters",
			event:    "a creature you control of the chosen type enters",
			kind:     TriggerWhenever,
			cardName: "Kindred Discovery",
			want: TriggerPattern{
				Kind:       TriggerWhenever,
				Event:      TriggerEventPermanentEnteredBattlefield,
				Controller: ControllerYou,
				SubjectSelection: TriggerSelection{
					RequiredTypes:          []types.Card{types.Creature},
					SubtypeFromEntryChoice: true,
				},
			},
		},
		{
			name:  "qualified subtype enters",
			event: "another nontoken legendary green Dragon you control with power 4 or greater enters",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:        TriggerWhenever,
				Event:       TriggerEventPermanentEnteredBattlefield,
				Controller:  ControllerYou,
				ExcludeSelf: true,
				SubjectSelection: TriggerSelection{
					Supertypes:  []types.Super{types.Legendary},
					SubtypesAny: []types.Sub{types.Dragon},
					ColorsAny:   []color.Color{color.Green},
					NonToken:    true,
					Power:       compare.Int{Op: compare.GreaterOrEqual, Value: 4},
				},
			},
		},
		{
			name:  "keyword and creature dies",
			event: "a creature with flying dies",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:  TriggerWhenever,
				Event: TriggerEventPermanentDied,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					Keyword:       parser.KeywordFlying,
				},
			},
		},
		{
			name:  "power toughness and tapped entry",
			event: "a 1/1 creature you control enters tapped",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:       TriggerWhenever,
				Event:      TriggerEventPermanentEnteredBattlefield,
				Controller: ControllerYou,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					Tapped:        TriggerTriTrue,
					Power:         compare.Int{Op: compare.Equal, Value: 1},
					Toughness:     compare.Int{Op: compare.Equal, Value: 1},
				},
			},
		},
		{
			name:  "type union and untapped entry",
			event: "an artifact or creature an opponent controls enters untapped",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:       TriggerWhenever,
				Event:      TriggerEventPermanentEnteredBattlefield,
				Controller: ControllerOpponent,
				SubjectSelection: TriggerSelection{
					RequiredTypesAny: []types.Card{types.Artifact, types.Creature},
					Tapped:           TriggerTriFalse,
				},
			},
		},
		{
			name:  "owner-relative origin",
			event: "a creature enters from your graveyard",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:          TriggerWhenever,
				Event:         TriggerEventPermanentEnteredBattlefield,
				Player:        TriggerPlayerYou,
				MatchFromZone: true,
				FromZone:      TriggerZoneGraveyard,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "controller and owner relations",
			event: "a permanent you control but don't own leaves the battlefield",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:          TriggerWhenever,
				Event:         TriggerEventZoneChanged,
				Controller:    ControllerYou,
				Player:        TriggerPlayerOpponent,
				MatchFromZone: true,
				FromZone:      TriggerZoneBattlefield,
			},
		},
		{
			name:  "owner-relative destination",
			event: "an artifact is returned to your hand",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:          TriggerWhenever,
				Event:         TriggerEventZoneChanged,
				Player:        TriggerPlayerYou,
				MatchFromZone: true,
				FromZone:      TriggerZoneBattlefield,
				MatchToZone:   true,
				ToZone:        TriggerZoneHand,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Artifact},
				},
			},
		},
		{
			name:  "nonblack entry",
			event: "a nonblack creature enters",
			kind:  TriggerWhen,
			want: TriggerPattern{
				Kind:  TriggerWhen,
				Event: TriggerEventPermanentEnteredBattlefield,
				SubjectSelection: TriggerSelection{
					RequiredTypes:  []types.Card{types.Creature},
					ExcludedColors: []color.Color{color.Black},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := compileParsedTriggerPattern(test.event, test.kind, shared.Span{}, test.cardName, nil)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("pattern = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestPermanentZoneChangeTriggerPatternsBindExtendedSlots(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		event    string
		cardName string
		want     TriggerPattern
	}{
		{
			name:  "leaves without dying",
			event: "a creature leaves the battlefield without dying",
			want: TriggerPattern{
				Event:         TriggerEventZoneChanged,
				MatchFromZone: true,
				FromZone:      TriggerZoneBattlefield,
				ExcludeToZone: true,
				ToZone:        TriggerZoneGraveyard,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "face-down creature dies",
			event: "a face-down creature dies",
			want: TriggerPattern{
				Event:         TriggerEventPermanentDied,
				MatchFaceDown: true,
				FaceDown:      true,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "attacking creature dies",
			event: "an attacking creature dies",
			want: TriggerPattern{
				Event: TriggerEventPermanentDied,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					CombatState:   TriggerCombatStateAttacking,
				},
			},
		},
		{
			name:  "subtype and type",
			event: "a Dragon creature dies",
			want: TriggerPattern{
				Event: TriggerEventPermanentDied,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					SubtypesAny:   []types.Sub{types.Dragon},
				},
			},
		},
		{
			name:  "composite outlaw subtype",
			event: "an outlaw you control dies",
			want: TriggerPattern{
				Event:      TriggerEventPermanentDied,
				Controller: ControllerYou,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					SubtypesAny: []types.Sub{
						types.Assassin, types.Mercenary, types.Pirate, types.Rogue, types.Warlock,
					},
				},
			},
		},
		{
			name:  "noun suffix token",
			event: "a creature token you control dies",
			want: TriggerPattern{
				Event:      TriggerEventPermanentDied,
				Controller: ControllerYou,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					TokenOnly:     true,
				},
			},
		},
		{
			name:     "other than named self",
			cardName: "Yomiji, Who Bars the Way",
			event:    "a legendary permanent other than Yomiji dies",
			want: TriggerPattern{
				Event:       TriggerEventPermanentDied,
				ExcludeSelf: true,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
					Supertypes:    []types.Super{types.Legendary},
				},
			},
		},
		{
			name:  "card put into your graveyard from anywhere",
			event: "a creature card is put into your graveyard from anywhere",
			want: TriggerPattern{
				Event:       TriggerEventZoneChanged,
				Player:      TriggerPlayerYou,
				MatchToZone: true,
				ToZone:      TriggerZoneGraveyard,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
		{
			name:  "card put into graveyard excluding battlefield",
			event: "a creature card is put into a graveyard from anywhere other than the battlefield",
			want: TriggerPattern{
				Event:           TriggerEventZoneChanged,
				MatchToZone:     true,
				ToZone:          TriggerZoneGraveyard,
				ExcludeFromZone: true,
				FromZone:        TriggerZoneBattlefield,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []types.Card{types.Creature},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.want.Kind = TriggerWhenever
			got := compileParsedTriggerPattern(test.event, TriggerWhenever, shared.Span{}, test.cardName, nil)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("pattern = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestPermanentZoneChangeTriggerPatternsRejectMissingRuntimeSlots(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"a creature you control with a +1/+1 counter on it dies",
		"a non-Human creature dies",
		"a creature dies during your turn",
	} {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			got := compileParsedTriggerPattern(event, TriggerWhenever, shared.Span{}, "", nil)
			want := TriggerPattern{Kind: TriggerWhenever}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("near-miss pattern = %#v, want %#v", got, want)
			}
		})
	}
}

func TestPermanentZoneChangeTriggerRejectsPartialCardName(t *testing.T) {
	t.Parallel()
	got := compileParsedTriggerPattern("The dies", TriggerWhen, shared.Span{}, "The One Ring", nil)
	want := TriggerPattern{Kind: TriggerWhen}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("partial-name pattern = %#v, want %#v", got, want)
	}
}

func TestCapitalizedEquipmentFaceUpTriggerPattern(t *testing.T) {
	t.Parallel()
	got := compileParsedTriggerPattern("this Equipment is turned face up", TriggerWhen, shared.Span{}, "", nil)
	want := TriggerPattern{
		Kind:   TriggerWhen,
		Event:  TriggerEventPermanentTurnedFaceUp,
		Source: TriggerSourceSelf,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("pattern = %#v, want %#v", got, want)
	}
}

func TestActivatedAbilityTriggerPatternsRequireNonManaExclusion(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"you activate an ability",
		"a player activates an ability",
		"an opponent activates an ability of a creature",
	} {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			got := compileParsedTriggerPattern(event, TriggerWhenever, shared.Span{}, "", nil)
			want := TriggerPattern{Kind: TriggerWhenever}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("unrestricted activation pattern = %#v, want %#v", got, want)
			}
		})
	}

	got := compileParsedTriggerPattern(
		"an opponent activates an ability of a creature that isn't a mana ability",
		TriggerWhenever,
		shared.Span{},
		"",
		nil,
	)
	if got.Event != TriggerEventAbilityActivated || !got.ExcludeManaAbility {
		t.Fatalf("non-mana activation pattern = %#v", got)
	}
}

func TestTypedTriggerEventPathDoesNotRecognizePlayerEvents(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"you draw a card",
		"you discard one or more cards",
		"a player cycles a card",
		"you scry",
		"an opponent gains life",
		"you draw your second card each turn",
		"you surveil for the first time each turn",
	} {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			got := compileParsedTriggerPattern(event, TriggerWhenever, shared.Span{}, "", nil)
			if got.Event != TriggerEventUnknown {
				t.Fatalf("pattern = %#v, want text-blind unknown event", got)
			}
		})
	}
}

func TestCompileConstructedTriggerEventIsTextBlind(t *testing.T) {
	t.Parallel()
	trigger := compileTrigger(&parser.Ability{
		Kind: parser.AbilityTriggered,
		Trigger: &parser.TriggerClause{
			Introduction: parser.TriggerIntroduction{Kind: parser.TriggerIntroductionWhenever},
			Event:        "text must not determine trigger meaning",
			TriggerEvent: &parser.TriggerEventClause{
				Kind:  parser.TriggerEventKindSpellCast,
				Actor: parser.TriggerEventActor{Kind: parser.TriggerEventActorOpponent},
			},
		},
	}, Context{})
	want := TriggerPattern{
		Kind:       TriggerWhenever,
		Event:      TriggerEventSpellCast,
		Controller: ControllerOpponent,
	}
	if !reflect.DeepEqual(trigger.Pattern, want) {
		t.Fatalf("pattern = %#v, want %#v", trigger.Pattern, want)
	}
}

func TestCompileConstructedTriggerEventsFailClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		clause parser.TriggerEventClause
	}{
		{
			name: "partial zone destination",
			clause: parser.TriggerEventClause{
				Kind:       parser.TriggerEventKindZoneChange,
				ZoneChange: parser.TriggerEventZoneChange{Kind: parser.TriggerEventZoneChangeMoved},
				Subject: parser.TriggerEventSubject{
					Kind:      parser.TriggerEventSubjectSelection,
					Selection: parser.TriggerSelection{RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeCreature}},
				},
				Zone: parser.TriggerEventZoneContext{MatchToZone: true},
			},
		},
		{
			name:   "missing spell actor",
			clause: parser.TriggerEventClause{Kind: parser.TriggerEventKindSpellCast},
		},
		{
			name: "missing counter kind",
			clause: parser.TriggerEventClause{
				Kind:    parser.TriggerEventKindCounterAdded,
				Subject: parser.TriggerEventSubject{Kind: parser.TriggerEventSubjectSelf},
			},
		},
		{
			name: "unsupported targeting stack object",
			clause: parser.TriggerEventClause{
				Kind:    parser.TriggerEventKindBecameTarget,
				Subject: parser.TriggerEventSubject{Kind: parser.TriggerEventSubjectSelf},
				StackObject: parser.TriggerEventStackObject{
					Kind: parser.TriggerEventStackObjectKind("invalid"),
				},
			},
		},
		{
			name: "damage source spell with kicker qualifier",
			clause: parser.TriggerEventClause{
				Kind:                      parser.TriggerEventKindDamageDealt,
				DamageSourceIsStackObject: true,
				DamageSourceSpellSelection: parser.TriggerEventSpellSelection{
					Kicker: true,
				},
				StackObject:     parser.TriggerEventStackObject{Kind: parser.TriggerEventStackObjectSpell},
				DamageRecipient: parser.TriggerEventDamageRecipient{Kind: parser.TriggerEventDamageRecipientPlayer},
			},
		},
		{
			name: "damage source spell with historic qualifier",
			clause: parser.TriggerEventClause{
				Kind:                      parser.TriggerEventKindDamageDealt,
				DamageSourceIsStackObject: true,
				DamageSourceSpellSelection: parser.TriggerEventSpellSelection{
					Historic: true,
				},
				StackObject:     parser.TriggerEventStackObject{Kind: parser.TriggerEventStackObjectSpell},
				DamageRecipient: parser.TriggerEventDamageRecipient{Kind: parser.TriggerEventDamageRecipientPlayer},
			},
		},
		{
			name: "damage source spell with origin zone",
			clause: parser.TriggerEventClause{
				Kind:                      parser.TriggerEventKindDamageDealt,
				DamageSourceIsStackObject: true,
				DamageSourceSpellSelection: parser.TriggerEventSpellSelection{
					FromZone: parser.TriggerEventZone{Kind: parser.TriggerEventZoneGraveyard},
				},
				StackObject:     parser.TriggerEventStackObject{Kind: parser.TriggerEventStackObjectSpell},
				DamageRecipient: parser.TriggerEventDamageRecipient{Kind: parser.TriggerEventDamageRecipientPlayer},
			},
		},
	}
	for i := range tests {
		test := &tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trigger := compileTrigger(&parser.Ability{
				Kind: parser.AbilityTriggered,
				Trigger: &parser.TriggerClause{
					Introduction: parser.TriggerIntroduction{Kind: parser.TriggerIntroductionWhenever},
					Event:        "this creature attacks",
					TriggerEvent: &test.clause,
				},
			}, Context{})
			if trigger.Pattern.Event != TriggerEventUnknown {
				t.Fatalf("pattern = %#v, want unknown event", trigger.Pattern)
			}
		})
	}
}

func TestTypedTriggerEventsPreserveSpan(t *testing.T) {
	t.Parallel()
	span := shared.Span{
		Start: shared.Position{Offset: 5, Line: 2, Column: 3},
		End:   shared.Position{Offset: 28, Line: 2, Column: 26},
	}
	pattern := compileParsedTriggerPattern("this creature attacks", TriggerWhenever, span, "", nil)
	if pattern.Span != span {
		t.Fatalf("span = %#v, want %#v", pattern.Span, span)
	}
}
