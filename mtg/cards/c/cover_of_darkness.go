package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CoverOfDarkness is the card definition for Cover of Darkness.
//
// Type: Enchantment
// Cost: {1}{B}
//
// Oracle text:
//
//	As this enchantment enters, choose a creature type.
//	Creatures of the chosen type have fear. (They can't be blocked except by artifact creatures and/or black creatures.)
var CoverOfDarkness = newCoverOfDarkness()

func newCoverOfDarkness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Cover of Darkness",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							AddKeywords: []game.Keyword{
								game.Fear,
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
			Creatures of the chosen type have fear. (They can't be blocked except by artifact creatures and/or black creatures.)
		`,
		},
	}
}
