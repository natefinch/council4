package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FinaleOfDevastation is the card definition for Finale of Devastation.
//
// Type: Sorcery
// Cost: {X}{G}{G}
//
// Oracle text:
//
//	Search your library and/or graveyard for a creature card with mana value X or less and put it onto the battlefield. If you search your library this way, shuffle. If X is 10 or more, creatures you control get +X/+X and gain haste until end of turn.
var FinaleOfDevastation = newFinaleOfDevastation

func newFinaleOfDevastation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Finale of Devastation",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:         zone.Library,
								Destination:        zone.Battlefield,
								Filter:             game.Selection{RequiredTypes: []types.Card{types.Creature}},
								MaxManaValueFromX:  true,
								AlsoGraveyard:      true,
								ConditionalShuffle: true,
							},
							Amount: game.Fixed(1),
						},
						PublishResult: game.ResultKey("multizone-search-library"),
					},
					{
						Primitive: game.ShuffleLibrary{
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:             "multizone-search-library",
							SearchedLibrary: game.TriTrue,
						}),
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerPowerToughnessModify,
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									PowerDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:       game.DynamicAmountX,
										Multiplier: 1,
									}),
									ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:       game.DynamicAmountX,
										Multiplier: 1,
									}),
								},
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Haste,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateSpellX, Op: compare.GreaterOrEqual, Value: 10}},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Search your library and/or graveyard for a creature card with mana value X or less and put it onto the battlefield. If you search your library this way, shuffle. If X is 10 or more, creatures you control get +X/+X and gain haste until end of turn.
		`,
		},
	}
}
