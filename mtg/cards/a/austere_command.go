package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AustereCommand is the card definition for Austere Command.
//
// Type: Sorcery
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Choose two —
//	• Destroy all artifacts.
//	• Destroy all enchantments.
//	• Destroy all creatures with mana value 3 or less.
//	• Destroy all creatures with mana value 4 or greater.
var AustereCommand = newAustereCommand

func newAustereCommand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Austere Command",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Destroy all artifacts.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy all enchantments.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy all creatures with mana value 3 or less.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy all creatures with mana value 4 or greater.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}),
								},
							},
						},
					},
				},
				MinModes: 2,
				MaxModes: 2,
			}),
			OracleText: `
			Choose two —
			• Destroy all artifacts.
			• Destroy all enchantments.
			• Destroy all creatures with mana value 3 or less.
			• Destroy all creatures with mana value 4 or greater.
		`,
		},
	}
}
