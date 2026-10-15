package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TetsukoUmezawaFugitive is the card definition for Tetsuko Umezawa, Fugitive.
//
// Type: Legendary Creature — Human Rogue
// Cost: {1}{U}
//
// Oracle text:
//
//	Creatures you control with power or toughness 1 or less can't be blocked.
var TetsukoUmezawaFugitive = newTetsukoUmezawaFugitive

func newTetsukoUmezawaFugitive() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tetsuko Umezawa, Fugitive",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control with power or toughness 1 or less can't be blocked.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectCantBeBlocked,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
							AffectedSelection: game.Selection{AnyOf: []game.Selection{game.Selection{Power: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 1,
							})}, game.Selection{Toughness: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 1,
							})}}},
						},
					},
				},
			},
			OracleText: `
			Creatures you control with power or toughness 1 or less can't be blocked.
		`,
		},
	}
}
