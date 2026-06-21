package parser

import "testing"

// TestParseThresholdInsteadMana covers Cabal Ritual's second paragraph
// "Threshold — Add {B}{B}{B}{B}{B} instead if there are seven or more cards in
// your graveyard.": the add-mana body's trailing "instead" is recognized as the
// conditional-alternative marker, the five {B} symbols are captured, and the
// graveyard-size threshold is split off as the effect's condition.
func TestParseThresholdInsteadMana(t *testing.T) {
	t.Parallel()
	text := "Add {B}{B}{B}.\nThreshold — Add {B}{B}{B}{B}{B} instead if there are seven or more cards in your graveyard."
	document, diagnostics := Parse(text, Context{CardName: "Cabal Ritual", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 2 {
		t.Fatalf("abilities = %d, want 2", len(document.Abilities))
	}
	base := findAddManaEffect(t, &document.Abilities[0])
	if base.Mana.Instead {
		t.Fatalf("base production marked Instead: %#v", base.Mana)
	}
	if len(base.Mana.Symbols) != 3 {
		t.Fatalf("base symbols = %#v, want 3", base.Mana.Symbols)
	}
	alternative := findAddManaEffect(t, &document.Abilities[1])
	if !alternative.Mana.Instead {
		t.Fatalf("alternative not marked Instead: %#v", alternative.Mana)
	}
	if len(alternative.Mana.Symbols) != 5 {
		t.Fatalf("alternative symbols = %#v, want 5", alternative.Mana.Symbols)
	}
}

func findAddManaEffect(t *testing.T, ability *Ability) *EffectSyntax {
	t.Helper()
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			if ability.Sentences[i].Effects[j].Kind == EffectAddMana {
				return &ability.Sentences[i].Effects[j]
			}
		}
	}
	t.Fatalf("no add-mana effect: %#v", ability.Sentences)
	return nil
}
