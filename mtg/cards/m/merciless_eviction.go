package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MercilessEviction is the card definition for Merciless Eviction.
//
// Type: Sorcery
// Cost: {4}{W}{B}
//
// Oracle text:
//
//	Choose one —
//	• Exile all artifacts.
//	• Exile all creatures.
//	• Exile all enchantments.
//	• Exile all planeswalkers.
var MercilessEviction = newMercilessEviction

func newMercilessEviction() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Merciless Eviction",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile all artifacts.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Exile all creatures.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Exile all enchantments.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Exile all planeswalkers.",
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Planeswalker}}),
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
			• Exile all artifacts.
			• Exile all creatures.
			• Exile all enchantments.
			• Exile all planeswalkers.
		`,
		},
	}
}
