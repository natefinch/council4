package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AccursedHorde is the card definition for Accursed Horde.
//
// Type: Creature — Zombie
// Cost: {3}{B}
//
// Oracle text:
//
//	{1}{B}: Target attacking Zombie gains indestructible until end of turn. (Damage and effects that say "destroy" don't destroy it. If its toughness is 0 or less, it still dies.)
var AccursedHorde = newAccursedHorde

func newAccursedHorde() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Accursed Horde",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{B}: Target attacking Zombie gains indestructible until end of turn. (Damage and effects that say \"destroy\" don't destroy it. If its toughness is 0 or less, it still dies.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.B,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target attacking Zombie",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Zombie}, CombatState: game.CombatStateAttacking}),
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
			{1}{B}: Target attacking Zombie gains indestructible until end of turn. (Damage and effects that say "destroy" don't destroy it. If its toughness is 0 or less, it still dies.)
		`,
		},
	}
}
