package parser

import (
	"slices"
	"testing"
)

// quartzwoodOracle is Quartzwood Crasher's full rules text: a trample keyword
// plus the reusable "Whenever one or more creatures you control with trample
// deal combat damage to a player" batch trigger that creates an X/X token sized
// by the damage those creatures dealt to that player.
const quartzwoodOracle = "Trample\n" +
	"Whenever one or more creatures you control with trample deal combat damage to a player, " +
	"create an X/X green Dinosaur Beast creature token with trample, " +
	"where X is the amount of damage those creatures dealt to that player."

// TestParseQuartzwoodCombatDamageBatchTrigger proves the parser recognizes the
// combat-damage batch trigger with a trample-restricted controlled source, the
// per-player "one or more" framing, and the "the amount of damage those
// creatures dealt to that player" dynamic token size.
func TestParseQuartzwoodCombatDamageBatchTrigger(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(quartzwoodOracle, Context{CardName: "Quartzwood Crasher"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	var triggered *Ability
	for i := range document.Abilities {
		if document.Abilities[i].Kind == AbilityTriggered {
			triggered = &document.Abilities[i]
			break
		}
	}
	if triggered == nil {
		t.Fatal("no triggered ability parsed")
	}
	if triggered.Trigger == nil || triggered.Trigger.TriggerEvent == nil {
		t.Fatal("triggered ability has no trigger event")
	}

	event := triggered.Trigger.TriggerEvent
	if event.Kind != TriggerEventKindDamageDealt {
		t.Errorf("event kind = %q, want %q", event.Kind, TriggerEventKindDamageDealt)
	}
	if !event.OneOrMore {
		t.Error("OneOrMore = false, want true")
	}
	if event.CombatQualifier.Kind != TriggerEventCombatQualifierCombat {
		t.Errorf("combat qualifier = %q, want %q", event.CombatQualifier.Kind, TriggerEventCombatQualifierCombat)
	}
	if event.DamageRecipient.Kind != TriggerEventDamageRecipientPlayer {
		t.Errorf("damage recipient = %v, want player", event.DamageRecipient.Kind)
	}
	if event.DamageRecipient.IsSource {
		t.Error("damage recipient IsSource = true, want false (an independent player)")
	}
	if event.DamageSource.Kind != TriggerEventSubjectSelection {
		t.Fatalf("damage source kind = %q, want %q", event.DamageSource.Kind, TriggerEventSubjectSelection)
	}
	if !slices.Contains(event.DamageSource.Selection.RequiredTypes, TriggerCardTypeCreature) {
		t.Errorf("damage source types = %v, want to contain creature", event.DamageSource.Selection.RequiredTypes)
	}
	if event.DamageSource.Selection.Keyword != KeywordTrample {
		t.Errorf("damage source keyword = %q, want %q", event.DamageSource.Selection.Keyword, KeywordTrample)
	}

	effect := quartzwoodCreateEffect(t, triggered)
	if effect.Amount.DynamicKind != EffectDynamicAmountTriggeringEventTotalCombatDamage {
		t.Errorf("amount dynamic kind = %q, want %q", effect.Amount.DynamicKind, EffectDynamicAmountTriggeringEventTotalCombatDamage)
	}
	if effect.Amount.DynamicForm != EffectDynamicAmountFormWhereX {
		t.Errorf("amount dynamic form = %q, want %q", effect.Amount.DynamicForm, EffectDynamicAmountFormWhereX)
	}
	if !effect.TokenPTVariableX {
		t.Error("TokenPTVariableX = false, want true (X/X token)")
	}
	if !slices.Contains(effect.TokenKeywords, KeywordTrample) {
		t.Errorf("token keywords = %v, want to contain trample", effect.TokenKeywords)
	}
}

// quartzwoodCreateEffect returns the sole EffectCreate in the triggered
// ability's content.
func quartzwoodCreateEffect(t *testing.T, ability *Ability) EffectSyntax {
	t.Helper()
	for _, sentence := range ability.Sentences {
		for _, effect := range sentence.Effects {
			if effect.Kind == EffectCreate {
				return effect
			}
		}
	}
	t.Fatal("no EffectCreate parsed in triggered ability")
	return EffectSyntax{}
}
