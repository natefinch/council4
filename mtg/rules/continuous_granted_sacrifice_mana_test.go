package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// addSacrificeAnyColorFoodGrantSource models Ninja Pizza: it grants every Food
// the controller owns the count-1 tap-and-sacrifice "Add one mana of any color"
// ability.
func addSacrificeAnyColorFoodGrantSource(g *game.Game, controller game.PlayerID) *game.Permanent {
	ability := game.TapSacrificeAnyColorManaAbility("{T}, Sacrifice this artifact: Add one mana of any color.")
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ninja Pizza",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{SubtypesAny: []types.Sub{types.Food}},
				),
				AddAbilities: []game.Ability{&ability},
			}},
		}},
	}})
}

func addFoodPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Food",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Food},
	}})
}

func sacrificeAnyColorManaAbilityIndex(g *game.Game, permanent *game.Permanent) (int, bool) {
	for index, ability := range permanentEffectiveAbilities(g, permanent) {
		if body, ok := ability.(*game.ManaAbility); ok && game.IsTapSacrificeAnyColorManaAbility(body) {
			return index, true
		}
	}
	return 0, false
}

// TestGrantedSacrificeAnyColorManaAbility exercises Ninja Pizza's granted
// continuous mana ability end to end: a controlled Food gains the ability,
// activating it produces the chosen color, and the source Food is sacrificed.
func TestGrantedSacrificeAnyColorManaAbility(t *testing.T) {
	t.Run("granted only to controlled Foods", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		addSacrificeAnyColorFoodGrantSource(g, game.Player1)
		owned := addFoodPermanent(g, game.Player1)
		opposing := addFoodPermanent(g, game.Player2)
		nonFood := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Bear", Types: []types.Card{types.Creature},
		}})

		if _, ok := sacrificeAnyColorManaAbilityIndex(g, owned); !ok {
			t.Fatal("controlled Food did not gain the granted sacrifice mana ability")
		}
		if _, ok := sacrificeAnyColorManaAbilityIndex(g, opposing); ok {
			t.Fatal("opposing Food incorrectly gained the granted ability")
		}
		if _, ok := sacrificeAnyColorManaAbilityIndex(g, nonFood); ok {
			t.Fatal("controlled non-Food incorrectly gained the granted ability")
		}
	})

	t.Run("activation produces chosen color and sacrifices the Food", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addSacrificeAnyColorFoodGrantSource(g, game.Player1)
		food := addFoodPermanent(g, game.Player1)
		index, ok := sacrificeAnyColorManaAbilityIndex(g, food)
		if !ok {
			t.Fatal("granted sacrifice any-color mana ability not found")
		}
		activate := action.ActivateAbility(food.ObjectID, index, nil, 0)
		if !containsAction(engine.legalActions(g, game.Player1), activate) {
			t.Fatal("granted sacrifice mana ability was not exposed for manual activation")
		}
		agents := [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
		}
		if !engine.applyActionWithChoices(g, game.Player1, activate, agents, &TurnLog{}) {
			t.Fatal("manual granted sacrifice mana activation failed")
		}
		if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 1 {
			t.Fatalf("chosen blue mana = %d, want 1", got)
		}
		if slices.Contains(g.Battlefield, food) {
			t.Fatal("Food was not sacrificed after activating its sacrifice mana ability")
		}
	})
}
