package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RuneOfProtectionRed is the card definition for Rune of Protection: Red.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	{W}: The next time a red source of your choice would deal damage to you this turn, prevent that damage.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var RuneOfProtectionRed = newRuneOfProtectionRed

func newRuneOfProtectionRed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rune of Protection: Red",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{W}: The next time a red source of your choice would deal damage to you this turn, prevent that damage.",
					ManaCost:       opt.Val(cost.Mana{cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Player:       game.ControllerReference(),
									All:          true,
									OneShot:      true,
									SourceColors: []color.Color{color.Red},
								},
							},
						},
					}.Ability(),
				},
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			{W}: The next time a red source of your choice would deal damage to you this turn, prevent that damage.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
