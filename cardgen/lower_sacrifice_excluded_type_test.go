package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerSacrificeNonlandPermanentCost verifies the "Sacrifice N nonland
// permanents" activation cost (Bolas's Citadel, Magmaw, Rite of Oblivion) lowers
// to a sacrifice whose ExcludePermanentType bars lands from being chosen.
func TestLowerSacrificeNonlandPermanentCost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantAmount int
	}{
		{
			name:       "single nonland permanent",
			oracleText: "{1}, Sacrifice a nonland permanent: Draw a card.",
			wantAmount: 1,
		},
		{
			name:       "ten nonland permanents",
			oracleText: "{T}, Sacrifice ten nonland permanents: Draw a card.",
			wantAmount: 10,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Citadel",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: test.oracleText,
			})
			if len(face.ActivatedAbilities) != 1 {
				t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
			}
			costs := face.ActivatedAbilities[0].AdditionalCosts
			var sacrifice *cost.Additional
			for i := range costs {
				if costs[i].Kind == cost.AdditionalSacrifice {
					sacrifice = &costs[i]
				}
			}
			if sacrifice == nil {
				t.Fatalf("additional costs = %#v, want a sacrifice", costs)
			}
			if sacrifice.Amount != test.wantAmount {
				t.Fatalf("sacrifice amount = %d, want %d", sacrifice.Amount, test.wantAmount)
			}
			if sacrifice.ExcludePermanentType != types.Land {
				t.Fatalf("sacrifice exclude type = %q, want Land", sacrifice.ExcludePermanentType)
			}
			if sacrifice.MatchPermanentType {
				t.Fatalf("sacrifice must not restrict to a single permanent type: %#v", sacrifice)
			}
		})
	}
}
