package game

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestEternalizeAbilityBuildsKeywordActivation(t *testing.T) {
	cost := mana.Cost{mana.GenericMana(2), mana.ColoredMana(mana.Green)}
	ability := EternalizeAbility(cost, types.Snake, types.Druid)
	cost[0] = mana.GenericMana(9)

	if ability.Kind != ActivatedAbility || !slices.Equal(ability.Keywords, []Keyword{Eternalize}) {
		t.Fatalf("ability kind/keywords = %v/%+v, want Eternalize activated ability", ability.Kind, ability.Keywords)
	}
	if ability.ZoneOfFunction != ZoneGraveyard || ability.Timing != SorceryOnly {
		t.Fatalf("zone/timing = %v/%v, want graveyard sorcery", ability.ZoneOfFunction, ability.Timing)
	}
	if !ability.ManaCost.Exists || !slices.Equal(ability.ManaCost.Val, []mana.Symbol{mana.GenericMana(2), mana.ColoredMana(mana.Green)}) {
		t.Fatalf("mana cost = %+v, want copied eternalize cost", ability.ManaCost)
	}
	if len(ability.AdditionalCosts) != 1 || ability.AdditionalCosts[0].Kind != AdditionalCostExileSource {
		t.Fatalf("additional costs = %+v, want source exile", ability.AdditionalCosts)
	}
	if len(ability.Effects) != 1 || ability.Effects[0].Type != EffectCreateToken || !ability.Effects[0].TokenCopy.Exists {
		t.Fatalf("effects = %+v, want create token-copy effect", ability.Effects)
	}
	spec := ability.Effects[0].TokenCopy.Val
	if spec.Source != TokenCopySourceSourceCard || !spec.NoManaCost {
		t.Fatalf("token copy source/no-cost = %v/%v, want source card with no mana cost", spec.Source, spec.NoManaCost)
	}
	if !slices.Equal(spec.SetColors, []mana.Color{mana.Black}) || !slices.Equal(spec.SetSubtypes, []types.Sub{types.Zombie, types.Snake, types.Druid}) {
		t.Fatalf("token colors/subtypes = %+v/%+v, want black Zombie Snake Druid", spec.SetColors, spec.SetSubtypes)
	}
	if !spec.SetPower.Exists || spec.SetPower.Val.Value != 4 || !spec.SetToughness.Exists || spec.SetToughness.Val.Value != 4 {
		t.Fatalf("token P/T = %+v/%+v, want 4/4", spec.SetPower, spec.SetToughness)
	}
}
