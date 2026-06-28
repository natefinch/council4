package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BallyrushBanneret is the card definition for Ballyrush Banneret.
//
// Type: Creature — Kithkin Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	Kithkin spells and Soldier spells you cast cost {1} less to cast.
var BallyrushBanneret = newBallyrushBanneret()

func newBallyrushBanneret() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ballyrush Banneret",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kithkin, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Sub("Kithkin"), types.Sub("Soldier")}},
								GenericReduction: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Kithkin spells and Soldier spells you cast cost {1} less to cast.
		`,
		},
	}
}
