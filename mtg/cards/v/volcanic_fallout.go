package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VolcanicFallout is the card definition for Volcanic Fallout.
//
// Type: Instant
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	This spell can't be countered.
//	Volcanic Fallout deals 2 damage to each creature and each player.
var VolcanicFallout = newVolcanicFallout

func newVolcanicFallout() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Volcanic Fallout",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.PlayerGroupDamageRecipient(game.AllPlayersReference()),
						},
					},
				},
			}.Ability()),
			OracleText: `
			This spell can't be countered.
			Volcanic Fallout deals 2 damage to each creature and each player.
		`,
		},
	}
}
