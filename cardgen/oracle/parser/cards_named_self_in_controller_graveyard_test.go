package parser

import "testing"

// TestParseCardsNamedSelfInControllerGraveyardForEach verifies that the
// "for each card named <this card> in your graveyard" count subject types to the
// controller-scoped self-named graveyard kind, distinct from the all-graveyards
// "in each graveyard" wording.
func TestParseCardsNamedSelfInControllerGraveyardForEach(t *testing.T) {
	t.Parallel()
	const name = "Growth Cycle"
	document, diagnostics := Parse(
		"Target creature gets +2/+2 until end of turn for each card named Growth Cycle in your graveyard.",
		Context{CardName: name, InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v", effects)
	}
	effect := effects[0]
	if !effect.Exact ||
		effect.Amount.DynamicKind != EffectDynamicAmountCardsNamedSelfInControllerGraveyard ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach {
		t.Fatalf("amount = %#v", effect.Amount)
	}
}

// TestParseCardsNamedSelfInEachVersusYourGraveyard verifies that the two
// graveyard-scope wordings type to distinct kinds: "in each graveyard" counts
// every graveyard, "in your graveyard" only the controller's.
func TestParseCardsNamedSelfInEachVersusYourGraveyard(t *testing.T) {
	t.Parallel()
	const name = "Growth Cycle"
	cases := []struct {
		text string
		want EffectDynamicAmountKind
	}{
		{
			text: "Target creature gets +2/+2 until end of turn for each card named Growth Cycle in each graveyard.",
			want: EffectDynamicAmountCardsNamedSelfInGraveyards,
		},
		{
			text: "Target creature gets +2/+2 until end of turn for each card named Growth Cycle in your graveyard.",
			want: EffectDynamicAmountCardsNamedSelfInControllerGraveyard,
		},
	}
	for _, tc := range cases {
		document, diagnostics := Parse(tc.text, Context{CardName: name, InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatalf("Parse(%q) diagnostics = %#v", tc.text, diagnostics)
		}
		effect := document.Abilities[0].Sentences[0].Effects[0]
		if effect.Amount.DynamicKind != tc.want {
			t.Fatalf("Parse(%q) kind = %q, want %q", tc.text, effect.Amount.DynamicKind, tc.want)
		}
	}
}

// TestParseCardsNamedSelfInControllerGraveyardForeignNameFailsClosed verifies
// that a foreign card name in the "in your graveyard" count subject is not
// recognized as a self-named graveyard count.
func TestParseCardsNamedSelfInControllerGraveyardForeignNameFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Target creature gets +2/+2 until end of turn for each card named Lightning Bolt in your graveyard.",
		Context{CardName: "Growth Cycle", InstantOrSorcery: true},
	)
	if len(document.Abilities) == 1 && len(document.Abilities[0].Sentences) == 1 {
		effect := document.Abilities[0].Sentences[0].Effects[0]
		if effect.Amount.DynamicKind == EffectDynamicAmountCardsNamedSelfInControllerGraveyard {
			t.Fatalf("foreign name typed to controller graveyard count: %#v", effect.Amount)
		}
	}
}
