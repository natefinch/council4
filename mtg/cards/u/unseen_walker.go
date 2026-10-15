package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UnseenWalker is the card definition for Unseen Walker.
//
// Type: Creature — Dryad
// Cost: {1}{G}
//
// Oracle text:
//
//	Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
//	{1}{G}{G}: Target creature gains forestwalk until end of turn.
var UnseenWalker = newUnseenWalker

func newUnseenWalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Unseen Walker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dryad},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ForestwalkStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{G}{G}: Target creature gains forestwalk until end of turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.G,
						cost.G,
					}),
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
													Duration: game.DurationPermanent,
													Layer:    game.LayerAbility,
													AddAbilities: []game.Ability{
														&game.StaticAbility{
															Text: "Forestwalk",
															KeywordAbilities: []game.KeywordAbility{
																game.LandwalkKeyword{
																	Subtype: types.Forest,
																},
															},
														},
													},
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
			Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
			{1}{G}{G}: Target creature gains forestwalk until end of turn.
		`,
		},
	}
}
