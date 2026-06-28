package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StonybrookBanneret is the card definition for Stonybrook Banneret.
//
// Type: Creature — Merfolk Wizard
// Cost: {1}{U}
//
// Oracle text:
//
//	Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
//	Merfolk spells and Wizard spells you cast cost {1} less to cast.
var StonybrookBanneret = newStonybrookBanneret()

func newStonybrookBanneret() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Stonybrook Banneret",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.IslandwalkStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Sub("Merfolk"), types.Sub("Wizard")}},
								GenericReduction: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
			Merfolk spells and Wizard spells you cast cost {1} less to cast.
		`,
		},
	}
}
