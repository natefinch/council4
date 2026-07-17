package parser

import "testing"

func TestCompoundTakeInitiativeAndCreateParsesAsOrderedEffects(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"When this enchantment enters, you take the initiative and create a Treasure token.",
		Context{CardName: "Example"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("document = %#v", document)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 2 ||
		effects[0].Kind != EffectTakeInitiative ||
		effects[1].Kind != EffectCreate ||
		!effects[0].Exact ||
		!effects[1].Exact {
		t.Fatalf("effects = %#v, want exact initiative then create", effects)
	}
}

func TestLootDisputeTriggerEventsParseCompositionally(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		check  func(*testing.T, *TriggerEventClause)
	}{
		{
			name:   "attack initiative holder",
			source: "Whenever you attack the player who has the initiative, create a Treasure token.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindAttack ||
					clause.Actor.Kind != TriggerEventActorYou ||
					clause.Player.Kind != TriggerPlayerSelectorInitiative ||
					clause.AttackRecipient.Kind != TriggerEventAttackRecipientPlayer ||
					!clause.OneOrMore ||
					!clause.OneOrMorePerAttackTarget {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
		{
			name:   "controller completes dungeon",
			source: "Loud Ruckus — Whenever you complete a dungeon, create a 5/5 red Dragon creature token with flying.",
			check: func(t *testing.T, clause *TriggerEventClause) {
				t.Helper()
				if clause.Kind != TriggerEventKindCompletedDungeon ||
					clause.Player.Kind != TriggerPlayerSelectorYou {
					t.Fatalf("clause = %#v", clause)
				}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			abilities := document.Abilities
			if len(abilities) != 1 || abilities[0].Trigger == nil ||
				abilities[0].Trigger.TriggerEvent == nil {
				t.Fatalf("abilities = %#v", abilities)
			}
			test.check(t, abilities[0].Trigger.TriggerEvent)
		})
	}
}

func TestCompletedDungeonTriggerPlayerScopes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		want   TriggerPlayerSelectorKind
	}{
		{"Whenever you complete a dungeon, draw a card.", TriggerPlayerSelectorYou},
		{"Whenever an opponent completes a dungeon, draw a card.", TriggerPlayerSelectorOpponent},
		{"Whenever a player completes a dungeon, draw a card.", TriggerPlayerSelectorAny},
	}
	for _, test := range tests {
		document, diagnostics := Parse(test.source, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("%q diagnostics = %#v", test.source, diagnostics)
		}
		abilities := document.Abilities
		if len(abilities) != 1 || abilities[0].Trigger == nil ||
			abilities[0].Trigger.TriggerEvent == nil {
			t.Fatalf("%q abilities = %#v", test.source, abilities)
		}
		event := abilities[0].Trigger.TriggerEvent
		if event.Kind != TriggerEventKindCompletedDungeon || event.Player.Kind != test.want {
			t.Fatalf("%q event = %#v, want player %v", test.source, event, test.want)
		}
	}
}
