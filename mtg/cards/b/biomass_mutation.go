package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BiomassMutation is the card definition for Biomass Mutation.
//
// Type: Instant
// Cost: {X}{G/U}{G/U}
//
// Oracle text:
//
//	Creatures you control have base power and toughness X/X until end of turn.
var BiomassMutation = newBiomassMutation

func newBiomassMutation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Biomass Mutation",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors: []color.Color{color.Green, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration: game.DurationPermanent,
									Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:    game.LayerPowerToughnessSet,
									SetPowerDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountX,
										CounterKind: counter.PlusOnePlusOne,
									}),
									SetToughnessDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountX,
										CounterKind: counter.PlusOnePlusOne,
									}),
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control have base power and toughness X/X until end of turn.
		`,
		},
	}
}
