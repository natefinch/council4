package parser

import (
	"slices"
	"testing"
)

func becomeColorEffect(t *testing.T, name, text string) (EffectSyntax, bool) {
	t.Helper()
	doc, _ := Parse(text, Context{CardName: name})
	for a := range doc.Abilities {
		ability := &doc.Abilities[a]
		for s := range ability.Sentences {
			for e := range ability.Sentences[s].Effects {
				effect := ability.Sentences[s].Effects[e]
				if effect.Kind == EffectBecomeColor {
					return effect, true
				}
			}
		}
	}
	return EffectSyntax{}, false
}

func TestParseBecomeColorSelfColorless(t *testing.T) {
	effect, ok := becomeColorEffect(t, "Blazing Blade Askari",
		"{1}{R}: This creature becomes colorless until end of turn.")
	if !ok {
		t.Fatal("no become-color effect parsed")
	}
	if !effect.BecomeColorSource {
		t.Error("expected source subject")
	}
	if !effect.BecomeColorColorless {
		t.Error("expected colorless")
	}
	if len(effect.BecomeColorColors) != 0 {
		t.Errorf("colors = %v, want none", effect.BecomeColorColors)
	}
	if !effect.BecomeColorUntilEndOfTurn {
		t.Error("expected until-end-of-turn duration")
	}
}

func TestParseBecomeColorTargetNamedColor(t *testing.T) {
	effect, ok := becomeColorEffect(t, "Fylamarid",
		"{T}: Target permanent becomes blue until end of turn.")
	if !ok {
		t.Fatal("no become-color effect parsed")
	}
	if effect.BecomeColorSource {
		t.Error("expected target subject, not source")
	}
	if effect.BecomeColorColorless {
		t.Error("did not expect colorless")
	}
	if !slices.Equal(effect.BecomeColorColors, []Color{ColorBlue}) {
		t.Errorf("colors = %v, want [ColorBlue]", effect.BecomeColorColors)
	}
}

func TestParseBecomeColorMultiColor(t *testing.T) {
	effect, ok := becomeColorEffect(t, "Test",
		"{1}: Target permanent becomes white and blue until end of turn.")
	if !ok {
		t.Fatal("no become-color effect parsed")
	}
	if !slices.Equal(effect.BecomeColorColors, []Color{ColorWhite, ColorBlue}) {
		t.Errorf("colors = %v, want [ColorWhite ColorBlue]", effect.BecomeColorColors)
	}
}

// TestParseBecomeColorFailsClosed asserts the recognizer rejects wordings it
// must not represent: the resolution-time "color of your choice" form, an
// additive "becomes a <color> <type>" form, and the form missing the required
// "until end of turn" duration.
func TestParseBecomeColorFailsClosed(t *testing.T) {
	for _, text := range []string{
		"{1}: Target creature becomes the color of your choice until end of turn.",
		"{1}: This creature becomes a blue artifact until end of turn.",
		"{1}: Target permanent becomes blue.",
	} {
		if _, ok := becomeColorEffect(t, "Test", text); ok {
			t.Errorf("expected no become-color effect for %q", text)
		}
	}
}
