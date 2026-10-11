package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func TestSearchTargetProjection(t *testing.T) {
	t.Parallel()
	original := game.Search{
		Player: game.TargetPlayerReference(0), Controller: opt.Val(game.TargetPlayerReference(1)),
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountObjectPower, Object: game.TargetPermanentReference(2)}),
	}
	projected, ok := rebaseTargetedPrimitive(original, 5, 3)
	if !ok {
		t.Fatal("Search target projection refused")
	}
	search, ok := projected.(game.Search)
	if !ok {
		t.Fatalf("projected primitive=%T, want Search", projected)
	}
	if search.Player.TargetIndex() != 5 || search.Controller.Val.TargetIndex() != 6 ||
		search.Amount.DynamicAmount().Val.Object.TargetIndex() != 7 ||
		original.Amount.DynamicAmount().Val.Object.TargetIndex() != 2 {
		t.Fatal("independent Search operands were aliased or mutated")
	}
	sequence := []game.Instruction{{Primitive: original}}
	if remapTargetedSequence(sequence, []int{4}) || !reflect.DeepEqual(sequence[0].Primitive, original) {
		t.Fatal("a missing Controller/Amount target silently survived projection")
	}
}

func TestSearchNestedTargetBoundsRefuse(t *testing.T) {
	t.Parallel()
	targetGroup := game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{})
	bound := game.Selection{ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{
		Kind: game.DynamicAmountCountSelector, Group: &targetGroup,
	})}
	for _, spec := range []game.SearchSpec{
		{Filter: bound},
		{Filter: game.Selection{AnyOf: []game.Selection{{AnyOf: []game.Selection{bound}}}}},
		{SlotFilters: []game.Selection{{AnyOf: []game.Selection{bound}}}},
	} {
		search := game.Search{Player: game.ControllerReference(), Amount: game.Fixed(1), Spec: spec}
		if _, ok := rebaseTargetedPrimitive(search, 1, 0); ok {
			t.Fatal("a target-dependent nested MV bound was treated as executable without a stack object")
		}
	}
	wholeGroup := game.PlayerGroupControlledGroup(game.TargetedPlayersReference(), game.Selection{})
	if !searchSelectionHasTargetBound(game.Selection{ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{
		Kind: game.DynamicAmountCountSelector, Group: &wholeGroup,
	})}) {
		t.Fatal("a whole targeted-player group escaped the Search bound refusal")
	}
	controllerGroup := game.PlayerControlledGroup(game.ControllerReference(), game.Selection{})
	spec := game.SearchSpec{Filter: game.Selection{AnyOf: []game.Selection{{ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{
		Kind: game.DynamicAmountCountSelector, Group: &controllerGroup,
	})}}}}
	if _, ok := rebaseTargetedPrimitive(game.Search{Player: game.ControllerReference(), Amount: game.Fixed(1), Spec: spec}, 1, 0); !ok {
		t.Fatal("the existing controller-only bound was unnecessarily refused")
	}
}
