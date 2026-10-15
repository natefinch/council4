package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Replenish is the card definition for Replenish.
//
// Type: Sorcery
// Cost: {3}{W}
//
// Oracle text:
//
//	Return all enchantment cards from your graveyard to the battlefield. (Auras with nothing to enchant remain in your graveyard.)
var Replenish = newReplenish

func newReplenish() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Replenish",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.MassReturnFromGraveyard{
							Player:      game.ControllerReference(),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou},
							Destination: zone.Battlefield,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return all enchantment cards from your graveyard to the battlefield. (Auras with nothing to enchant remain in your graveyard.)
		`,
		},
	}
}
