package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UlrichSKindred is the card definition for Ulrich's Kindred.
//
// Type: Creature — Wolf
// Cost: {2}{R}
//
// Oracle text:
//
//	Trample
//	{3}{G}: Target attacking Wolf or Werewolf gains indestructible until end of turn. (Damage and effects that say "destroy" don't destroy it.)
var UlrichSKindred = newUlrichSKindred

func newUlrichSKindred() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Ulrich's Kindred",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{G}: Target attacking Wolf or Werewolf gains indestructible until end of turn. (Damage and effects that say \"destroy\" don't destroy it.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
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
										Constraint: "target attacking Wolf or Werewolf",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Wolf, types.Werewolf}, CombatState: game.CombatStateAttacking}),
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
													AddKeywords: []game.Keyword{game.Indestructible},
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
			Trample
			{3}{G}: Target attacking Wolf or Werewolf gains indestructible until end of turn. (Damage and effects that say "destroy" don't destroy it.)
		`,
		},
	}
}
