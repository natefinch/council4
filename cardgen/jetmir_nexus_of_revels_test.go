package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerJetmirNexusOfRevelsThresholdAnthems verifies that Jetmir's three
// stacked, threshold-gated anthem lines lower into three independent conditional
// static abilities. Each line grants the controller's creatures +1/+0 and a
// keyword, gated by a "you control N or more creatures" control-count condition,
// and each threshold stacks on top of the previous one. The second and third
// lines print an "also" adverb between the group and its verb ("Creatures you
// control also get ..."); the lowering must treat them identically to the first
// line so all three anthems land.
func TestLowerJetmirNexusOfRevelsThresholdAnthems(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Jetmir, Nexus of Revels",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Cat Demon",
		ManaCost:   "{2}{R}{G}{W}",
		OracleText: "Creatures you control get +1/+0 and have vigilance as long as you control three or more creatures.\nCreatures you control also get +1/+0 and have trample as long as you control six or more creatures.\nCreatures you control also get +1/+0 and have double strike as long as you control nine or more creatures.",
	})

	wantGroup := game.ObjectControlledGroup(
		game.SourcePermanentReference(),
		game.Selection{RequiredTypes: []types.Card{types.Creature}},
	)
	thresholds := map[int]game.Keyword{
		3: game.Vigilance,
		6: game.Trample,
		9: game.DoubleStrike,
	}

	found := map[int]bool{}
	for i := range face.StaticAbilities {
		body := face.StaticAbilities[i].Body
		if !body.Condition.Exists || !body.Condition.Val.ControlsMatching.Exists {
			continue
		}
		count := body.Condition.Val.ControlsMatching.Val
		if !reflect.DeepEqual(count.Selection.RequiredTypes, []types.Card{types.Creature}) {
			t.Fatalf("static[%d] condition selection = %#v, want creatures", i, count.Selection)
		}
		wantKeyword, ok := thresholds[count.MinCount]
		if !ok {
			t.Fatalf("static[%d] unexpected threshold MinCount = %d", i, count.MinCount)
		}

		var sawPowerToughness, sawKeyword bool
		for _, effect := range body.ContinuousEffects {
			if !reflect.DeepEqual(effect.Group, wantGroup) {
				t.Fatalf("static[%d] effect group = %#v, want creatures you control", i, effect.Group)
			}
			switch effect.Layer {
			case game.LayerPowerToughnessModify:
				if effect.PowerDelta != 1 || effect.ToughnessDelta != 0 {
					t.Fatalf("static[%d] power/toughness delta = %d/%d, want +1/+0", i, effect.PowerDelta, effect.ToughnessDelta)
				}
				sawPowerToughness = true
			case game.LayerAbility:
				if !reflect.DeepEqual(effect.AddKeywords, []game.Keyword{wantKeyword}) {
					t.Fatalf("static[%d] keywords = %v, want %v", i, effect.AddKeywords, wantKeyword)
				}
				sawKeyword = true
			default:
			}
		}
		if !sawPowerToughness || !sawKeyword {
			t.Fatalf("static[%d] threshold %d missing effects: pt=%v keyword=%v", i, count.MinCount, sawPowerToughness, sawKeyword)
		}
		found[count.MinCount] = true
	}

	for threshold := range thresholds {
		if !found[threshold] {
			t.Fatalf("no conditional anthem for %d-creature threshold; static abilities = %#v", threshold, face.StaticAbilities)
		}
	}
}
