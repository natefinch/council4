package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// expressiveIterationParserText is the exact Expressive Iteration Oracle text the
// look-and-route dig recognizer targets.
const expressiveIterationParserText = "Look at the top three cards of your library. " +
	"Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. " +
	"You may play the exiled card this turn."

// TestParseDigRouteExpressiveIteration proves the recognizer folds the three
// Expressive Iteration sentences into one exact EffectDig carrying the typed
// look-and-route payload: a look of three fanned into ordered hand, library-
// bottom, and exile routes, with the exile route granting play this turn.
func TestParseDigRouteExpressiveIteration(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		expressiveIterationParserText,
		Context{CardName: "Expressive Iteration", InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("diagnostics = %#v, abilities = %#v", diagnostics, document.Abilities)
	}
	sentences := document.Abilities[0].Sentences
	if len(sentences) != 3 {
		t.Fatalf("sentences = %d, want 3", len(sentences))
	}
	// The recognizer consolidates all three sentences onto the first; the trailing
	// two carry no sibling effects.
	if len(sentences[0].Effects) != 1 {
		t.Fatalf("first sentence effects = %#v, want one consolidated effect", sentences[0].Effects)
	}
	if len(sentences[1].Effects) != 0 || len(sentences[2].Effects) != 0 {
		t.Fatalf("trailing sentences carried effects: %#v / %#v", sentences[1].Effects, sentences[2].Effects)
	}
	effect := sentences[0].Effects[0]
	if effect.Kind != EffectDig || !effect.Exact || !effect.DigRouteSequence {
		t.Fatalf("effect = %#v, want exact DigRouteSequence EffectDig", effect)
	}
	if effect.Context != EffectContextController {
		t.Fatalf("effect context = %v, want controller", effect.Context)
	}
	route := effect.DigRoute
	want := []DigRouteSlotSyntax{
		{Count: 1, Destination: zone.Hand},
		{Count: 1, Destination: zone.Library, Bottom: true},
		{Count: 1, Destination: zone.Exile, PlayThisTurn: true},
	}
	if route.Look != 3 || len(route.Slots) != len(want) {
		t.Fatalf("route = %#v, want look 3 with 3 slots", route)
	}
	for i, slot := range want {
		if route.Slots[i] != slot {
			t.Fatalf("slot[%d] = %#v, want %#v", i, route.Slots[i], slot)
		}
	}
}

// TestParseDigRouteFailsClosed proves the recognizer is exact: near misses to the
// three-way routing never produce a consolidated DigRouteSequence effect, so the
// unmodeled wording falls through to per-sentence handling and stays unsupported.
func TestParseDigRouteFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		text string
	}{
		{
			name: "look count not partitioned",
			text: "Look at the top four cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card this turn.",
		},
		{
			name: "non-unique routing count",
			text: "Look at the top four cards of your library. Put two of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card this turn.",
		},
		{
			name: "routes out of order",
			text: "Look at the top three cards of your library. Exile one of them, put one of them into your hand, and put one of them on the bottom of your library. You may play the exiled card this turn.",
		},
		{
			name: "graveyard destination",
			text: "Look at the top three cards of your library. Put one of them into your hand, put one of them into your graveyard, and exile one of them. You may play the exiled card this turn.",
		},
		{
			name: "library top destination",
			text: "Look at the top three cards of your library. Put one of them into your hand, put one of them on top of your library, and exile one of them. You may play the exiled card this turn.",
		},
		{
			name: "until end of turn duration",
			text: "Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card until end of turn.",
		},
		{
			name: "cast only permission",
			text: "Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may cast the exiled card this turn.",
		},
		{
			name: "without paying mana cost",
			text: "Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card this turn without paying its mana cost.",
		},
		{
			name: "two routes only",
			text: "Look at the top two cards of your library. Put one of them into your hand, and exile one of them. You may play the exiled card this turn.",
		},
		{
			name: "no play permission",
			text: "Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(tc.text, Context{CardName: "Near Miss", InstantOrSorcery: true})
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.DigRouteSequence {
							t.Fatalf("near miss produced a DigRouteSequence effect: %#v", effect)
						}
					}
				}
			}
		})
	}
}
