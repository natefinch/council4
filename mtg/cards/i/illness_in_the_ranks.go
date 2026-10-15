package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IllnessInTheRanks is the card definition for Illness in the Ranks.
//
// Type: Enchantment
// Cost: {B}
//
// Oracle text:
//
//	Creature tokens get -1/-1.
var IllnessInTheRanks = newIllnessInTheRanks

func newIllnessInTheRanks() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Illness in the Ranks",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature tokens get -1/-1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     -1,
							ToughnessDelta: -1,
						},
					},
				},
			},
			OracleText: `
			Creature tokens get -1/-1.
		`,
		},
	}
}
