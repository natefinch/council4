package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerSelfAndTargetBounce proves the source-and-target battlefield bounce
// "Return this creature and (another) target creature to their owners' hands."
// (Wizard Mentor, Coastal Wizard, Snow Hound, Lady Sun) lowers to one
// single-target spec and two Bounce instructions: the source first, then the
// chosen target. It is the self sibling of the dual-target bounce, where one of
// the two returned permanents is the ability's own source rather than a second
// target.
func TestLowerSelfAndTargetBounce(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
	}{
		{
			name:       "this creature and target creature you control",
			oracleText: "{T}: Return this creature and target creature you control to their owner's hand.",
		},
		{
			name:       "this creature and another target creature",
			oracleText: "{T}: Return this creature and another target creature to their owners' hands.",
		},
		{
			name:       "named source and another target creature",
			oracleText: "{T}: Return Test Self Bounce and another target creature to their owners' hands.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Self Bounce",
				Layout:     "normal",
				TypeLine:   "Creature — Wizard",
				OracleText: test.oracleText,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			mode := face.ActivatedAbilities[0].Content.Modes[0]
			if len(mode.Targets) != 1 {
				t.Fatalf("targets = %#v, want one spec", mode.Targets)
			}
			if len(mode.Sequence) != 2 {
				t.Fatalf("sequence len = %d, want 2", len(mode.Sequence))
			}
			source, ok := mode.Sequence[0].Primitive.(game.Bounce)
			if !ok || source.Object != game.SourcePermanentReference() {
				t.Fatalf("sequence[0] = %#v, want Bounce of source permanent", mode.Sequence[0].Primitive)
			}
			target, ok := mode.Sequence[1].Primitive.(game.Bounce)
			if !ok || target.Object != game.TargetPermanentReference(0) {
				t.Fatalf("sequence[1] = %#v, want Bounce of TargetPermanentReference(0)", mode.Sequence[1].Primitive)
			}
		})
	}
}

// TestLowerSelfAndTargetBounceFailClosed proves the self-and-target bounce path
// declines wordings outside its template so neighboring paths stay unchanged: a
// two-target bounce stays on the dual-target path, and a third permanent in the
// clause has no representable form.
func TestLowerSelfAndTargetBounceFailClosed(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Return this creature and target creature and target land to their owners' hands.",
	}
	for _, text := range tests {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Self Bounce Fail",
				Layout:     "normal",
				TypeLine:   "Creature — Wizard",
				OracleText: "{T}: " + text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
		})
	}
}
