package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OwlinShieldmage is the card definition for Owlin Shieldmage.
//
// Type: Creature — Bird Warlock
// Cost: {3}{W}{B}
//
// Oracle text:
//
//	Flying
//	Ward—Pay 3 life. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays 3 life.)
var OwlinShieldmage = newOwlinShieldmage

func newOwlinShieldmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Owlin Shieldmage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Warlock},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalPayLife,
						Text:   "Pay 3 life",
						Amount: 3,
					},
				}),
			},
			OracleText: `
			Flying
			Ward—Pay 3 life. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays 3 life.)
		`,
		},
	}
}
