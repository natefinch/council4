package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CorrosiveMentor is the card definition for Corrosive Mentor.
//
// Type: Creature — Elemental Rogue
// Cost: {2}{B}
//
// Oracle text:
//
//	Black creatures you control have wither. (They deal damage to creatures in the form of -1/-1 counters.)
var CorrosiveMentor = newCorrosiveMentor

func newCorrosiveMentor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Corrosive Mentor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Black creatures you control have wither. (They deal damage to creatures in the form of -1/-1 counters.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Black}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Wither},
						},
					},
				},
			},
			OracleText: `
			Black creatures you control have wither. (They deal damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
