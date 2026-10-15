package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HiddenPath is the card definition for Hidden Path.
//
// Type: Enchantment
// Cost: {2}{G}{G}{G}{G}
//
// Oracle text:
//
//	Green creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)
var HiddenPath = newHiddenPath

func newHiddenPath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hidden Path",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Green creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Forestwalk",
									KeywordAbilities: []game.KeywordAbility{
										game.LandwalkKeyword{
											Subtype: types.Forest,
										},
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Green creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)
		`,
		},
	}
}
