package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestParseTurnFaceDownThenBecomeCharacteristics(t *testing.T) {
	source := "Turn target creature face down. It's a 2/2 Cyberman artifact creature."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v diagnostics = %#v", document.Abilities, diagnostics)
	}

	ability := document.Abilities[0]
	if len(ability.Sentences) != 2 {
		t.Fatalf("sentences = %#v", ability.Sentences)
	}
	turn := ability.Sentences[0]
	if len(turn.Targets) != 1 ||
		turn.Targets[0].Selection.Kind != SelectionCreature ||
		turn.Targets[0].Text != "target creature" ||
		len(turn.Effects) != 1 ||
		turn.Effects[0].Kind != EffectTurnFaceDown {
		t.Fatalf("turn sentence = %#v", turn)
	}
	become := ability.Sentences[1]
	if len(become.Effects) != 1 {
		t.Fatalf("become effects = %#v", become.Effects)
	}
	effect := become.Effects[0]
	if effect.Kind != EffectPolymorph ||
		effect.Context != EffectContextReferencedObject ||
		!effect.PolymorphPermanent ||
		effect.PolymorphLosesAllAbilities ||
		effect.PolymorphBasePower != 2 ||
		effect.PolymorphBaseToughness != 2 ||
		!slices.Equal(effect.PolymorphTypes, []types.Card{types.Artifact, types.Creature}) ||
		!slices.Equal(effect.PolymorphSubtypes, []types.Sub{types.Cyberman}) {
		t.Fatalf("become effect = %#v", effect)
	}
}

func TestSentenceInitialItsIsObjectPronoun(t *testing.T) {
	document, diagnostics := Parse(
		"Turn target creature face down. It's a 2/2 Cyberman artifact creature.",
		Context{CardName: "Cyber Conversion", InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	references := document.Abilities[0].SemanticReferences
	if len(references) != 1 ||
		references[0].Kind != ReferencePronoun ||
		references[0].Pronoun != PronounIt ||
		references[0].Text != "It's" {
		t.Fatalf("semantic references = %#v", references)
	}
}

func TestParseTurnFaceDownCompositionIsGeneric(t *testing.T) {
	source := "Turn target artifact face down. It becomes a 3/3 Golem artifact creature."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 ||
		len(document.Abilities[0].Sentences) != 2 {
		t.Fatalf("abilities = %#v diagnostics = %#v", document.Abilities, diagnostics)
	}
	turn := document.Abilities[0].Sentences[0]
	become := document.Abilities[0].Sentences[1].Effects[0]
	if turn.Targets[0].Selection.Kind != SelectionArtifact ||
		become.Context != EffectContextReferencedObject ||
		become.PolymorphBasePower != 3 ||
		become.PolymorphBaseToughness != 3 ||
		!slices.Equal(become.PolymorphSubtypes, []types.Sub{types.Golem}) {
		t.Fatalf("turn = %#v become = %#v", turn, become)
	}
}
