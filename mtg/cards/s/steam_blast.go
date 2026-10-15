package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SteamBlast is the card definition for Steam Blast.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Steam Blast deals 2 damage to each creature and each player.
var SteamBlast = newSteamBlast

func newSteamBlast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Steam Blast",
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
			Steam Blast deals 2 damage to each creature and each player.
		`,
		},
	}
}
