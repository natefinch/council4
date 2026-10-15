package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VirulentPlague is the card definition for Virulent Plague.
//
// Type: Enchantment
// Cost: {2}{B}
//
// Oracle text:
//
//	Creature tokens get -2/-2.
var VirulentPlague = newVirulentPlague

func newVirulentPlague() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Virulent Plague",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature tokens get -2/-2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     -2,
							ToughnessDelta: -2,
						},
					},
				},
			},
			OracleText: `
			Creature tokens get -2/-2.
		`,
		},
	}
}
