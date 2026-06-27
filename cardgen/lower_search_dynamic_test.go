package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerDynamicCountSearchSpecs covers the dynamic-count basic-land ramp
// search family: "Search your library for up to X <filter> cards, where X is
// <rules-derived count>", in both the single-sentence ("..., put them ...") and
// two-sentence ("... . Put those cards ...") forms. The search count lowers to a
// dynamic Quantity whose formula matches the "where X is ..." clause.
func TestLowerDynamicCountSearchSpecs(t *testing.T) {
	t.Parallel()
	landsYouControl := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Land},
		Controller:    game.ControllerYou,
	})
	creaturesYouControl := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	})
	basicLandBattlefieldTapped := game.SearchSpec{
		SourceZone:   zone.Library,
		Destination:  zone.Battlefield,
		EntersTapped: true,
		Filter: game.Selection{
			RequiredTypes: []types.Card{types.Land},
			Supertypes:    []types.Super{types.Basic},
		},
	}
	tests := []struct {
		name       string
		oracleText string
		spec       game.SearchSpec
		dynamic    game.DynamicAmount
	}{
		{
			name:       "single-sentence count of lands you control",
			oracleText: "Search your library for up to X basic land cards, where X is the number of lands you control, put them onto the battlefield tapped, then shuffle.",
			spec:       basicLandBattlefieldTapped,
			dynamic: game.DynamicAmount{
				Kind:       game.DynamicAmountCountSelector,
				Multiplier: 1,
				Group:      landsYouControl,
			},
		},
		{
			name:       "two-sentence greatest power among creatures you control",
			oracleText: "Search your library for up to X basic land cards, where X is the greatest power among creatures you control. Put those cards onto the battlefield tapped, then shuffle.",
			spec:       basicLandBattlefieldTapped,
			dynamic: game.DynamicAmount{
				Kind:       game.DynamicAmountGreatestPowerInGroup,
				Multiplier: 1,
				Group:      creaturesYouControl,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			search := loweredSearch(t, "Sorcery", test.oracleText)
			if !search.Amount.IsDynamic() {
				t.Fatalf("amount = %+v, want dynamic", search.Amount)
			}
			got := search.Amount.DynamicAmount()
			if !got.Exists || !reflect.DeepEqual(got.Val, test.dynamic) {
				t.Errorf("dynamic amount = %+v, want %+v", got.Val, test.dynamic)
			}
			if !searchSpecEqual(search.Spec, test.spec) {
				t.Errorf("spec = %+v, want %+v", search.Spec, test.spec)
			}
		})
	}
}
