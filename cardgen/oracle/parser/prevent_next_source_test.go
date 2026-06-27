package parser

import (
	"reflect"
	"testing"
)

// TestParsePreventNextDamageFromSourceEffect covers the Circle of Protection /
// Rune of Protection wording "The next time a <color> source of your choice
// would deal damage to you this turn, prevent that damage." The recognizer
// captures the optional single source color and fails closed on the deferred
// variants (other recipients, partial prevention, of-the-chosen-color riders).
func TestParsePreventNextDamageFromSourceEffect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source     string
		recognized bool
		colors     []Color
	}{
		{"The next time a white source of your choice would deal damage to you this turn, prevent that damage.", true, []Color{ColorWhite}},
		{"The next time a black source of your choice would deal damage to you this turn, prevent that damage.", true, []Color{ColorBlack}},
		{"The next time a blue source of your choice would deal damage to you this turn, prevent that damage.", true, []Color{ColorBlue}},
		{"The next time a red source of your choice would deal damage to you this turn, prevent that damage.", true, []Color{ColorRed}},
		{"The next time a green source of your choice would deal damage to you this turn, prevent that damage.", true, []Color{ColorGreen}},
		{"The next time a source of your choice would deal damage to you this turn, prevent that damage.", true, nil},
		// Deferred variants must fail closed.
		{"The next time a source of your choice would deal damage to target creature this turn, prevent that damage.", false, nil},
		{"The next time a white source of your choice would deal damage to you this turn, prevent half that damage, rounded up.", false, nil},
		{"The next time a source of your choice would deal damage to any target this turn, prevent that damage.", false, nil},
		{"Prevent the next 3 damage that would be dealt to you this turn.", false, nil},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			got := len(effects) == 1 &&
				effects[0].Kind == EffectPreventDamage &&
				effects[0].PreventDamageNextFromSource
			if got != test.recognized {
				t.Fatalf("recognized = %v, want %v (effects=%#v)", got, test.recognized, effects)
			}
			if got && !reflect.DeepEqual(effects[0].PreventDamageSourceColors, test.colors) {
				t.Fatalf("PreventDamageSourceColors = %#v, want %#v", effects[0].PreventDamageSourceColors, test.colors)
			}
		})
	}
}
