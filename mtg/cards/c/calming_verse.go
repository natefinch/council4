package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CalmingVerse is the card definition for Calming Verse.
//
// Type: Sorcery
// Cost: {3}{G}
//
// Oracle text:
//
//	Destroy all enchantments you don't control. Then if you control an untapped land, destroy all enchantments you control.
var CalmingVerse = newCalmingVerse

func newCalmingVerse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Calming Verse",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerNotYou}),
						},
					},
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, Controller: game.ControllerYou}),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Tapped: game.TriFalse},
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all enchantments you don't control. Then if you control an untapped land, destroy all enchantments you control.
		`,
		},
	}
}
