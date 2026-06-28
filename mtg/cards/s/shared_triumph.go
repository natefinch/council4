package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SharedTriumph is the card definition for Shared Triumph.
//
// Type: Enchantment
// Cost: {1}{W}
//
// Oracle text:
//
//	As this enchantment enters, choose a creature type.
//	Creatures of the chosen type get +1/+1.
var SharedTriumph = newSharedTriumph()

func newSharedTriumph() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Shared Triumph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this enchantment enters, choose a creature type."),
			},
			OracleText: `
			As this enchantment enters, choose a creature type.
			Creatures of the chosen type get +1/+1.
		`,
		},
	}
}
