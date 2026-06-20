package parser

import "testing"

// TestParseChosenColorDevotionMana covers Nykthos, Shrine to Nyx's second
// ability "Choose a color. Add an amount of mana of that color equal to your
// devotion to that color.": the add-mana body is recognized as the chosen-color
// devotion form, the leading "Choose a color." sentence is credited onto the
// add-mana effect span, and no diagnostic is raised.
func TestParseChosenColorDevotionMana(t *testing.T) {
	t.Parallel()
	text := "{2}, {T}: Choose a color. Add an amount of mana of that color equal to your devotion to that color. " +
		"(Your devotion to a color is the number of mana symbols of that color in the mana costs of permanents you control.)"
	document, diagnostics := Parse(text, Context{CardName: "Nykthos, Shrine to Nyx"})
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
	if !manaEffect.Mana.ChosenColorDevotion {
		t.Fatalf("ChosenColorDevotion not set: %#v", manaEffect.Mana)
	}
	if !manaEffect.Exact {
		t.Fatal("add-mana effect not exact")
	}
	if manaEffect.HasUnrecognizedSibling {
		t.Fatal("Choose a color sentence left the effect with an unrecognized sibling")
	}
	// The effect span must widen to cover the leading "Choose a color." sentence
	// so the mana ability's coverage scan credits the choice.
	if manaEffect.Span.Start.Offset > 12 {
		t.Fatalf("effect span start = %d, want it widened over \"Choose a color.\"", manaEffect.Span.Start.Offset)
	}
}
