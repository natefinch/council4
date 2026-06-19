package parser

import "testing"

func TestParseSentenceLeadingThenIfAsResolvingCondition(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(anchorOracle, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if len(ability.ConditionBoundaries) != 1 || !ability.ConditionBoundaries[0].Resolving {
		t.Fatalf("condition boundaries = %#v, want one resolving boundary", ability.ConditionBoundaries)
	}
	if len(ability.Sentences) != 2 ||
		len(ability.Sentences[1].Effects) != 1 ||
		ability.Sentences[1].Effects[0].Connection != EffectConnectionThen {
		t.Fatalf("second sentence effects = %#v, want then-connected untap", ability.Sentences[1].Effects)
	}
}

func TestParseOrdinaryIfIsNotResolvingCondition(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"{T}: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. If you control four or more lands, untap that land.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if document.Abilities[0].ConditionBoundaries[0].Resolving {
		t.Fatal("ordinary sentence-leading if classified as resolving Then if")
	}
}

const anchorOracle = "{T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. Then if you control four or more lands, untap that land."
