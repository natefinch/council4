package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCapturePublicationProvesExpandedActualProducer(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		sequence []game.Instruction
		ok       bool
		index    int
	}{
		{"expanded optional token", []game.Instruction{{Primitive: game.GainLife{}}, {Primitive: game.CreateToken{}, Optional: true}, {Primitive: game.Draw{}}}, true, 1},
		{"expanded return", []game.Instruction{{Primitive: game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget})}}, {Primitive: game.AddCounter{}}}, true, 0},
		{"exiled card", []game.Instruction{{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(1), Destination: zone.Exile}}}, true, 0},
		{"top exiled card", []game.Instruction{{Primitive: game.MoveTopOfLibrary{Player: game.ControllerReference(), Amount: game.Fixed(1), Destination: zone.Exile}}}, true, 0},
		{"competing products", []game.Instruction{{Primitive: game.CreateToken{}}, {Primitive: game.CreateToken{}}}, false, 0},
		{"opaque expansion", []game.Instruction{{Primitive: game.CreateToken{}}, {Primitive: game.ChooseFromZone{}}}, false, 0},
		{"wrong destination", []game.Instruction{{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Hand}}}, false, 0},
		{"stack object", []game.Instruction{{Primitive: game.MovePermanent{Object: game.EventStackObjectReference(), Destination: zone.Exile}}}, false, 0},
		{"persistent link", []game.Instruction{{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinked: "persistent"}}}, false, 0},
		{"persistent canonical link", []game.Instruction{{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Exile, PublishLinked: sequenceProductKey(0)}}}, false, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			references := []compiler.CompiledReference{{Binding: compiler.ReferenceBindingPriorInstructionResult, PriorInstruction: 0}}
			before := make([]game.LinkedKey, len(test.sequence))
			for i := range test.sequence {
				before[i] = game.PublishedLinkedKey(test.sequence[i].Primitive)
			}
			producer, key, ok := sequencePriorInstructionPublication(references, test.sequence, [][2]int{{0, len(test.sequence)}}, true)
			if ok != test.ok {
				t.Fatalf("accepted=%t, want %t", ok, test.ok)
			}
			if ok {
				if producer != 0 || key != sequenceProductKey(0) || game.PublishedLinkedKey(test.sequence[test.index].Primitive) != key {
					t.Fatal("expanded producer did not publish its exact actual result")
				}
				if captureMovePublisher(test.sequence[test.index].Primitive) && !test.sequence[test.index].ClearLinkedBeforeGate {
					t.Fatal("new transient move publication did not opt into pre-gate clearing")
				}
			} else {
				for i := range test.sequence {
					if game.PublishedLinkedKey(test.sequence[i].Primitive) != before[i] || test.sequence[i].ClearLinkedBeforeGate {
						t.Fatal("failed publication proof mutated an instruction")
					}
				}
			}
		})
	}
}
