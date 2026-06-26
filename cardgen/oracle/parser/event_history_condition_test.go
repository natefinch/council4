package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestEventHistoryConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		condition   string
		window      EventHistoryWindowKind
		triggerKind TriggerEventKind
		playerEvent PlayerEventActionKind
		player      TriggerPlayerSelectorKind
		negated     bool
		minCount    int
	}{
		{
			name:        "controller attacked current turn",
			condition:   "you attacked this turn",
			window:      EventHistoryWindowCurrentTurn,
			triggerKind: TriggerEventKindAttack,
		},
		{
			name:        "controller attacked with two or more creatures current turn",
			condition:   "you attacked with two or more creatures this turn",
			window:      EventHistoryWindowCurrentTurn,
			triggerKind: TriggerEventKindAttack,
			minCount:    2,
		},
		{
			name:        "controller attacked with a creature current turn",
			condition:   "you attacked with a creature this turn",
			window:      EventHistoryWindowCurrentTurn,
			triggerKind: TriggerEventKindAttack,
			minCount:    1,
		},
		{
			name:        "creature died current turn",
			condition:   "a creature died this turn",
			window:      EventHistoryWindowCurrentTurn,
			triggerKind: TriggerEventKindZoneChange,
		},
		{
			name:        "controller gained life current turn",
			condition:   "you gained life this turn",
			window:      EventHistoryWindowCurrentTurn,
			playerEvent: PlayerEventActionGainLife,
			player:      TriggerPlayerSelectorYou,
		},
		{
			name:        "opponent lost life current turn",
			condition:   "an opponent lost life this turn",
			window:      EventHistoryWindowCurrentTurn,
			playerEvent: PlayerEventActionLoseLife,
			player:      TriggerPlayerSelectorOpponent,
		},
		{
			name:        "controller lost life previous turn",
			condition:   "you lost life last turn",
			window:      EventHistoryWindowPreviousTurn,
			playerEvent: PlayerEventActionLoseLife,
			player:      TriggerPlayerSelectorYou,
		},
		{
			name:        "no spells previous turn",
			condition:   "no spells were cast last turn",
			window:      EventHistoryWindowPreviousTurn,
			triggerKind: TriggerEventKindSpellCast,
			negated:     true,
		},
	}
	for i := range tests {
		test := &tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(
				"When this creature enters, if "+test.condition+", draw a card.",
				Context{},
			)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 || len(document.Abilities[0].EventHistoryConditions) != 1 {
				t.Fatalf("event history conditions = %#v", document.Abilities)
			}
			condition := &document.Abilities[0].EventHistoryConditions[0]
			if condition.Span == (condition.Window.Span) || condition.Span.Start == condition.Span.End {
				t.Fatalf("condition span = %#v, window span = %#v", condition.Span, condition.Window.Span)
			}
			if condition.Window.Kind != test.window || condition.Negated != test.negated {
				t.Fatalf("condition = %#v", condition)
			}
			if condition.MinCount != test.minCount {
				t.Fatalf("condition MinCount = %d, want %d", condition.MinCount, test.minCount)
			}
			if test.triggerKind != TriggerEventKindUnknown {
				if condition.TriggerEvent == nil || condition.TriggerEvent.Kind != test.triggerKind ||
					condition.PlayerEvent != nil {
					t.Fatalf("condition = %#v", condition)
				}
				return
			}
			if condition.PlayerEvent == nil ||
				condition.PlayerEvent.Action.Kind != test.playerEvent ||
				condition.PlayerEvent.Player.Kind != test.player ||
				condition.TriggerEvent != nil {
				t.Fatalf("condition = %#v", condition)
			}
		})
	}
}

func TestEventHistoryLeftBattlefieldRevolt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		condition     string
		requiredTypes []TriggerCardType
	}{
		{
			name:      "any permanent",
			condition: "a permanent left the battlefield under your control",
		},
		{
			name:          "creature only",
			condition:     "a creature left the battlefield under your control",
			requiredTypes: []TriggerCardType{TriggerCardTypeCreature},
		},
	}
	for i := range tests {
		test := &tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(
				"When this creature enters, if "+test.condition+" this turn, draw a card.",
				Context{},
			)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 || len(document.Abilities[0].EventHistoryConditions) != 1 {
				t.Fatalf("event history conditions = %#v", document.Abilities)
			}
			condition := &document.Abilities[0].EventHistoryConditions[0]
			if condition.Window.Kind != EventHistoryWindowCurrentTurn || condition.Negated {
				t.Fatalf("condition = %#v", condition)
			}
			event := condition.TriggerEvent
			if event == nil || condition.PlayerEvent != nil {
				t.Fatalf("condition = %#v", condition)
			}
			if event.Kind != TriggerEventKindZoneChange {
				t.Fatalf("event kind = %q, want %q", event.Kind, TriggerEventKindZoneChange)
			}
			if event.Controller != ControllerYou {
				t.Fatalf("event controller = %q, want %q", event.Controller, ControllerYou)
			}
			if event.ZoneChange.Kind != TriggerEventZoneChangeMoved {
				t.Fatalf("zone change kind = %q, want %q", event.ZoneChange.Kind, TriggerEventZoneChangeMoved)
			}
			if !event.Zone.MatchFromZone || event.Zone.FromZone.Kind != TriggerEventZoneBattlefield {
				t.Fatalf("zone context = %#v", event.Zone)
			}
			if event.Zone.MatchToZone {
				t.Fatalf("zone context should not match a destination: %#v", event.Zone)
			}
			got := event.Subject.Selection.RequiredTypes
			if len(got) != len(test.requiredTypes) {
				t.Fatalf("required types = %#v, want %#v", got, test.requiredTypes)
			}
			for j := range got {
				if got[j] != test.requiredTypes[j] {
					t.Fatalf("required types = %#v, want %#v", got, test.requiredTypes)
				}
			}
		})
	}
}

func TestEventHistoryConditionActivationOnlyIfSpan(t *testing.T) {
	t.Parallel()
	const activationSource = "{1}: Draw a card. Activate only if you attacked this turn."
	activation, diagnostics := Parse(activationSource, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := &activation.Abilities[0]
	if len(ability.EventHistoryConditions) != 1 || len(ability.ConditionSegments) != 1 {
		t.Fatalf("event histories = %#v, segments = %#v", ability.EventHistoryConditions, ability.ConditionSegments)
	}
	history := ability.EventHistoryConditions[0]
	segment := ability.ConditionSegments[0]
	if history.Span != segment.Span {
		t.Fatalf("history span %#v != segment span %#v", history.Span, segment.Span)
	}
	if segment.EventHistoryIndex != 0 {
		t.Fatalf("segment EventHistoryIndex = %d, want 0", segment.EventHistoryIndex)
	}
	if got := shared.SliceSpan(activationSource, history.Span); got != "only if you attacked this turn" {
		t.Fatalf("history span text = %q, want %q", got, "only if you attacked this turn")
	}

	const interveningSource = "When this creature enters, if you attacked this turn, draw a card."
	intervening, diagnostics := Parse(interveningSource, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	bare := intervening.Abilities[0].EventHistoryConditions[0]
	if got := shared.SliceSpan(interveningSource, bare.Span); got != "if you attacked this turn" {
		t.Fatalf("intervening span text = %q, want %q", got, "if you attacked this turn")
	}
}

func TestEventHistoryConditionsFailClosed(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"you attacked during this turn",
		"a permanent died this turn",
		"an opponent gained life this turn",
		"spells were cast last turn",
		"no spells were cast this turn",
		"you lost life",
		"a permanent left the battlefield",
		"a permanent left the battlefield under an opponent's control",
		"an artifact left the battlefield under your control",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(
				"When this creature enters, if "+condition+", draw a card.",
				Context{},
			)
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %d", len(document.Abilities))
			}
			if got := document.Abilities[0].EventHistoryConditions; len(got) != 0 {
				t.Fatalf("event history conditions = %#v, want none", got)
			}
		})
	}
}

func TestEventHistoryYouCastSpellConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		condition     string
		minCount      int
		excludedTypes []TriggerCardType
		typesAny      []TriggerCardType
	}{
		{
			name:          "noncreature spell",
			condition:     "you've cast a noncreature spell this turn",
			excludedTypes: []TriggerCardType{TriggerCardTypeCreature},
		},
		{
			name:      "instant or sorcery spell",
			condition: "you've cast an instant or sorcery spell this turn",
			typesAny:  []TriggerCardType{TriggerCardTypeInstant, TriggerCardTypeSorcery},
		},
		{
			name:      "two or more spells",
			condition: "you've cast two or more spells this turn",
			minCount:  2,
		},
		{
			name:      "bare you cast a spell",
			condition: "you cast a spell this turn",
		},
		{
			name:      "you have cast a spell",
			condition: "you have cast a spell this turn",
		},
	}
	for i := range tests {
		test := &tests[i]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			source := "{T}: Draw a card. Activate only if " + test.condition + "."
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 || len(document.Abilities[0].EventHistoryConditions) != 1 {
				t.Fatalf("event history conditions = %#v", document.Abilities)
			}
			condition := &document.Abilities[0].EventHistoryConditions[0]
			if condition.Window.Kind != EventHistoryWindowCurrentTurn || condition.Negated {
				t.Fatalf("condition = %#v", condition)
			}
			if condition.MinCount != test.minCount {
				t.Fatalf("condition MinCount = %d, want %d", condition.MinCount, test.minCount)
			}
			event := condition.TriggerEvent
			if event == nil || condition.PlayerEvent != nil {
				t.Fatalf("condition = %#v", condition)
			}
			if event.Kind != TriggerEventKindSpellCast {
				t.Fatalf("event kind = %q, want %q", event.Kind, TriggerEventKindSpellCast)
			}
			if event.Actor.Kind != TriggerEventActorYou {
				t.Fatalf("event actor = %q, want %q", event.Actor.Kind, TriggerEventActorYou)
			}
			if got := event.SpellSelection.ExcludedTypes; !equalTriggerCardTypes(got, test.excludedTypes) {
				t.Fatalf("excluded types = %#v, want %#v", got, test.excludedTypes)
			}
			if got := event.SpellSelection.TypesAny; !equalTriggerCardTypes(got, test.typesAny) {
				t.Fatalf("types any = %#v, want %#v", got, test.typesAny)
			}
		})
	}
}

func TestEventHistoryYouCastSpellFailClosed(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"you've cast a spell last turn",
		"you've cast one or more spells this turn",
		"an opponent cast a spell this turn",
		"you've cast a creature spell",
	} {
		t.Run(condition, func(t *testing.T) {
			t.Parallel()
			source := "{T}: Draw a card. Activate only if " + condition + "."
			document, diagnostics := Parse(source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %d", len(document.Abilities))
			}
			if got := document.Abilities[0].EventHistoryConditions; len(got) != 0 {
				t.Fatalf("event history conditions = %#v, want none", got)
			}
		})
	}
}

func equalTriggerCardTypes(got, want []TriggerCardType) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
