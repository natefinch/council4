package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ElvishPromenade is the card definition for Elvish Promenade.
//
// Type: Kindred Sorcery — Elf
// Cost: {3}{G}
//
// Oracle text:
//
//	Create a 1/1 green Elf Warrior creature token for each Elf you control.
var ElvishPromenade = newElvishPromenade

func newElvishPromenade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Elvish Promenade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Kindred, types.Sorcery},
			Subtypes: []types.Sub{types.Elf},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Elf")}, Controller: game.ControllerYou}),
							}),
							Source: game.TokenDef(elvishPromenadeToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 green Elf Warrior creature token for each Elf you control.
		`,
		},
	}
}

var elvishPromenadeToken = newElvishPromenadeToken()

func newElvishPromenadeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elf Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
