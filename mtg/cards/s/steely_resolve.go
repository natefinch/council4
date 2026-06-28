package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SteelyResolve is the card definition for Steely Resolve.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	As this enchantment enters, choose a creature type.
//	Creatures of the chosen type have shroud. (They can't be the targets of spells or abilities.)
var SteelyResolve = newSteelyResolve()

func newSteelyResolve() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Steely Resolve",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							AddKeywords: []game.Keyword{
								game.Shroud,
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this enchantment enters, choose a creature type."),
			},
			OracleText: `
			As this enchantment enters, choose a creature type.
			Creatures of the chosen type have shroud. (They can't be the targets of spells or abilities.)
		`,
		},
	}
}
