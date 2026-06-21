package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// floweringOfTheWhiteTree mirrors the static ability the executable backend
// generates for "Legendary creatures you control get +2/+1 and have ward {1}.
// Nonlegendary creatures you control get +1/+1." — a supertype-filtered anthem
// pair: the Legendary group gets the P/T buff plus a granted Ward keyword, and
// the excluded-Legendary (nonlegendary) group gets its own smaller P/T buff.
func floweringOfTheWhiteTree(g *game.Game, controller game.PlayerID) *game.Permanent {
	ward := game.WardStaticAbility(cost.Mana{cost.O(1)})
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Flowering of the White Tree",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer: game.LayerPowerToughnessModify,
					Group: game.ObjectControlledGroup(
						game.SourcePermanentReference(),
						game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Supertypes:    []types.Super{types.Legendary},
						},
					),
					PowerDelta:     2,
					ToughnessDelta: 1,
				},
				{
					Layer: game.LayerAbility,
					Group: game.ObjectControlledGroup(
						game.SourcePermanentReference(),
						game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Supertypes:    []types.Super{types.Legendary},
						},
					),
					AddAbilities: []game.Ability{&ward},
				},
				{
					Layer: game.LayerPowerToughnessModify,
					Group: game.ObjectControlledGroup(
						game.SourcePermanentReference(),
						game.Selection{
							RequiredTypes:     []types.Card{types.Creature},
							ExcludedSupertype: types.Legendary,
						},
					),
					PowerDelta:     1,
					ToughnessDelta: 1,
				},
			},
		}},
	}})
}

func permanentHasGrantedWard(g *game.Game, permanent *game.Permanent) bool {
	for _, ability := range permanentEffectiveAbilities(g, permanent) {
		if body, ok := ability.(*game.StaticAbility); ok {
			if _, ok := game.StaticBodyWardCost(body); ok {
				return true
			}
		}
	}
	return false
}

// TestFloweringSupertypeAnthemBuffsByLegendaryStatus verifies the supertype
// filter routes each anthem to the right group: a legendary creature you control
// gets +2/+1 and ward, a nonlegendary creature you control gets +1/+1 and no
// ward, and an opponent's legendary creature is unaffected.
func TestFloweringSupertypeAnthemBuffsByLegendaryStatus(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	floweringOfTheWhiteTree(g, game.Player1)

	legendary := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Legendary Hero",
		Types:      []types.Card{types.Creature},
		Supertypes: []types.Super{types.Legendary},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	}})
	nonlegendary := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	opponentLegendary := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:       "Rival Legend",
		Types:      []types.Card{types.Creature},
		Supertypes: []types.Super{types.Legendary},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	}})

	if got := effectivePower(g, legendary); got != 4 {
		t.Fatalf("legendary power = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, legendary); !ok || got != 3 {
		t.Fatalf("legendary toughness = %d (ok=%v), want 3", got, ok)
	}
	if !permanentHasGrantedWard(g, legendary) {
		t.Fatal("legendary creature lacks granted ward")
	}

	if got := effectivePower(g, nonlegendary); got != 3 {
		t.Fatalf("nonlegendary power = %d, want 3", got)
	}
	if got, ok := effectiveToughness(g, nonlegendary); !ok || got != 3 {
		t.Fatalf("nonlegendary toughness = %d (ok=%v), want 3", got, ok)
	}
	if permanentHasGrantedWard(g, nonlegendary) {
		t.Fatal("nonlegendary creature should not have ward")
	}

	if got := effectivePower(g, opponentLegendary); got != 2 {
		t.Fatalf("opponent legendary power = %d, want 2 (unaffected)", got)
	}
	if permanentHasGrantedWard(g, opponentLegendary) {
		t.Fatal("opponent legendary creature should not have ward")
	}
}
