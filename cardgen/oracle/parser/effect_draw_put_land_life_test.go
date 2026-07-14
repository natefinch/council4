package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseDrawPutLandSubtypeLifeSequence(t *testing.T) {
	t.Parallel()
	const source = "When this enchantment enters, draw a card, then you may put a land card from your hand onto the battlefield. If you put a Cave onto the battlefield this way, you gain 4 life."
	document, diagnostics := Parse(source, Context{CardName: "Spelunking"})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("document = %#v, diagnostics = %#v", document, diagnostics)
	}
	sequence := document.Abilities[0].ExactSequence
	if sequence == nil ||
		sequence.Kind != ExactSequenceDrawPutLandSubtypeLife ||
		sequence.PutSubtype != types.Cave ||
		sequence.LifeAmount != 4 {
		t.Fatalf("sequence = %#v", sequence)
	}
}

func TestParseGroupEntersUntappedEffect(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Lands you control enter untapped.", Context{})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("document = %#v, diagnostics = %#v", document, diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 ||
		!effects[0].Exact ||
		!effects[0].EntersUntappedGroup() ||
		effects[0].GroupEntryModification.ControllerScope != EntersTappedGroupControllerYou {
		t.Fatalf("effects = %#v", effects)
	}
}
