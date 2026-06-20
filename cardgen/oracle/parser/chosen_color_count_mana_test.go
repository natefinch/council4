package parser

import "testing"

// TestParseChosenColorCountMana covers Three Tree City's third ability "Choose a
// color. Add an amount of mana of that color equal to the number of creatures you
// control of the chosen type.": the add-mana body is recognized as the
// chosen-color dynamic-count form, its amount is a battlefield creature count
// restricted to the source's chosen creature type, the leading "Choose a color."
// sentence is credited onto the add-mana effect span, and no diagnostic is
// raised.
func TestParseChosenColorCountMana(t *testing.T) {
	t.Parallel()
	text := "{2}, {T}: Choose a color. Add an amount of mana of that color equal to the number of creatures you control of the chosen type."
	document, diagnostics := Parse(text, Context{CardName: "Three Tree City"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	ability := document.Abilities[0]
	var manaEffect *EffectSyntax
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			if ability.Sentences[i].Effects[j].Kind == EffectAddMana {
				manaEffect = &ability.Sentences[i].Effects[j]
			}
		}
	}
	if manaEffect == nil {
		t.Fatalf("no add-mana effect: %#v", ability.Sentences)
	}
	if !manaEffect.Mana.ChosenColorDynamic {
		t.Fatalf("ChosenColorDynamic not set: %#v", manaEffect.Mana)
	}
	if !manaEffect.Exact {
		t.Fatal("add-mana effect not exact")
	}
	if manaEffect.HasUnrecognizedSibling {
		t.Fatal("Choose a color sentence left the effect with an unrecognized sibling")
	}
	if manaEffect.Amount.DynamicKind != EffectDynamicAmountCount {
		t.Fatalf("amount kind = %v, want count", manaEffect.Amount.DynamicKind)
	}
	if manaEffect.Amount.Selection == nil || !manaEffect.Amount.Selection.SubtypeFromEntryChoice {
		t.Fatalf("amount selection SubtypeFromEntryChoice not set: %#v", manaEffect.Amount.Selection)
	}
	if manaEffect.Amount.Selection.Kind != SelectionCreature ||
		manaEffect.Amount.Selection.Controller != SelectionControllerYou {
		t.Fatalf("amount selection = %#v, want your creatures", manaEffect.Amount.Selection)
	}
	// The effect span must widen to cover the leading "Choose a color." sentence
	// so the mana ability's coverage scan credits the choice.
	if manaEffect.Span.Start.Offset > 12 {
		t.Fatalf("effect span start = %d, want it widened over \"Choose a color.\"", manaEffect.Span.Start.Offset)
	}
}

// TestParseChosenTypeDynamicCount covers the generic "of the chosen type"
// qualifier on a dynamic count subject outside the mana context, confirming the
// trailing qualifier extends a draw amount and sets the chosen-type predicate.
func TestParseChosenTypeDynamicCount(t *testing.T) {
	t.Parallel()
	text := "Draw cards equal to the number of creatures you control of the chosen type."
	document, diagnostics := Parse(text, Context{CardName: "Probe"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var drawEffect *EffectSyntax
	for i := range document.Abilities {
		for j := range document.Abilities[i].Sentences {
			for k := range document.Abilities[i].Sentences[j].Effects {
				if document.Abilities[i].Sentences[j].Effects[k].Kind == EffectDraw {
					drawEffect = &document.Abilities[i].Sentences[j].Effects[k]
				}
			}
		}
	}
	if drawEffect == nil {
		t.Fatalf("no draw effect: %#v", document.Abilities)
	}
	if !drawEffect.Exact {
		t.Fatal("draw effect not exact")
	}
	if drawEffect.Amount.Selection == nil || !drawEffect.Amount.Selection.SubtypeFromEntryChoice {
		t.Fatalf("draw amount selection SubtypeFromEntryChoice not set: %#v", drawEffect.Amount.Selection)
	}
}
