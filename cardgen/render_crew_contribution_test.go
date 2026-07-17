package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestRenderCrewPowerContributionCost(t *testing.T) {
	t.Parallel()
	rendered, err := renderAdditional(newRenderCtx(), cost.Additional{
		Kind:               cost.AdditionalTapPermanents,
		MatchPermanentType: true,
		PermanentType:      types.Creature,
		TotalPowerAtLeast:  3,
		PowerContribution:  cost.PowerContributionCrew,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"TotalPowerAtLeast: 3",
		"PowerContribution: cost.PowerContributionCrew",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("rendered cost %q missing %q", rendered, want)
		}
	}
}
