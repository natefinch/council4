package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CircleOfProtectionWhite is the card definition for Circle of Protection: White.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	{1}: The next time a white source of your choice would deal damage to you this turn, prevent that damage.
var CircleOfProtectionWhite = newCircleOfProtectionWhite

func newCircleOfProtectionWhite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Circle of Protection: White",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}: The next time a white source of your choice would deal damage to you this turn, prevent that damage.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player:       game.ControllerReference(),
									All:          true,
									OneShot:      true,
									SourceColors: []color.Color{color.White},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{1}: The next time a white source of your choice would deal damage to you this turn, prevent that damage.
		`,
		},
	}
}
