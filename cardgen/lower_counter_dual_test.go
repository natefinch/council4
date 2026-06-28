package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerDualReferencedCounterPlacement proves the enters-trigger sibling
// family that places one +1/+1 counter on the entering creature and one on the
// source lowers to two AddCounter instructions, one per referenced permanent in
// source order (Juniper Order Ranger, X-23, Deadly Weapon).
func TestLowerDualReferencedCounterPlacement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		oracleText string
	}{
		{
			name:       "Juniper Order Ranger",
			typeLine:   "Creature — Human Knight",
			oracleText: "Whenever another creature you control enters, put a +1/+1 counter on that creature and a +1/+1 counter on this creature.",
		},
		{
			name:       "X-23, Deadly Weapon",
			typeLine:   "Legendary Creature — Mutant Assassin",
			oracleText: "Whenever another Mutant you control enters, put a +1/+1 counter on that creature and a +1/+1 counter on X-23.",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			if len(mode.Sequence) != 2 {
				t.Fatalf("sequence = %#v, want two AddCounter instructions", mode.Sequence)
			}
			first, ok := mode.Sequence[0].Primitive.(game.AddCounter)
			if !ok || first.Object != game.EventPermanentReference() {
				t.Fatalf("first instruction = %#v, want AddCounter on event permanent", mode.Sequence[0].Primitive)
			}
			second, ok := mode.Sequence[1].Primitive.(game.AddCounter)
			if !ok || second.Object != game.SourcePermanentReference() {
				t.Fatalf("second instruction = %#v, want AddCounter on source", mode.Sequence[1].Primitive)
			}
		})
	}
}
