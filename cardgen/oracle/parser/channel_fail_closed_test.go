package parser

import "testing"

func TestChannelExactTargetUnionRejectsNearMisses(t *testing.T) {
	t.Parallel()

	for _, text := range []string{
		"Destroy target artifact, creature, or nonbasic land an opponent controls.",
		"Destroy target artifact, enchantment, or nonbasic land you control.",
		"Destroy target artifact, enchantment, or nonbasic land an opponent controls with mana value 3 or less.",
	} {
		document, diagnostics := Parse(text, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", text, diagnostics)
		}
		targets := document.Abilities[0].Sentences[0].Targets
		if len(targets) == 1 && targets[0].Exact && len(targets[0].Selection.Alternatives) == 3 {
			t.Fatalf("near-miss target unexpectedly received exact qualified-union semantics: %q", text)
		}
	}
}

func TestBasicLandAndBasicLandTypeSearchRemainDistinct(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse(
		"Search your library for a basic land card, put it onto the battlefield, then shuffle.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	search := document.Abilities[0].Sentences[0].Effects[0]
	if search.Selection.BasicLandType {
		t.Fatal("basic land card search was conflated with land card with a basic land type")
	}
	if len(search.Selection.Supertypes) != 1 || search.Selection.Supertypes[0] != SupertypeBasic {
		t.Fatalf("search selection = %#v, want Basic supertype", search.Selection)
	}
}

func TestChannelOtherCostsDoNotBecomeDiscardSelfHandCosts(t *testing.T) {
	t.Parallel()

	for _, text := range []string{
		"Channel — {1}{G}, Discard a card: Draw a card.",
		"Channel — {1}{G}, Sacrifice this card: Draw a card.",
	} {
		document, diagnostics := Parse(text, Context{})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", text, diagnostics)
		}
		for _, component := range document.Abilities[0].CostSyntax.Components {
			if component.Kind == CostComponentDiscard && component.SourceSelf {
				t.Fatalf("near-miss cost unexpectedly became discard-self: %q", text)
			}
		}
	}
}

func TestSourceAbilityCostReductionRejectsColoredReduction(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse(
		"{1}{G}, Discard this card: Draw a card. This ability costs {G} less to activate for each legendary creature you control.",
		Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if document.Abilities[0].SourceAbilityCostReduction != nil {
		t.Fatalf("colored reduction unexpectedly recognized: %#v", document.Abilities[0].SourceAbilityCostReduction)
	}
}
