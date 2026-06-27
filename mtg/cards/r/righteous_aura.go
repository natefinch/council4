package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RighteousAura is the card definition for Righteous Aura.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	{W}, Pay 2 life: The next time a source of your choice would deal damage to you this turn, prevent that damage.
var RighteousAura = newRighteousAura()

func newRighteousAura() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Righteous Aura",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{W}, Pay 2 life: The next time a source of your choice would deal damage to you this turn, prevent that damage.",
					ManaCost: opt.Val(cost.Mana{cost.W}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Text:   "Pay 2 life",
							Amount: 2,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player:  game.ControllerReference(),
									All:     true,
									OneShot: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{W}, Pay 2 life: The next time a source of your choice would deal damage to you this turn, prevent that damage.
		`,
		},
	}
}
