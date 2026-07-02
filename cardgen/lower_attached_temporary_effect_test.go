package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerAttachedTemporaryModifyPT(t *testing.T) {
	t.Parallel()
	continuous := attachedTemporaryContinuous(t, &ScryfallCard{
		Name:       "Test Aura",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		OracleText: "Enchant creature\n{R}: Enchanted creature gets +1/+0 until end of turn.",
	})
	if continuous.Layer != game.LayerPowerToughnessModify ||
		continuous.PowerDelta != 1 ||
		continuous.ToughnessDelta != 0 {
		t.Fatalf("continuous effect = %#v, want attached +1/+0", continuous)
	}
}

func TestLowerAttachedTemporaryKeywordGrant(t *testing.T) {
	t.Parallel()
	continuous := attachedTemporaryContinuous(t, &ScryfallCard{
		Name:       "Test Equipment",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "{1}: Equipped creature gains flying until end of turn.\nEquip {2}",
	})
	if continuous.Layer != game.LayerAbility ||
		len(continuous.AddKeywords) != 1 ||
		continuous.AddKeywords[0] != game.Flying {
		t.Fatalf("continuous effect = %#v, want attached flying", continuous)
	}
}

func attachedTemporaryContinuous(t *testing.T, card *ScryfallCard) game.ContinuousEffect {
	t.Helper()
	face := lowerSingleFace(t, card)
	for _, ability := range face.ActivatedAbilities {
		for _, mode := range ability.Content.Modes {
			for _, instruction := range mode.Sequence {
				apply, ok := instruction.Primitive.(game.ApplyContinuous)
				if !ok || len(apply.ContinuousEffects) != 1 {
					continue
				}
				if apply.Duration != game.DurationUntilEndOfTurn {
					t.Fatalf("duration = %v, want until end of turn", apply.Duration)
				}
				group := apply.ContinuousEffects[0].Group
				anchor, ok := group.Anchor()
				if group.Domain() != game.GroupDomainAttachedObject ||
					!ok ||
					anchor != game.SourcePermanentReference() {
					t.Fatalf("group = %#v, want source-attached object", group)
				}
				return apply.ContinuousEffects[0]
			}
		}
	}
	t.Fatal("no attached ApplyContinuous instruction")
	return game.ContinuousEffect{}
}
