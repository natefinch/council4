package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EndTheFestivities is the card definition for End the Festivities.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	End the Festivities deals 1 damage to each opponent and each creature and planeswalker they control.
var EndTheFestivities = newEndTheFestivities()

func newEndTheFestivities() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "End the Festivities",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Controller: game.ControllerOpponent})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			End the Festivities deals 1 damage to each opponent and each creature and planeswalker they control.
		`,
		},
	}
}
