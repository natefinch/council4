package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Slagstorm is the card definition for Slagstorm.
//
// Type: Sorcery
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Choose one —
//	• Slagstorm deals 3 damage to each creature.
//	• Slagstorm deals 3 damage to each player.
var Slagstorm = newSlagstorm

func newSlagstorm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Slagstorm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Slagstorm deals 3 damage to each creature.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
								},
							},
						},
					},
					game.Mode{
						Text: "Slagstorm deals 3 damage to each player.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.PlayerGroupDamageRecipient(game.AllPlayersReference()),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Slagstorm deals 3 damage to each creature.
			• Slagstorm deals 3 damage to each player.
		`,
		},
	}
}
