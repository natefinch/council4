package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

const commandeerOracleText = "You may exile two blue cards from your hand rather than pay this spell's mana cost.\n" +
	"Gain control of target noncreature spell. You may choose new targets for it. " +
	"(If that spell is an artifact, enchantment, or planeswalker, the permanent enters under your control.)"

func TestParseCommandeerMechanics(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(commandeerOracleText, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want alternative cost and spell body", len(document.Abilities))
	}
	alternative := document.Abilities[0]
	if alternative.AlternativeCost == nil || alternative.AlternativeCost.Kind != SpellAlternativeCostPitch {
		t.Fatalf("alternative cost = %#v, want pitch", alternative.AlternativeCost)
	}
	if alternative.CostSyntax == nil || len(alternative.CostSyntax.Components) != 1 {
		t.Fatalf("pitch cost = %#v, want one exile component", alternative.CostSyntax)
	}
	exile := alternative.CostSyntax.Components[0]
	if exile.Kind != CostComponentExile ||
		exile.AmountValue != 2 ||
		!exile.AmountKnown ||
		exile.ObjectColor != ColorBlue ||
		!exile.ObjectColorKnown ||
		exile.SourceZone != zone.Hand {
		t.Fatalf("pitch exile = %#v, want exactly two blue cards from hand", exile)
	}

	body := document.Abilities[1]
	if len(body.Sentences) < 2 {
		t.Fatalf("sentences = %d, want gain-control and retarget sentences", len(body.Sentences))
	}
	if len(body.Sentences[0].Targets) != 1 || len(body.Sentences[0].Effects) != 1 {
		t.Fatalf("first sentence = %#v, want one target and one effect", body.Sentences[0])
	}
	if body.Sentences[0].Effects[0].Kind != EffectGainControl || !body.Sentences[0].Effects[0].Exact {
		t.Fatalf("first effect = %#v, want exact gain control", body.Sentences[0].Effects[0])
	}
	if len(body.Sentences[1].Effects) != 1 {
		t.Fatalf("second sentence = %#v, want one effect", body.Sentences[1])
	}
	retarget := body.Sentences[1].Effects[0]
	if retarget.Kind != EffectChooseNewTargets || !retarget.Exact || !retarget.Optional {
		t.Fatalf("second effect = %#v, want exact optional retarget", retarget)
	}
	if len(retarget.References) != 1 ||
		retarget.References[0].Kind != ReferencePronoun ||
		retarget.References[0].Pronoun != PronounIt {
		t.Fatalf("retarget references = %#v, want pronoun it", retarget.References)
	}
}
