package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrompTheDomains is the card definition for Tromp the Domains.
//
// Type: Sorcery
// Cost: {5}{G}
//
// Oracle text:
//
//	Domain — Until end of turn, creatures you control gain trample and get +1/+1 for each basic land type among lands you control.
var TrompTheDomains = newTrompTheDomains

func newTrompTheDomains() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tromp the Domains",
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
										Kind:        game.DynamicAmountControllerBasicLandTypeCount,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
									}),
									ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountControllerBasicLandTypeCount,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
									}),
								},
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Trample},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Domain — Until end of turn, creatures you control gain trample and get +1/+1 for each basic land type among lands you control.
		`,
		},
	}
}
