package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FrogtosserBanneret is the card definition for Frogtosser Banneret.
//
// Type: Creature — Goblin Rogue
// Cost: {1}{B}
//
// Oracle text:
//
//	Haste
//	Goblin spells and Rogue spells you cast cost {1} less to cast.
var FrogtosserBanneret = newFrogtosserBanneret()

func newFrogtosserBanneret() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Frogtosser Banneret",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin"), types.Sub("Rogue")}},
								GenericReduction: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Haste
			Goblin spells and Rogue spells you cast cost {1} less to cast.
		`,
		},
	}
}
