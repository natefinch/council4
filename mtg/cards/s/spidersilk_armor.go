package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpidersilkArmor is the card definition for Spidersilk Armor.
//
// Type: Enchantment
// Cost: {2}{G}
//
// Oracle text:
//
//	Creatures you control get +0/+1 and have reach. (They can block creatures with flying.)
var SpidersilkArmor = newSpidersilkArmor

func newSpidersilkArmor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Spidersilk Armor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control get +0/+1 and have reach. (They can block creatures with flying.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Reach},
						},
					},
				},
			},
			OracleText: `
			Creatures you control get +0/+1 and have reach. (They can block creatures with flying.)
		`,
		},
	}
}
