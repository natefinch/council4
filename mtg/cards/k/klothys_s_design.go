package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KlothysSDesign is the card definition for Klothys's Design.
//
// Type: Sorcery
// Cost: {5}{G}
//
// Oracle text:
//
//	Creatures you control get +X/+X until end of turn, where X is your devotion to green. (Each {G} in the mana costs of permanents you control counts toward your devotion to green.)
var KlothysSDesign = newKlothysSDesign

func newKlothysSDesign() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Klothys's Design",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration: game.DurationPermanent,
									Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:    game.LayerPowerToughnessModify,
									PowerDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountDevotion,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
										Colors:      []color.Color{color.Green},
									}),
									ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountDevotion,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
										Colors:      []color.Color{color.Green},
									}),
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control get +X/+X until end of turn, where X is your devotion to green. (Each {G} in the mana costs of permanents you control counts toward your devotion to green.)
		`,
		},
	}
}
