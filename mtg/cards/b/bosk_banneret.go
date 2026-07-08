package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BoskBanneret is the card definition for Bosk Banneret.
//
// Type: Creature — Treefolk Shaman
// Cost: {1}{G}
//
// Oracle text:
//
//	Treefolk spells and Shaman spells you cast cost {1} less to cast.
var BoskBanneret = newBoskBanneret

func newBoskBanneret() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bosk Banneret",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Shaman},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Sub("Treefolk"), types.Sub("Shaman")}},
								GenericReduction: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Treefolk spells and Shaman spells you cast cost {1} less to cast.
		`,
		},
	}
}
