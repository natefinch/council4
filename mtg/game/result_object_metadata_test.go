package game

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestResultObjectMetadataViews(t *testing.T) {
	selection := Selection{RequiredTypes: []types.Card{types.Creature}}
	primitive := GainLife{Player: ControllerReference(), Amount: Dynamic(DynamicAmount{
		Kind: DynamicAmountMaxOf, Operands: []DynamicAmount{
			{Kind: DynamicAmountPreviousEffectResult, ResultKey: "scalar"},
			{Kind: DynamicAmountPreviousEffectResult, ResultKey: "cards", Selection: &selection},
		},
	})}
	if got := ConsumedCardResultKeys(primitive); !slices.Equal(got, []ResultKey{"cards"}) {
		t.Fatalf("card-result keys=%v, want only cards", got)
	}
	keys := ConsumedCardResultKeys(primitive)
	keys[0] = "mutated"
	if !slices.Equal(ConsumedCardResultKeys(primitive), []ResultKey{"cards"}) {
		t.Fatal("metadata view aliased stored quantity keys")
	}
	if ConsumedCardResultKeys(nil) != nil || PublishesResultObjects(nil) || PublishesResultObjects(primitive) {
		t.Fatal("nil or scalar-only primitive advertised result-object capability")
	}
	if !PublishesResultObjects(MoveTopOfLibrary{Destination: zone.Graveyard}) ||
		!PublishesResultObjects(Discard{}) || !PublishesResultObjects(Destroy{}) ||
		PublishesResultObjects(MoveCard{Destination: zone.Battlefield}) {
		t.Fatal("metadata helper changed the sealed result-object publisher capability")
	}
}
