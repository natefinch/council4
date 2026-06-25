package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestLowerStaticSelectionUnifiedRouting pins the unified routing of
// lowerStaticSelection: its shared dimension cluster (required card types,
// excluded supertypes, controller, tap state) flows through the canonical
// SelectionForSelector projector while the static-affected-group extras
// (MatchAnyCounter, the power threshold) ride the projector result. The expected
// Selection matches the hand-written projector this routing replaced, guarding
// against any divergence between the shared core and the local rider.
func TestLowerStaticSelectionUnifiedRouting(t *testing.T) {
	t.Parallel()
	selection := compiler.StaticSelection{
		RequiredTypes:      []types.Card{types.Creature},
		ExcludedSupertypes: []types.Super{types.Legendary},
		Controller:         compiler.ControllerYou,
		TapState:           compiler.StaticTapStateTapped,
		MatchAnyCounter:    true,
		MatchPower:         true,
		Power:              compare.Int{Op: compare.LessOrEqual, Value: 2},
	}
	got, ok := lowerStaticSelection(selection)
	if !ok {
		t.Fatal("lowerStaticSelection ok = false, want true")
	}
	want := game.Selection{
		RequiredTypes:     []types.Card{types.Creature},
		ExcludedSupertype: types.Legendary,
		Controller:        game.ControllerYou,
		Tapped:            game.TriTrue,
		MatchAnyCounter:   true,
		Power:             opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2}),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("lowerStaticSelection = %#v, want %#v", got, want)
	}
}
