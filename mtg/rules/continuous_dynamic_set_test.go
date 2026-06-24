package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestSnapshotContinuousXFreezesDynamicBasePowerToughness verifies that
// snapshotContinuousX locks the dynamic base power/toughness SET (Mirror Entity,
// Biomass Mutation) to the resolving ability's chosen X at resolution, folding
// SetPowerDynamic/SetToughnessDynamic into the fixed SetPower/SetToughness so the
// frozen value survives later changes to X.
func TestSnapshotContinuousXFreezesDynamicBasePowerToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		XValue:     5,
	}
	effect := game.ContinuousEffect{
		Layer:               game.LayerPowerToughnessSet,
		SetPowerDynamic:     opt.Val(game.DynamicAmount{Kind: game.DynamicAmountX}),
		SetToughnessDynamic: opt.Val(game.DynamicAmount{Kind: game.DynamicAmountX}),
	}

	snapshotContinuousX(g, obj, &effect)

	if effect.SetPowerDynamic.Exists || effect.SetToughnessDynamic.Exists {
		t.Fatalf("snapshot did not clear dynamic set fields: %+v", effect)
	}
	if !effect.SetPower.Exists || effect.SetPower.Val.Value != 5 {
		t.Fatalf("frozen base power = %+v, want fixed 5", effect.SetPower)
	}
	if !effect.SetToughness.Exists || effect.SetToughness.Val.Value != 5 {
		t.Fatalf("frozen base toughness = %+v, want fixed 5", effect.SetToughness)
	}
}

// TestDynamicBasePowerToughnessSetAppliesFrozenValue verifies that a base
// power/toughness SET effect carrying the frozen fixed value sets a creature's
// effective power and toughness through the LayerPowerToughnessSet layer.
func TestDynamicBasePowerToughnessSetAppliesFrozenValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	four := game.PT{Value: 4}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               1,
		AffectedObjectID: creature.ObjectID,
		Layer:            game.LayerPowerToughnessSet,
		SetPower:         opt.Val(four),
		SetToughness:     opt.Val(four),
	})

	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power = %d, want frozen 4", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 4 {
		t.Fatalf("effective toughness = %d ok=%v, want frozen 4 true", got, ok)
	}
}

// TestDynamicBasePowerToughnessSetWithEveryCreatureType verifies the combined
// Mirror Entity shape: a creature group set to a fixed base power/toughness that
// also gains every creature type applies both the P/T set and the type grant.
func TestDynamicBasePowerToughnessSetWithEveryCreatureType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Shifting Ally",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Shapeshifter},
	}})
	three := game.PT{Value: 3}
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerPowerToughnessSet,
			SetPower:         opt.Val(three),
			SetToughness:     opt.Val(three),
		},
		game.ContinuousEffect{
			ID:                   2,
			AffectedObjectID:     creature.ObjectID,
			Layer:                game.LayerType,
			AddEveryCreatureType: true,
		},
	)

	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("effective power = %d, want 3", got)
	}
	for _, subtype := range []types.Sub{types.Goblin, types.Elf, types.Sliver} {
		if !permanentHasSubtype(g, creature, subtype) {
			t.Fatalf("creature not treated as %s after every-creature-type grant", subtype)
		}
	}
}
