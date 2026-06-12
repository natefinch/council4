package oracle

import (
	"reflect"
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
		{event: "you activate an ability", kind: TriggerWhenever},
		{event: "this creature becomes the target of an ability", kind: TriggerWhenever},
		{event: "this creature attacks alone", kind: TriggerWhenever},
		{event: "one or more creatures you control attack", kind: TriggerWhenever},
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
