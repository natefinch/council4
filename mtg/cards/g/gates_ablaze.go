package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GatesAblaze is the card definition for Gates Ablaze.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Gates Ablaze deals X damage to each creature, where X is the number of Gates you control.
var GatesAblaze = newGatesAblaze

func newGatesAblaze() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Gates Ablaze",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Gate")}, Controller: game.ControllerYou}),
							}),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Gates Ablaze deals X damage to each creature, where X is the number of Gates you control.
		`,
		},
	}
}
