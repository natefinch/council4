package parser

import "testing"

func TestArcanisShortNameReferences(t *testing.T) {
	document, _ := Parse(
		"{2}{U}{U}: Return Arcanis to its owner's hand.",
		Context{CardName: "Arcanis the Omnipotent", Legendary: true},
	)
	references := document.Abilities[0].SemanticReferences
	if len(references) != 2 ||
		references[0].Kind != ReferenceSelfName ||
		references[1].Pronoun != PronounIts {
		t.Fatalf("references = %#v, want short self-name and possessive", references)
	}
}
