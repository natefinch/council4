package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EngineeredPlague is the card definition for Engineered Plague.
//
// Type: Enchantment
// Cost: {2}{B}
//
// Oracle text:
//
//	As this enchantment enters, choose a creature type.
//	All creatures of the chosen type get -1/-1.
var EngineeredPlague = newEngineeredPlague()

func newEngineeredPlague() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Engineered Plague",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							PowerDelta:     -1,
							ToughnessDelta: -1,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this enchantment enters, choose a creature type."),
			},
			OracleText: `
			As this enchantment enters, choose a creature type.
			All creatures of the chosen type get -1/-1.
		`,
		},
	}
}
