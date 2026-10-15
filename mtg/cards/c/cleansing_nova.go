package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CleansingNova is the card definition for Cleansing Nova.
//
// Type: Sorcery
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Choose one —
//	• Destroy all creatures.
//	• Destroy all artifacts and enchantments.
var CleansingNova = newCleansingNova

func newCleansingNova() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Cleansing Nova",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Destroy all creatures.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy all artifacts and enchantments.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
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
			• Destroy all creatures.
			• Destroy all artifacts and enchantments.
		`,
		},
	}
}
