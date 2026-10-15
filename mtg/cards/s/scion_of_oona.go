package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScionOfOona is the card definition for Scion of Oona.
//
// Type: Creature — Faerie Soldier
// Cost: {2}{U}
//
// Oracle text:
//
//	Flash
//	Flying
//	Other Faerie creatures you control get +1/+1.
//	Other Faeries you control have shroud. (They can't be the targets of spells or abilities.)
var ScionOfOona = newScionOfOona

func newScionOfOona() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Scion of Oona",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Other Faerie creatures you control get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Faerie")}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
				game.StaticAbility{
					Text: "Other Faeries you control have shroud. (They can't be the targets of spells or abilities.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Faerie")}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Shroud},
						},
					},
				},
			},
			OracleText: `
			Flash
			Flying
			Other Faerie creatures you control get +1/+1.
			Other Faeries you control have shroud. (They can't be the targets of spells or abilities.)
		`,
		},
	}
}
