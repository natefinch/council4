package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RuinousRampage is the card definition for Ruinous Rampage.
//
// Type: Sorcery
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Choose one —
//	• Ruinous Rampage deals 3 damage to each opponent.
//	• Exile all artifacts with mana value 3 or less.
var RuinousRampage = newRuinousRampage

func newRuinousRampage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ruinous Rampage",
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
						Text: "Ruinous Rampage deals 3 damage to each opponent.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.PlayerGroupDamageRecipient(game.OpponentsReference()),
								},
							},
						},
					},
					game.Mode{
						Text: "Exile all artifacts with mana value 3 or less.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
									Destination: zone.Exile,
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
			• Ruinous Rampage deals 3 damage to each opponent.
			• Exile all artifacts with mana value 3 or less.
		`,
		},
	}
}
