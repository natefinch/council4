package payment

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// TestPayGenericByTappingLeavesNeededSources checks the selection of permanents
// a nonmana mechanic taps for generic mana (CR 702.51a, 702.126a). The stub
// remaining payment needs source untapped (CR 106.12), which taps for one mana
// and so pays at most one generic mana the mechanic doesn't.
func TestPayGenericByTappingLeavesNeededSources(t *testing.T) {
	permanent := func(objectID id.ID) *game.Permanent { return &game.Permanent{ObjectID: objectID} }
	source, first, second, colored := permanent(1), permanent(2), permanent(3), permanent(4)
	for _, tc := range []struct {
		name       string
		generic    int
		tapped     []*game.Permanent
		candidates []*game.Permanent
		want       []*game.Permanent
	}{
		{name: "all candidates", generic: 3, candidates: []*game.Permanent{first, second}, want: []*game.Permanent{first, second}},
		{name: "fewer than the generic mana", generic: 2, candidates: []*game.Permanent{first, source}, want: []*game.Permanent{first}},
		{name: "skips a needed earlier source", generic: 3, candidates: []*game.Permanent{source, first, second}, want: []*game.Permanent{first, second}},
		{name: "keeps colored taps", generic: 1, tapped: []*game.Permanent{colored}, candidates: []*game.Permanent{source}, want: []*game.Permanent{colored}},
		{name: "too much generic mana", generic: 4, candidates: []*game.Permanent{source, first}},
		{name: "taps nothing", generic: 1, candidates: []*game.Permanent{source}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &nonManaPaymentCost{generic: tc.generic}
			payRemaining := func(c *nonManaPaymentCost, excluded []*game.Permanent) (paymentPlan, bool) {
				if c.genericDue() > 1 || slices.Contains(excluded, source) {
					return paymentPlan{}, false
				}
				return paymentPlan{manaTaps: []manaTap{{permanent: source}}}, true
			}

			taps, plan, ok := payGenericByTapping(c, tc.tapped, tc.candidates, payRemaining)

			if ok != (tc.want != nil) || !slices.Equal(taps, tc.want) {
				t.Fatalf("taps=%v ok=%t, want %v", taps, ok, tc.want)
			}
			if ok && (len(plan.manaTaps) != 1 || slices.Contains(taps, source) || c.reduction != len(taps)-len(tc.tapped)) {
				t.Fatalf("plan=%+v reduction=%d, want source tapped for mana and %d generic paid by tapping", plan, c.reduction, len(taps)-len(tc.tapped))
			}
		})
	}
}
