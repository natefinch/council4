package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// expressiveIterationText is the exact Scryfall Oracle text of Expressive
// Iteration, the real card the look-and-route dig recognizer targets.
const expressiveIterationText = "Look at the top three cards of your library. " +
	"Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. " +
	"You may play the exiled card this turn."

// expressiveIterationDig is the exact Dig the pipeline must emit for Expressive
// Iteration; it mirrors the authoritative runtime target in
// mtg/rules/dig_slots_test.go: a hidden look at three cards whose primary Take
// is the hand route, with ordered library-bottom and exile slots (the exile
// slot granting play this turn).
func expressiveIterationDig() game.Dig {
	return game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(3),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderLibraryBottom,
		Slots: []game.DigSlot{
			{Count: game.Fixed(1), Destination: zone.Library, Bottom: true},
			{Count: game.Fixed(1), Destination: zone.Exile, Play: opt.Val(game.ImpulsePlayGrant{Duration: game.DurationThisTurn})},
		},
	}
}

// TestLowerDigRouteExpressiveIteration proves the real Expressive Iteration
// Oracle text lowers end-to-end (parser recognizer -> typed compiler payload ->
// lowering) to exactly one Dig with ordered slots and no leftover diagnostics.
func TestLowerDigRouteExpressiveIteration(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Expressive Iteration",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{U}{R}",
		OracleText: expressiveIterationText,
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %d, want 1", len(content.Modes))
	}
	mode := content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %+v, want no targets and one instruction", mode)
	}
	dig, ok := mode.Sequence[0].Primitive.(game.Dig)
	if !ok {
		t.Fatalf("primitive = %T, want game.Dig", mode.Sequence[0].Primitive)
	}
	if want := expressiveIterationDig(); !reflect.DeepEqual(dig, want) {
		t.Fatalf("dig =\n\t%#v\nwant\n\t%#v", dig, want)
	}
}

// TestLowerDigRouteFailsClosed proves the recognizer is exact: every near miss
// that the Dig-with-slots shape cannot model faithfully leaves the card fully
// unsupported (diagnostics, no partial spell ability) rather than lowering to a
// wrong Dig.
func TestLowerDigRouteFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		oracle string
	}{
		{
			// Look count exceeds the routed cards, so the routes do not
			// partition the looked-at cards.
			name:   "look count not partitioned",
			oracle: "Look at the top four cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card this turn.",
		},
		{
			// A route moves more than one card: not the unique 1/1/1 routing.
			name:   "non-unique routing count",
			oracle: "Look at the top four cards of your library. Put two of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card this turn.",
		},
		{
			// Routes appear in a different printed order.
			name:   "routes out of order",
			oracle: "Look at the top three cards of your library. Exile one of them, put one of them into your hand, and put one of them on the bottom of your library. You may play the exiled card this turn.",
		},
		{
			// A route sends cards to a destination the recognizer does not model.
			name:   "graveyard destination",
			oracle: "Look at the top three cards of your library. Put one of them into your hand, put one of them into your graveyard, and exile one of them. You may play the exiled card this turn.",
		},
		{
			// Library route puts cards on top rather than the bottom.
			name:   "library top destination",
			oracle: "Look at the top three cards of your library. Put one of them into your hand, put one of them on top of your library, and exile one of them. You may play the exiled card this turn.",
		},
		{
			// Permission window is a different duration than this turn.
			name:   "until end of turn duration",
			oracle: "Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card until end of turn.",
		},
		{
			// Permission is cast-only rather than the modeled play permission.
			name:   "cast only permission",
			oracle: "Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may cast the exiled card this turn.",
		},
		{
			// Free-cast rider is not the modeled plain play grant.
			name:   "without paying mana cost",
			oracle: "Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them. You may play the exiled card this turn without paying its mana cost.",
		},
		{
			// Only two routes, not the three-way hand/bottom/exile shape.
			name:   "two routes only",
			oracle: "Look at the top two cards of your library. Put one of them into your hand, and exile one of them. You may play the exiled card this turn.",
		},
		{
			// No impulse permission sentence at all.
			name:   "no play permission",
			oracle: "Look at the top three cards of your library. Put one of them into your hand, put one of them on the bottom of your library, and exile one of them.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Near Miss " + tc.name,
				Layout:     "normal",
				TypeLine:   "Sorcery",
				ManaCost:   "{U}{R}",
				OracleText: tc.oracle,
			})
			if face.SpellAbility.Exists {
				t.Fatalf("near miss lowered a spell ability: %#v", face.SpellAbility.Val)
			}
		})
	}
}
