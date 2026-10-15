package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SmittenSwordmaster is the card definition for Smitten Swordmaster // Curry Favor.
//
// Type: Creature — Human Knight // Sorcery — Adventure
// Cost: {1}{B} // {B}
// Face: Curry Favor — Sorcery — Adventure ({B})
//
// Oracle text:
//
//	Lifelink
var SmittenSwordmaster = newSmittenSwordmaster

func newSmittenSwordmaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Smitten Swordmaster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			OracleText: `
			Lifelink
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Curry Favor",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Knight")}, Controller: game.ControllerYou}),
							}),
							PlayerGroup: game.OpponentsReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You gain X life and each opponent loses X life, where X is the number of Knights you control.
		`,
		}),
	}
}
