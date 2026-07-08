package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Wildfire is the card definition for Wildfire.
//
// Type: Sorcery
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Each player sacrifices four lands of their choice. Wildfire deals 4 damage to each creature.
var Wildfire = newWildfire

func newWildfire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Wildfire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							Amount:      game.Fixed(4),
							PlayerGroup: game.AllPlayersReference(),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each player sacrifices four lands of their choice. Wildfire deals 4 damage to each creature.
		`,
		},
	}
}
