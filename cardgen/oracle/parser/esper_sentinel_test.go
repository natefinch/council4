package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
)

func TestParseEsperSentinelTypedSyntax(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever an opponent casts their first noncreature spell each turn, draw a card unless that player pays {X}, where X is this creature's power.",
		Context{CardName: "Esper Sentinel"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.TriggerEvent == nil {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	event := ability.Trigger.TriggerEvent
	if event.Kind != TriggerEventKindSpellCast ||
		event.Actor.Kind != TriggerEventActorOpponent ||
		event.SpellSelection.Ordinal != 1 ||
		!slices.Equal(event.SpellSelection.ExcludedTypes, []TriggerCardType{TriggerCardTypeCreature}) {
		t.Fatalf("event clause = %#v", event)
	}
	effect := ability.Sentences[0].Effects[0]
	amount := effect.Payment.GenericManaAmount
	if effect.Payment.Payer != EffectPaymentPayerEventPlayer ||
		!slices.Equal(effect.Payment.ManaCost, cost.Mana{cost.X}) ||
		amount.DynamicKind != EffectDynamicAmountSourcePower ||
		amount.DynamicForm != EffectDynamicAmountFormWhereX ||
		amount.Multiplier != 1 {
		t.Fatalf("payment = %#v", effect.Payment)
	}
}
