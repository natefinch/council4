package compiler

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestLootDisputeTriggerPatternsCompileCompositionally(t *testing.T) {
	t.Parallel()
	tests := []struct {
		event string
		want  TriggerPattern
	}{
		{
			event: "you attack the player who has the initiative",
			want: TriggerPattern{
				Kind:                     TriggerWhenever,
				Event:                    TriggerEventAttackerDeclared,
				Controller:               ControllerYou,
				Player:                   TriggerPlayerInitiative,
				AttackRecipient:          TriggerAttackRecipientPlayer,
				OneOrMore:                true,
				OneOrMorePerAttackTarget: true,
			},
		},
		{
			event: "you complete a dungeon",
			want: TriggerPattern{
				Kind:   TriggerWhenever,
				Event:  TriggerEventCompletedDungeon,
				Player: TriggerPlayerYou,
			},
		},
	}
	for _, test := range tests {
		got := compileParsedTriggerPattern(test.event, TriggerWhenever, shared.Span{}, "", nil)
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("%q pattern = %#v, want %#v", test.event, got, test.want)
		}
	}
}

func TestCompoundTakeInitiativeAndCreateCompilesInOrder(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When this enchantment enters, you take the initiative and create a Treasure token.",
		pipelineContext{CardName: "Example"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 2 ||
		effects[0].Kind != EffectTakeInitiative ||
		effects[1].Kind != EffectCreate {
		t.Fatalf("effects = %#v, want initiative then create", effects)
	}
}
