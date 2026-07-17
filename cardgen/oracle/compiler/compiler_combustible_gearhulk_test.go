package compiler

import "testing"

func TestCompileReferencedCardsTotalManaValue(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"When this creature enters, target opponent may have you draw three cards. If the player doesn't, you mill three cards, then this creature deals damage to that player equal to the total mana value of those cards.",
		pipelineContext{CardName: "Combustible Gearhulk"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}

	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 4 {
		t.Fatalf("effects = %#v", effects)
	}
	damage := effects[3]
	if damage.Amount.DynamicKind != DynamicAmountReferencedCardsTotalManaValue {
		t.Fatalf("damage amount = %#v", damage.Amount)
	}
	var playerTarget, linkedBatch bool
	for _, reference := range damage.References {
		switch reference.NodeID {
		case damage.Amount.ReferenceNodeID:
			linkedBatch = reference.Binding == ReferenceBindingPriorInstructionResult &&
				reference.PriorInstruction == 2
		default:
			if reference.Kind == ReferenceThatPlayer {
				playerTarget = reference.Binding == ReferenceBindingTarget && reference.Occurrence == 0
			}
		}
	}
	if !playerTarget || !linkedBatch {
		t.Fatalf("damage references = %#v", damage.References)
	}
}

func TestCompileMillThoseCardsRequiresSupportedConsumer(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Mill four cards. You may put a creature card from among those cards into your hand.",
		pipelineContext{CardName: "Test Card"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, effect := range compilation.Abilities[0].Content.Effects {
		for _, reference := range effect.References {
			if reference.Kind == ReferencePronoun && reference.Pronoun == ReferencePronounThose &&
				reference.Binding == ReferenceBindingPriorInstructionResult {
				t.Fatalf("unsupported mill consumer bound as a prior-instruction result: %#v", reference)
			}
		}
	}
}
