package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UrabraskTheHidden is the card definition for Urabrask the Hidden.
//
// Type: Legendary Creature — Phyrexian Praetor
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Creatures you control have haste.
//	Creatures your opponents control enter tapped.
var UrabraskTheHidden = newUrabraskTheHidden

func newUrabraskTheHidden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Urabrask the Hidden",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Praetor},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedGroupReplacement("Creatures your opponents control enter tapped.", game.TriggerControllerOpponent, types.Creature),
			},
			OracleText: `
			Creatures you control have haste.
			Creatures your opponents control enter tapped.
		`,
		},
	}
}
