package parser

import "testing"

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
	}{
		{
			name:        "controller attacked current turn",
			condition:   "you attacked this turn",
			window:      EventHistoryWindowCurrentTurn,
			triggerKind: TriggerEventKindAttack,
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

func TestEventHistoryConditionsFailClosed(t *testing.T) {
	t.Parallel()
	for _, condition := range []string{
		"you attacked during this turn",
		"a permanent died this turn",
		"an opponent gained life this turn",
		"spells were cast last turn",
		"no spells were cast this turn",
		"you lost life",
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
