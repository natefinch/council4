package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UlvenwaldTracker is the card definition for Ulvenwald Tracker.
//
// Type: Creature — Human Shaman
// Cost: {G}
//
// Oracle text:
//
//	{1}{G}, {T}: Target creature you control fights another target creature.
var UlvenwaldTracker = newUlvenwaldTracker

func newUlvenwaldTracker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ulvenwald Tracker",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{G}, {T}: Target creature you control fights another target creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.G,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
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
										Constraint: "Target creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									},
									game.TargetSpec{
										MinTargets:               1,
										MaxTargets:               1,
										Constraint:               "another target creature",
										Allow:                    game.TargetAllowPermanent,
										Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
										DistinctFromPriorTargets: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Fight{
											Object:        game.TargetPermanentReference(0),
											RelatedObject: game.TargetPermanentReference(1),
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
			{1}{G}, {T}: Target creature you control fights another target creature.
		`,
		},
	}
}
