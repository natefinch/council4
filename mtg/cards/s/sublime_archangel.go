package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SublimeArchangel is the card definition for Sublime Archangel.
//
// Type: Creature — Angel
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Flying
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
//	Other creatures you control have exalted. (If a creature has multiple instances of exalted, each triggers separately.)
var SublimeArchangel = newSublimeArchangel

func newSublimeArchangel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sublime Archangel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ExaltedStaticBody,
				game.StaticAbility{
					Text: "Other creatures you control have exalted. (If a creature has multiple instances of exalted, each triggers separately.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Exalted},
						},
					},
				},
			},
			OracleText: `
			Flying
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
			Other creatures you control have exalted. (If a creature has multiple instances of exalted, each triggers separately.)
		`,
		},
	}
}
