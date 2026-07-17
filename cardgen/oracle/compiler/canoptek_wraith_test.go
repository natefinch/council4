package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestCompileCanoptekWraithBindsChoiceAndSourcePayment(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Wraith Form — This creature can't be blocked.\nTransdimensional Scout — When this creature deals combat damage to a player, you may pay {3} and sacrifice it. If you do, choose a land you control. Then search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle.",
		pipelineContext{CardName: "Canoptek Wraith"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := compilation.Abilities[1]
	if ability.Trigger == nil ||
		ability.Trigger.Pattern.Event != TriggerEventDamageDealt ||
		ability.Trigger.Pattern.Source != TriggerSourceSelf ||
		ability.Trigger.Pattern.CombatQualifier != TriggerCombatDamage {
		t.Fatalf("trigger = %#v", ability.Trigger)
	}
	effects := ability.Content.Effects
	if len(effects) != 5 || effects[1].Kind != EffectChoosePermanent || effects[2].Kind != EffectSearch {
		t.Fatalf("effects = %#v", effects)
	}
	if !effects[2].SearchNameFromChosenPermanent {
		t.Fatalf("search did not bind to typed permanent choice: %#v", effects[2])
	}
	components := effects[1].Payment.AdditionalCost.Components
	if len(components) != 1 || components[0].Kind != CostSacrifice || !components[0].SourceSelf {
		t.Fatalf("payment components = %#v, want source sacrifice", components)
	}
}

func TestChosenPermanentSearchBindingIsTextAndPositionBlind(t *testing.T) {
	t.Parallel()
	effects := []CompiledEffect{
		{Kind: EffectChoosePermanent, Text: "unrelated"},
		{
			Kind:                         EffectSearch,
			Text:                         "not oracle text",
			SearchSameNameAsChosenObject: true,
			References: []CompiledReference{{
				Kind:             ReferenceThatObject,
				Binding:          ReferenceBindingPriorInstructionResult,
				PriorInstruction: 0,
			}},
		},
	}

	resolveChosenPermanentSearchNames(effects)
	if !effects[1].SearchNameFromChosenPermanent || effects[1].UnsupportedDetail != "" {
		t.Fatalf("bound = %v, unsupported = %q", effects[1].SearchNameFromChosenPermanent, effects[1].UnsupportedDetail)
	}

	effects[0].Kind = EffectDraw
	effects[1].SearchNameFromChosenPermanent = false
	resolveChosenPermanentSearchNames(effects)
	if effects[1].UnsupportedDetail == "" {
		t.Fatal("non-choice antecedent was accepted")
	}
}

func TestPaymentSourceSacrificeBindingRejectsAmbiguousComposite(t *testing.T) {
	t.Parallel()
	effects := []CompiledEffect{
		{
			Kind:      EffectSacrifice,
			VerbOrder: shared.SourceOrder{Start: 2, End: 3},
			References: []CompiledReference{{
				Binding: ReferenceBindingEventPermanent,
			}},
		},
		{
			Payment: CompiledEffectPayment{
				Order: shared.SourceOrder{Start: 1, End: 4},
				AdditionalCost: &CompiledCost{Components: []CostComponent{
					{Kind: CostSacrifice},
					{Kind: CostSacrifice},
				}},
			},
		},
	}
	resolvePaymentSourceSacrifices(effects, true)
	for i, component := range effects[1].Payment.AdditionalCost.Components {
		if component.SourceSelf {
			t.Fatalf("ambiguous sacrifice component %d bound to source", i)
		}
	}
}
