package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerCanopyCover(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Canopy Cover",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		ManaCost:   "{1}{G}",
		OracleText: "Enchant creature\nEnchanted creature can't be blocked except by creatures with flying or reach.\nEnchanted creature can't be the target of spells or abilities your opponents control.",
	})
	if len(face.StaticAbilities) != 3 {
		t.Fatalf("static abilities = %#v", face.StaticAbilities)
	}
	var blockedExcept, targetRestricted bool
	for _, ability := range face.StaticAbilities {
		for _, effect := range ability.Body.RuleEffects {
			if effect.Kind == game.RuleEffectCantBeBlockedExceptBy {
				blockedExcept = effect.AffectedAttached &&
					effect.BlockerRestriction.Kind == game.BlockerRestrictionFlyingOrReach
			}
			if effect.Kind == game.RuleEffectCantBeTargetedByControllerOpponents {
				targetRestricted = effect.AffectedAttached
			}
		}
	}
	if !blockedExcept || !targetRestricted {
		t.Fatalf("static abilities = %#v", face.StaticAbilities)
	}
}
