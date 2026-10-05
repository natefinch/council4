package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestPublishedReturnConditionCompositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		text     string
		producer int
		consumer int
	}{
		{"Defy Death", "Return target creature card from your graveyard to the battlefield. If it's an Angel, put two +1/+1 counters on it.", 0, 1},
		{"Fearsome Awakening", "Return target creature card from your graveyard to the battlefield. If it's a Dragon, put two +1/+1 counters on it.", 0, 1},
		{"Return Upon the Tide", "Return target creature card from your graveyard to the battlefield. If it's an Elf, create two 1/1 green Elf Warrior creature tokens.", 0, 1},
		{"Essence Flux", "Exile target creature you control, then return that card to the battlefield under its owner's control. If it's a Spirit, put a +1/+1 counter on it.", 1, 2},
		{"Splash Portal", "Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, Frog, Otter, or Rat, draw a card.", 1, 2},
		{"nonadjacent return", "Return target creature card from your graveyard to the battlefield. You gain 1 life. If it's an Elf, draw a card.", 0, 2},
		{"second target", "Tap target creature. Return target creature card from your graveyard to the battlefield. If it's an Elf, put a +1/+1 counter on it.", 1, 2},
		{"two returned subjects", "Return target creature card from your graveyard to the battlefield. If it's an Elf, draw a card. Return target creature card from your graveyard to the battlefield. If it's a Dragon, put a +1/+1 counter on it.", 2, 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{Name: test.name, Layout: "normal", TypeLine: "Instant", OracleText: test.text}
			face := lowerSingleFace(t, card)
			sequence := face.SpellAbility.Val.Modes[0].Sequence
			put, ok := sequence[test.producer].Primitive.(game.PutOnBattlefield)
			if !ok || put.PublishLinked == "" {
				t.Fatalf("producer = %#v, want published battlefield entry", sequence[test.producer])
			}
			gate := effectConditionMatch(t, sequence[test.consumer])
			want := game.LinkedObjectReference(string(put.PublishLinked))
			if gate.Object.Val != want {
				t.Fatalf("condition subject = %#v, want actual returned object %#v", gate.Object.Val, want)
			}
			if test.name == "Splash Portal" && len(gate.ObjectMatches.Val.SubtypesAny) != 4 {
				t.Fatalf("subtype alternatives = %v, want all four", gate.ObjectMatches.Val.SubtypesAny)
			}
			if test.name == "two returned subjects" {
				first, ok := sequence[0].Primitive.(game.PutOnBattlefield)
				if !ok {
					t.Fatal("first return did not lower to a battlefield entry")
				}
				if first.PublishLinked == put.PublishLinked ||
					effectConditionMatch(t, sequence[1]).Object.Val != game.LinkedObjectReference(string(first.PublishLinked)) {
					t.Fatal("independent returned subjects share an identity")
				}
			}
			if add, ok := sequence[test.consumer].Primitive.(game.AddCounter); ok && add.Object != want {
				t.Fatalf("counter subject = %#v, want same published incarnation %#v", add.Object, want)
			}
			if test.producer > 0 && sequence[test.producer-1].Primitive.Kind() == game.PrimitiveMovePermanent {
				if key := game.PublishedLinkedKey(sequence[test.producer-1].Primitive); key == put.PublishLinked {
					t.Fatal("departed and returned objects share one publication")
				}
			}
			assertCardPaths(t, card, "SpellAbility.Val.Modes[0].Sequence")
		})
	}
}

func TestPublishedReturnConditionRefusesUnavailableSubjects(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"You may return target creature card from your graveyard to the battlefield. If it's an Elf, draw a card.",
		"Return two target creature cards from your graveyard to the battlefield. If it's an Elf, draw a card.",
		"Return target creature card from your graveyard to the battlefield. Reveal the top card of your library. If it's an Elf, draw a card.",
		"Return target creature card from your graveyard to the battlefield. If that land is an Island, draw a card.",
		"Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, Mystery, Otter, or Rat, draw a card.",
		"Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Mystery, Frog, Otter, or Rat, draw a card.",
		"Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, Frog, Otter, or Mystery, draw a card.",
		"Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, Frog, Otter, and Rat, draw a card.",
		"Exile target creature you control, then return it to the battlefield under its owner's control. If that creature is a Bird, , Otter, or Rat, draw a card.",
		"Return target creature card from your graveyard to the battlefield. Exile target creature. If it's an Elf, draw a card.",
	} {
		card := &ScryfallCard{Name: "Unavailable Subject", Layout: "normal", TypeLine: "Instant", OracleText: text}
		_, diagnostics := lowerExecutableFaces(card)
		if len(diagnostics) == 0 {
			t.Fatalf("unavailable subject admitted: %s", text)
		}
	}
}
