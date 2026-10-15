package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EleshNornGrandCenobite is the card definition for Elesh Norn, Grand Cenobite.
//
// Type: Legendary Creature — Phyrexian Praetor
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Vigilance
//	Other creatures you control get +2/+2.
//	Creatures your opponents control get -2/-2.
var EleshNornGrandCenobite = newEleshNornGrandCenobite

func newEleshNornGrandCenobite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Elesh Norn, Grand Cenobite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Praetor},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.StaticAbility{
					Text: "Other creatures you control get +2/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
				game.StaticAbility{
					Text: "Creatures your opponents control get -2/-2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     -2,
							ToughnessDelta: -2,
						},
					},
				},
			},
			OracleText: `
			Vigilance
			Other creatures you control get +2/+2.
			Creatures your opponents control get -2/-2.
		`,
		},
	}
}
