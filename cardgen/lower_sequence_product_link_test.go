package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func TestSequenceProductLinkRequiresExactPublisher(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name        string
		instruction game.Instruction
		references  []int
		ranges      [][2]int
		ok          bool
	}{
		{"card return", game.Instruction{Primitive: game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
		}}, []int{0}, [][2]int{{0, 1}}, true},
		{"single look", game.Instruction{Primitive: game.LookAtLibraryTop{}}, []int{0}, [][2]int{{0, 1}}, true},
		{"optional look", game.Instruction{Primitive: game.LookAtLibraryTop{}, Optional: true}, []int{0}, [][2]int{{0, 1}}, true},
		{"single reveal", game.Instruction{Primitive: game.Reveal{Amount: game.Fixed(1)}}, []int{0}, [][2]int{{0, 1}}, true},
		{"optional reveal", game.Instruction{Primitive: game.Reveal{Amount: game.Fixed(1)}, Optional: true}, []int{0}, [][2]int{{0, 1}}, true},
		{"plural reveal", game.Instruction{Primitive: game.Reveal{Amount: game.Fixed(2)}}, []int{0}, [][2]int{{0, 1}}, false},
		{"dynamic reveal", game.Instruction{Primitive: game.Reveal{Amount: game.Dynamic(game.DynamicAmount{})}}, []int{0}, [][2]int{{0, 1}}, false},
		{"blink return", game.Instruction{Primitive: game.PutOnBattlefield{
			Source: game.LinkedBattlefieldSource("departed"),
		}}, []int{0, 0}, [][2]int{{0, 1}}, true},
		{"matching publication", game.Instruction{Primitive: game.PutOnBattlefield{
			Source: game.LinkedBattlefieldSource("departed"), PublishLinked: sequenceProductKey(0),
		}}, []int{0}, [][2]int{{0, 1}}, true},
		{"competing identities", game.Instruction{Primitive: game.CreateToken{}}, []int{0, 1}, [][2]int{{0, 1}, {0, 1}}, false},
		{"conflicting publication", game.Instruction{Primitive: game.CreateToken{PublishLinked: "existing"}}, []int{0}, [][2]int{{0, 1}}, false},
		{"optional", game.Instruction{Primitive: game.CreateToken{}, Optional: true}, []int{0}, [][2]int{{0, 1}}, false},
		{"plural return", game.Instruction{Primitive: game.PutOnBattlefield{
			Sources: []game.BattlefieldSource{game.LinkedBattlefieldSource("group")},
		}}, []int{0}, [][2]int{{0, 1}}, false},
		{"multiple instructions", game.Instruction{Primitive: game.CreateToken{}}, []int{0}, [][2]int{{0, 2}}, false},
		{"future producer", game.Instruction{Primitive: game.CreateToken{}}, []int{1}, [][2]int{{0, 1}}, false},
		{"negative producer", game.Instruction{Primitive: game.CreateToken{}}, []int{-1}, [][2]int{{0, 1}}, false},
		{"invalid range", game.Instruction{Primitive: game.CreateToken{}}, []int{0}, [][2]int{{-1, 0}}, false},
		{"nonpublisher", game.Instruction{Primitive: game.GainLife{}}, []int{0}, [][2]int{{0, 1}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var references []compiler.CompiledReference
			for _, index := range test.references {
				references = append(references, compiler.CompiledReference{
					Binding: compiler.ReferenceBindingPriorInstructionResult, PriorInstruction: index,
				})
			}
			sequence := []game.Instruction{test.instruction}
			before := game.PublishedLinkedKey(sequence[0].Primitive)
			producer, key, ok := sequencePriorInstructionLink(references, sequence, test.ranges)
			if ok != test.ok {
				t.Fatalf("publisher accepted=%t, want %t", ok, test.ok)
			}
			if ok && (producer != 0 || key != sequenceProductKey(0) || game.PublishedLinkedKey(sequence[0].Primitive) != key) {
				t.Fatal("accepted producer did not share its exact canonical publication")
			}
			if !ok && game.PublishedLinkedKey(sequence[0].Primitive) != before {
				t.Fatal("failed linking mutated the publisher")
			}
		})
	}
}
