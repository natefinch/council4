package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestResultObjectSelectionValidation(t *testing.T) {
	for _, test := range []struct {
		name     string
		producer Primitive
		gate     InstructionResultGate
		valid    bool
	}{
		{"typed result", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{RequiredTypes: []types.Card{types.Land}})}, true},
		{"post-move cards", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{}), CardOnly: true}, true},
		{"departure is not post-move card", Destroy{Object: SourcePermanentReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{}), CardOnly: true}, false},
		{"filtered complement", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{RequiredTypes: []types.Card{types.Land}}), Negate: true}, true},
		{"unfiltered negation", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, Negate: true}, false},
		{"wildcard still requires member", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{})}, true},
		{"unsupported producer", GainLife{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{})}, false},
		{"unnamed result", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{})}, false},
		{"failure result", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriFalse, ObjectSelection: opt.Val(Selection{})}, false},
		{"contextual selection", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{Tapped: TriTrue})}, false},
		{"nested contextual selection", Discard{Amount: Fixed(1), Player: ControllerReference()}, InstructionResultGate{Key: "result", Succeeded: TriTrue, ObjectSelection: opt.Val(Selection{AnyOf: []Selection{{Tapped: TriTrue}}})}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			sequence := []Instruction{
				{Primitive: test.producer, PublishResult: "result"},
				{Primitive: GainLife{Amount: Fixed(1), Player: ControllerReference()}, ResultGate: opt.Val(test.gate)},
			}
			if err := ValidateInstructionSequence(sequence); (err == nil) != test.valid {
				t.Fatalf("validation = %v, want valid %t", err, test.valid)
			}
		})
	}
}

func TestStackObjectCloneIsolatesResultObjectSnapshots(t *testing.T) {
	original := &StackObject{
		ResolutionResults: map[string]InstructionResolutionResult{"result": {Accepted: true, Succeeded: true}},
		ResolutionResultObjects: map[string][]ObjectSnapshot{"result": {{
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{"Pirate"},
			EntryChoices: map[ChoiceKey]ResolutionChoiceResult{"entry": {Kind: ResolutionChoiceSubtype, Subtype: "Pirate"}},
			Counters:     counter.NewSet(),
		}}},
	}
	original.ResolutionResultObjects["result"][0].Counters.Add(counter.PlusOnePlusOne, 1)
	g := NewGame([NumPlayers]PlayerConfig{})
	g.Stack.Push(original)
	cloned, ok := g.Clone().Stack.Peek()
	if !ok {
		t.Fatal("clone lost stack object")
	}
	cloned.ResolutionResultObjects["result"][0].Types[0] = types.Land
	cloned.ResolutionResultObjects["result"][0].Subtypes[0] = "Elf"
	cloned.ResolutionResultObjects["result"][0].EntryChoices["entry"] = ResolutionChoiceResult{Kind: ResolutionChoiceSubtype, Subtype: "Elf"}
	cloned.ResolutionResultObjects["result"][0].Counters.Add(counter.PlusOnePlusOne, 1)
	cloned.ResolutionResults["result"] = InstructionResolutionResult{}
	snapshot := original.ResolutionResultObjects["result"][0]
	if snapshot.Types[0] != types.Creature || snapshot.Subtypes[0] != "Pirate" ||
		snapshot.EntryChoices["entry"].Subtype != "Pirate" || snapshot.Counters.Get(counter.PlusOnePlusOne) != 1 {
		t.Fatalf("clone mutated original snapshot: %#v", snapshot)
	}
	if original.ResolutionResults["result"] != (InstructionResolutionResult{Accepted: true, Succeeded: true}) {
		t.Fatal("clone mutated original comparable result")
	}
}
