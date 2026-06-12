package oracle

import (
	"reflect"
	"strings"
	"testing"
)

func TestTriggerPatternTemplatesBindClosedSlots(t *testing.T) {
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact},
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
				SubjectSelection:     TriggerSelection{RequiredTypes: []TriggerCardType{TriggerCardTypeCreature}},
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
					RequiredTypesAny: []TriggerCardType{TriggerCardTypeCreature, TriggerCardTypeLand},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
				},
			},
		},
		{
			name:      "phase template binds relation step and condition",
			event:     "the beginning of each opponent's postcombat main phase",
			kind:      TriggerAt,
			condition: condition,
			want: TriggerPattern{
				Kind:                 TriggerAt,
				Event:                TriggerEventBeginningOfStep,
				Controller:           ControllerOpponent,
				Step:                 TriggerStepPostcombatMain,
				InterveningCondition: condition,
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
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
					SubtypesAny: []TriggerSubtype{"forest"},
				},
			},
		},
		{
			name:  "player event binds relation and batching",
			event: "you discard one or more cards",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:      TriggerWhenever,
				Event:     TriggerEventCardDiscarded,
				Player:    TriggerPlayerYou,
				OneOrMore: true,
			},
		},
		{
			name:  "player event binds any-player cycling",
			event: "a player cycles a card",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventCycled,
				Player: TriggerPlayerAny,
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
					SubtypesAny: []TriggerSubtype{"clue"},
				},
			},
		},
		{
			name:  "player event binds scry",
			event: "you scry",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventScry,
				Player: TriggerPlayerYou,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := compileTriggerPattern(test.event, test.kind, Span{}, test.cardName, test.condition)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("pattern = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestTriggerPatternTemplatesFailClosedOnUnsupportedSlots(t *testing.T) {
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
		{event: "you cast a creature or artifact spell", kind: TriggerWhenever},
		{event: "you activate a boast ability", kind: TriggerWhenever},
		{event: "you turn a permanent face up", kind: TriggerWhenever},
		{event: "you create or sacrifice a token", kind: TriggerWhenever},
		{event: "you scry or surveil", kind: TriggerWhenever},
		{event: "this creature becomes the target of an ability", kind: TriggerWhenever},
		{event: "this creature becomes the target of a spell or ability for the first time each turn", kind: TriggerWhenever},
		{event: "this creature attacks alone", kind: TriggerWhenever},
		{event: "the beginning of a player's upkeep", kind: TriggerAt},
		{event: "the beginning of your next upkeep", kind: TriggerAt, condition: condition},
		{event: "you cast a spell", kind: TriggerWhen},
	}
	for _, test := range tests {
		t.Run(test.event, func(t *testing.T) {
			t.Parallel()
			got := compileTriggerPattern(test.event, test.kind, Span{}, "", test.condition)
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
				},
				AttackRecipientSelection: TriggerSelection{
					RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker},
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
			name:  "block related Selection",
			event: "this creature blocks a creature with flying",
			kind:  TriggerWhenever,
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventBlockerDeclared,
				Source: TriggerSourceSelf,
				RelatedSubjectSelection: TriggerSelection{
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
					Keyword:       TriggerKeywordFlying,
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature},
				},
				OneOrMore: true,
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
					RequiredTypes: []TriggerCardType{TriggerCardTypePlaneswalker},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
				},
			},
		},
		{
			name:  "all player end step",
			event: "the beginning of the end step",
			kind:  TriggerAt,
			want:  TriggerPattern{Kind: TriggerAt, Event: TriggerEventBeginningOfStep, Step: TriggerStepEnd},
		},
		{
			name:  "opponent first main phase",
			event: "the beginning of each opponent's first main phase",
			kind:  TriggerAt,
			want:  TriggerPattern{Kind: TriggerAt, Event: TriggerEventBeginningOfStep, Controller: ControllerOpponent, Step: TriggerStepPrecombatMain},
		},
		{
			name:  "at end of combat",
			event: "end of combat",
			kind:  TriggerAt,
			want:  TriggerPattern{Kind: TriggerAt, Event: TriggerEventBeginningOfStep, Step: TriggerStepEndOfCombat},
		},
		{
			name:  "attached permanent controller upkeep",
			event: "the beginning of the upkeep of enchanted creature's controller",
			kind:  TriggerAt,
			want: TriggerPattern{
				Kind:  TriggerAt,
				Event: TriggerEventBeginningOfStep,
				Step:  TriggerStepUpkeep,
				StepPlayerSourceAttachedSelection: TriggerSelection{
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			pattern := compileTriggerPattern(test.event, test.kind, Span{}, test.cardName, nil)
			if !reflect.DeepEqual(pattern, test.want) {
				t.Fatalf("pattern = %#v, want %#v", pattern, test.want)
			}
		})
	}
}

func TestCombatPhaseAndStepTriggerPatternsFailClosedOnMissingCapabilities(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"this creature attacks alone",
		"this creature becomes blocked by a nonblack creature",
		"you attack with two or more creatures",
		"the beginning of your declare attackers step",
		"the beginning of your next upkeep",
	} {
		kind := TriggerWhenever
		if strings.HasPrefix(event, "the beginning") {
			kind = TriggerAt
		}
		pattern := compileTriggerPattern(event, kind, Span{}, "", nil)
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
					SubtypesAny: []TriggerSubtype{"plains"},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
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
					Supertypes:  []TriggerSupertype{TriggerSupertypeLegendary},
					SubtypesAny: []TriggerSubtype{"dragon"},
					ColorsAny:   []TriggerColor{TriggerColorGreen},
					NonToken:    true,
					Power:       TriggerNumberFilter{Comparison: TriggerComparisonAtLeast, Value: 4},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
					Keyword:       TriggerKeywordFlying,
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
					Tapped:        TriggerTriTrue,
					Power:         TriggerNumberFilter{Comparison: TriggerComparisonEqual, Value: 1},
					Toughness:     TriggerNumberFilter{Comparison: TriggerComparisonEqual, Value: 1},
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
					RequiredTypesAny: []TriggerCardType{TriggerCardTypeArtifact, TriggerCardTypeCreature},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeArtifact},
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
					RequiredTypes:  []TriggerCardType{TriggerCardTypeCreature},
					ExcludedColors: []TriggerColor{TriggerColorBlack},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := compileTriggerPattern(test.event, test.kind, Span{}, test.cardName, nil)
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
				},
			},
		},
		{
			name:  "attacking creature dies",
			event: "an attacking creature dies",
			want: TriggerPattern{
				Event: TriggerEventPermanentDied,
				SubjectSelection: TriggerSelection{
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
					SubtypesAny:   []TriggerSubtype{"dragon"},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
					SubtypesAny: []TriggerSubtype{
						"assassin", "mercenary", "pirate", "rogue", "warlock",
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
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
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
					Supertypes:    []TriggerSupertype{TriggerSupertypeLegendary},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.want.Kind = TriggerWhenever
			got := compileTriggerPattern(test.event, TriggerWhenever, Span{}, test.cardName, nil)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("pattern = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestPermanentZoneChangeTriggerPatternsRejectMissingRuntimeSlots(t *testing.T) {
	t.Parallel()
	for _, event := range []string{
		"this creature or another creature you control enters",
		"a creature you control with a +1/+1 counter on it dies",
		"a non-Human creature dies",
		"a creature dies during your turn",
		"a creature card is put into your graveyard from anywhere",
	} {
		t.Run(event, func(t *testing.T) {
			t.Parallel()
			got := compileTriggerPattern(event, TriggerWhenever, Span{}, "", nil)
			want := TriggerPattern{Kind: TriggerWhenever}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("near-miss pattern = %#v, want %#v", got, want)
			}
		})
	}
}

func TestPermanentZoneChangeTriggerRejectsPartialCardName(t *testing.T) {
	t.Parallel()
	got := compileTriggerPattern("The dies", TriggerWhen, Span{}, "The One Ring", nil)
	want := TriggerPattern{Kind: TriggerWhen}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("partial-name pattern = %#v, want %#v", got, want)
	}
}

func TestCapitalizedEquipmentFaceUpTriggerPattern(t *testing.T) {
	t.Parallel()
	got := compileTriggerPattern("this Equipment is turned face up", TriggerWhen, Span{}, "", nil)
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
			got := compileTriggerPattern(event, TriggerWhenever, Span{}, "", nil)
			want := TriggerPattern{Kind: TriggerWhenever}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("unrestricted activation pattern = %#v, want %#v", got, want)
			}
		})
	}

	got := compileTriggerPattern(
		"an opponent activates an ability of a creature that isn't a mana ability",
		TriggerWhenever,
		Span{},
		"",
		nil,
	)
	if got.Event != TriggerEventAbilityActivated || !got.ExcludeManaAbility {
		t.Fatalf("non-mana activation pattern = %#v", got)
	}
}

func TestPlayerOrdinalTriggerPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		event   string
		kind    TriggerKind
		want    TriggerEvent
		player  TriggerPlayerRelation
		ordinal int
	}{
		{event: "you draw your second card each turn", kind: TriggerWhenever, want: TriggerEventCardDrawn, player: TriggerPlayerYou, ordinal: 2},
		{event: "an opponent draws their first card each turn", kind: TriggerWhenever, want: TriggerEventCardDrawn, player: TriggerPlayerOpponent, ordinal: 1},
		{event: "you gain life for the first time each turn", kind: TriggerWhenever, want: TriggerEventLifeGained, player: TriggerPlayerYou, ordinal: 1},
		{event: "you lose life for the first time each turn", kind: TriggerWhen, want: TriggerEventLifeLost, player: TriggerPlayerYou, ordinal: 1},
		{event: "you surveil for the first time each turn", kind: TriggerWhenever, want: TriggerEventSurveil, player: TriggerPlayerYou, ordinal: 1},
	}
	for _, test := range tests {
		t.Run(test.event, func(t *testing.T) {
			t.Parallel()
			got := compileTriggerPattern(test.event, test.kind, Span{}, "", nil)
			if got.Event != test.want || got.Player != test.player || got.PlayerEventOrdinalThisTurn != test.ordinal {
				t.Fatalf("pattern = %#v", got)
			}
		})
	}
}

func TestTriggerPatternTemplatesFailClosedOnOverlappingTemplates(t *testing.T) {
	t.Parallel()
	templates := []triggerPatternTemplate{
		{
			kinds: []TriggerKind{TriggerWhenever},
			bind: func(string, TriggerKind, string) (TriggerPattern, bool) {
				return TriggerPattern{Event: TriggerEventSpellCast}, true
			},
		},
		{
			kinds: []TriggerKind{TriggerWhenever},
			bind: func(string, TriggerKind, string) (TriggerPattern, bool) {
				return TriggerPattern{Event: TriggerEventCardDrawn}, true
			},
		},
	}
	if pattern, ok := bindTriggerPatternTemplates("ambiguous", TriggerWhenever, "", templates); ok {
		t.Fatalf("overlapping templates returned pattern %#v", pattern)
	}
}

func TestTriggerPatternTemplatesPreserveSpan(t *testing.T) {
	t.Parallel()
	span := Span{
		Start: Position{Offset: 5, Line: 2, Column: 3},
		End:   Position{Offset: 28, Line: 2, Column: 26},
	}
	pattern := compileTriggerPattern("this creature attacks", TriggerWhenever, span, "", nil)
	if pattern.Span != span {
		t.Fatalf("span = %#v, want %#v", pattern.Span, span)
	}
}
