package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// goreclawPowerReductionEffect models the active rule effect cardgen emits for
// "Creature spells you cast with power 4 or greater cost {2} less to cast.": a
// controller's creature spell cost modifier scoped to spells whose base printed
// power meets the threshold.
func goreclawPowerReductionEffect(g *game.Game, minPower, reduction int) game.RuleEffect {
	return game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCostModifier,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CostModifier: game.CostModifier{
			Kind: game.CostModifierSpell,
			CardSelection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: minPower}),
			},
			GenericReduction: reduction,
		},
	}
}

func creatureSpellWithPower(name string, power opt.V[game.PT]) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Creature},
		Power:    power,
		ManaCost: opt.Val(cost.Mana{cost.O(5)}),
	}}
}

func TestSpellCostModifierPowerThresholdDiscountsAtOrAboveThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, goreclawPowerReductionEffect(g, 4, 2))

	atThreshold := creatureSpellWithPower("Power Four", opt.Val(game.PT{Value: 4}))
	if got := spellGenericReductionFromZone(g, atThreshold, zone.Hand); got != 2 {
		t.Fatalf("reduction for power-4 creature spell = %d, want 2", got)
	}

	aboveThreshold := creatureSpellWithPower("Power Six", opt.Val(game.PT{Value: 6}))
	if got := spellGenericReductionFromZone(g, aboveThreshold, zone.Hand); got != 2 {
		t.Fatalf("reduction for power-6 creature spell = %d, want 2", got)
	}
}

func TestSpellCostModifierPowerThresholdSkipsBelowThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, goreclawPowerReductionEffect(g, 4, 2))

	belowThreshold := creatureSpellWithPower("Power Three", opt.Val(game.PT{Value: 3}))
	if got := spellGenericReductionFromZone(g, belowThreshold, zone.Hand); got != 0 {
		t.Fatalf("reduction for power-3 creature spell = %d, want 0", got)
	}

	starPower := creatureSpellWithPower("Star Power", opt.Val(game.PT{IsStar: true}))
	if got := spellGenericReductionFromZone(g, starPower, zone.Hand); got != 0 {
		t.Fatalf("reduction for star-power creature spell = %d, want 0", got)
	}

	noPower := &game.CardDef{CardFace: game.CardFace{
		Name:     "Bolt",
		Types:    []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(1)}),
	}}
	if got := spellGenericReductionFromZone(g, noPower, zone.Hand); got != 0 {
		t.Fatalf("reduction for powerless spell = %d, want 0", got)
	}
}
