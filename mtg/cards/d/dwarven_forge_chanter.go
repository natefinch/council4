package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DwarvenForgeChanter is the card definition for Dwarven Forge-Chanter.
//
// Type: Creature — Dwarf Wizard
// Cost: {1}{R}
//
// Oracle text:
//
//	Ward—Pay 2 life. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays 2 life.)
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
var DwarvenForgeChanter = newDwarvenForgeChanter

func newDwarvenForgeChanter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dwarven Forge-Chanter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dwarf, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalPayLife,
						Text:   "Pay 2 life",
						Amount: 2,
					},
				}),
				game.ProwessStaticBody,
			},
			OracleText: `
			Ward—Pay 2 life. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays 2 life.)
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
