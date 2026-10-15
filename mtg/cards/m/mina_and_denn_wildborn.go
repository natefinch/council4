package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MinaAndDennWildborn is the card definition for Mina and Denn, Wildborn.
//
// Type: Legendary Creature — Elf Ally
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	You may play an additional land on each of your turns.
//	{R}{G}, Return a land you control to its owner's hand: Target creature gains trample until end of turn.
var MinaAndDennWildborn = newMinaAndDennWildborn

func newMinaAndDennWildborn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Mina and Denn, Wildborn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Ally},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "You may play an additional land on each of your turns.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                game.RuleEffectAdditionalLandPlays,
							Duration:            game.DurationPermanent,
							AffectedPlayer:      game.PlayerYou,
							AdditionalLandPlays: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{R}{G}, Return a land you control to its owner's hand: Target creature gains trample until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.R,
						cost.G,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalReturnToHand,
							Text:               "Return a land you control to its owner's hand",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Land,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:    game.DurationPermanent,
													Layer:       game.LayerAbility,
													AddKeywords: []game.Keyword{game.Trample},
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			You may play an additional land on each of your turns.
			{R}{G}, Return a land you control to its owner's hand: Target creature gains trample until end of turn.
		`,
		},
	}
}
