package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AetherShockwave is the card definition for Aether Shockwave.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Choose one —
//	• Tap all Spirits.
//	• Tap all non-Spirit creatures.
var AetherShockwave = newAetherShockwave

func newAetherShockwave() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Aether Shockwave",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Tap all Spirits.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Spirit")}}),
								},
							},
						},
					},
					game.Mode{
						Text: "Tap all non-Spirit creatures.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Spirit")}),
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
			• Tap all Spirits.
			• Tap all non-Spirit creatures.
		`,
		},
	}
}
