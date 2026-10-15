package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HurlyBurly is the card definition for Hurly-Burly.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Choose one —
//	• Hurly-Burly deals 1 damage to each creature without flying.
//	• Hurly-Burly deals 1 damage to each creature with flying.
var HurlyBurly = newHurlyBurly

func newHurlyBurly() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hurly-Burly",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Hurly-Burly deals 1 damage to each creature without flying.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying})),
								},
							},
						},
					},
					game.Mode{
						Text: "Hurly-Burly deals 1 damage to each creature with flying.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying})),
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
			• Hurly-Burly deals 1 damage to each creature without flying.
			• Hurly-Burly deals 1 damage to each creature with flying.
		`,
		},
	}
}
