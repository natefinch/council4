package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPaidCostReferenceValidation(t *testing.T) {
	for _, test := range []struct {
		name  string
		ref   ObjectReference
		valid bool
	}{
		{"sacrifice noun", PaidCostReference("component", PaidCostSacrifice, types.Creature), true},
		{"any permanent", PaidCostReference("component", PaidCostSacrifice, ""), true},
		{"discard card", PaidCostReference("component", PaidCostDiscard, ""), true},
		{"missing component", PaidCostReference("", PaidCostDiscard, ""), false},
		{"unknown cost", PaidCostReference("component", PaidCostUnknown, ""), false},
		{"discard permanent noun", PaidCostReference("component", PaidCostDiscard, types.Creature), false},
		{"nonpermanent noun", PaidCostReference("component", PaidCostSacrifice, types.Instant), false},
		{"target and cost mixed", ObjectReference{kind: ObjectReferencePaidCost, targetIndex: 1, costKey: "component", costKind: PaidCostSacrifice}, false},
		{"source and cost mixed", ObjectReference{kind: ObjectReferenceSourcePermanent, costKey: "component"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := len(test.ref.Validate()) == 0; got != test.valid {
				t.Fatalf("valid=%v, want %v: %v", got, test.valid, test.ref.Validate())
			}
		})
	}
}

func TestPaidCostFactsDeepCloneAndStackCopies(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	obj := &StackObject{ID: 1, Kind: StackActivatedAbility, PaidCostSubjects: []PaidCostSubject{{
		Key: "component", Kind: PaidCostDiscard, CardZoneVersion: opt.Val(uint64(0)), CharacteristicsKnown: true,
		Snapshot: ObjectSnapshot{CardID: 4, Types: []types.Card{types.Creature}, Colors: []color.Color{color.Red},
			Subtypes: []types.Sub{types.Human}, EntryChoices: map[ChoiceKey]ResolutionChoiceResult{"type": {Subtype: types.Human}}},
	}}}
	g.Stack.Push(obj)
	clonedGame := g.Clone()
	cloned, _ := clonedGame.Stack.Peek()
	copied := NewStackObjectCopy(obj, 2)
	obj.PaidCostSubjects[0].Snapshot.Colors[0] = color.Blue
	obj.PaidCostSubjects[0].Snapshot.Subtypes[0] = types.Bear
	obj.PaidCostSubjects[0].Snapshot.EntryChoices["type"] = ResolutionChoiceResult{Subtype: types.Bear}
	obj.PaidCostSubjects[0].Key = "changed"
	for _, independent := range []*StackObject{cloned, copied} {
		fact := independent.PaidCostSubjects[0]
		if fact.Key != "component" || fact.Snapshot.Colors[0] != color.Red || fact.Snapshot.Subtypes[0] != types.Human ||
			fact.Snapshot.EntryChoices["type"].Subtype != types.Human || !fact.CardZoneVersion.Exists || fact.CardZoneVersion.Val != 0 {
			t.Fatalf("mutable payment fact shared between independent stack objects: %#v", fact)
		}
	}
}
