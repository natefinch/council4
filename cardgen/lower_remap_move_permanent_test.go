package cardgen

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestTransformMovePermanent(t *testing.T) {
	t.Parallel()
	for _, destination := range []zone.Type{zone.Hand, zone.Exile, zone.Library, zone.Graveyard} {
		for _, bottom := range []bool{false, true} {
			if bottom && destination != zone.Library {
				continue
			}
			t.Run(destination.String()+"/bottom="+strconv.FormatBool(bottom), func(t *testing.T) {
				t.Parallel()
				original := game.MovePermanent{
					Object:        game.TargetPermanentReference(1),
					Destination:   destination,
					LibraryBottom: bottom,
					Amount:        game.Fixed(2),
					PublishLinked: "moved",
				}
				for _, tc := range []struct {
					name      string
					transform targetIndexTransform
					wantIndex int
				}{
					{"zero", rebaseTransform(0, 7), 1},
					{"rebase", rebaseTransform(3, 7), 4},
					{"mixed", func(kind targetIndexKind, old int) (int, bool) {
						if kind != targetIndexObject || old < 0 || old > 1 {
							return 0, false
						}
						return []int{5, 2}[old], true
					}, 2},
				} {
					t.Run(tc.name, func(t *testing.T) {
						t.Parallel()
						primitive, ok := transformPrimitiveTargetIndices(original, tc.transform)
						if !ok {
							t.Fatal("single-object move was not transformed")
						}
						want := original
						want.Object = game.TargetPermanentReference(tc.wantIndex)
						if !reflect.DeepEqual(primitive, want) {
							t.Fatalf("transformed = %+v, want %+v", primitive, want)
						}
						if original.Object != game.TargetPermanentReference(1) {
							t.Fatal("transform changed its input")
						}
						if err := game.ValidateInstructionSequence([]game.Instruction{{Primitive: primitive}}); err != nil {
							t.Fatalf("transformed instruction is invalid: %v", err)
						}
					})
				}
			})
		}
	}
}

func TestTransformMovePermanentReferences(t *testing.T) {
	t.Parallel()
	for _, ref := range []struct {
		name  string
		build func(int) game.ObjectReference
	}{
		{"permanent", game.TargetPermanentReference},
		{"object", game.TargetObjectReference},
		{"stack", game.TargetStackObjectReference},
		{"attached", game.TargetAttachedPermanentReference},
		{"source", func(int) game.ObjectReference { return game.SourcePermanentReference() }},
		{"linked", func(int) game.ObjectReference { return game.LinkedObjectReference("prior") }},
	} {
		t.Run(ref.name, func(t *testing.T) {
			t.Parallel()
			move := game.MovePermanent{Object: ref.build(0), Destination: zone.Library}
			primitive, ok := rebaseTargetedPrimitive(move, 3, 7)
			if !ok {
				t.Fatal("reference was not transformed")
			}
			want := move
			want.Object = ref.build(3)
			if !reflect.DeepEqual(primitive, want) {
				t.Fatalf("transformed = %+v, want %+v", primitive, want)
			}
		})
	}
}

func TestTransformMovePermanentFailsClosed(t *testing.T) {
	t.Parallel()
	group := game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})
	for _, tc := range []struct {
		name string
		move game.MovePermanent
	}{
		{"missing object", game.MovePermanent{Destination: zone.Library}},
		{"invalid object", game.MovePermanent{Object: game.TargetPermanentReference(-1), Destination: zone.Library}},
		{"group", game.MovePermanent{Group: group, Destination: zone.Library}},
		{"object and group", game.MovePermanent{Object: game.TargetPermanentReference(0), Group: group, Destination: zone.Hand}},
		{"controlled choice", game.MovePermanent{Group: group, ControlledChoice: true, Amount: game.Fixed(1), Destination: zone.Exile}},
		{"invalid object choice", game.MovePermanent{Object: game.TargetPermanentReference(0), ControlledChoice: true, Destination: zone.Hand}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, ok := rebaseTargetedPrimitive(tc.move, 3, 0); ok {
				t.Fatal("unsupported move was rebased")
			}
			if remapTargetedSequence([]game.Instruction{{Primitive: tc.move}}, []int{2}) {
				t.Fatal("unsupported move was remapped")
			}
		})
	}
	for _, index := range []int{-1, 1} {
		move := game.MovePermanent{Object: game.TargetPermanentReference(index), Destination: zone.Library}
		if remapTargetedSequence([]game.Instruction{{Primitive: move}}, []int{2}) {
			t.Fatalf("missing local target %d was remapped", index)
		}
	}
}

func TestTransformMovePermanentQuantity(t *testing.T) {
	t.Parallel()
	move := game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Library,
		Amount: game.Dynamic(game.DynamicAmount{
			Kind: game.DynamicAmountObjectPower, Object: game.TargetPermanentReference(1), Multiplier: 1,
		}),
	}

	seq := []game.Instruction{{Primitive: move}}
	if !remapTargetedSequence(seq, []int{4, 2}) {
		t.Fatal("quantity reference was not remapped")
	}
	want := move
	want.Object = game.TargetPermanentReference(4)
	want.Amount = game.Dynamic(game.DynamicAmount{
		Kind: game.DynamicAmountObjectPower, Object: game.TargetPermanentReference(2), Multiplier: 1,
	})
	if !reflect.DeepEqual(seq[0].Primitive, want) {
		t.Fatalf("remapped = %+v, want %+v", seq[0].Primitive, want)
	}
	if remapTargetedSequence([]game.Instruction{{Primitive: move}}, []int{4}) {
		t.Fatal("missing quantity target was remapped")
	}
}

func TestRemapSharedMovePermanentAndPlayers(t *testing.T) {
	t.Parallel()
	seq := []game.Instruction{
		{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(1), Destination: zone.Library, LibraryBottom: true}},
		{Primitive: game.LoseLife{Player: game.ObjectControllerReference(game.TargetPermanentReference(1)), Amount: game.Fixed(2)}},
		{Primitive: game.GainLife{Player: game.ObjectOwnerReference(game.TargetPermanentReference(1)), Amount: game.Fixed(1)}},
	}
	if !remapTargetedSequence(seq, []int{5, 2}) {
		t.Fatal("shared move and player references were not remapped")
	}
	want := []game.Instruction{
		{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(2), Destination: zone.Library, LibraryBottom: true}},
		{Primitive: game.LoseLife{Player: game.ObjectControllerReference(game.TargetPermanentReference(2)), Amount: game.Fixed(2)}},
		{Primitive: game.GainLife{Player: game.ObjectOwnerReference(game.TargetPermanentReference(2)), Amount: game.Fixed(1)}},
	}
	if !reflect.DeepEqual(seq, want) {
		t.Fatalf("remapped = %+v, want %+v", seq, want)
	}
}

func TestRemappedMovePermanentRetainsValidation(t *testing.T) {
	t.Parallel()
	for _, move := range []game.MovePermanent{
		{Object: game.TargetPermanentReference(0)},
		{Object: game.TargetPermanentReference(0), Destination: zone.Battlefield},
		{Object: game.TargetPermanentReference(0), Destination: zone.Hand, LibraryBottom: true},
	} {
		primitive, ok := rebaseTargetedPrimitive(move, 1, 0)
		if !ok {
			continue
		}
		if err := game.ValidateInstructionSequence([]game.Instruction{{Primitive: primitive}}); err == nil {
			t.Fatalf("invalid move passed validation: %+v", primitive)
		}
	}
}
