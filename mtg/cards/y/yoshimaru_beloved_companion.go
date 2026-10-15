package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// YoshimaruBelovedCompanion is the card definition for Yoshimaru, Beloved Companion.
//
// Type: Legendary Creature — Dog
// Cost: {2}{W}
//
// Oracle text:
//
//	If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.
//	{6}: Put a +1/+1 counter on target legendary creature.
var YoshimaruBelovedCompanion = newYoshimaruBelovedCompanion

func newYoshimaruBelovedCompanion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Yoshimaru, Beloved Companion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dog},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{6}: Put a +1/+1 counter on target legendary creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(6),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target legendary creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.CounterPlacementReplacement("If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.", 0, 1, counter.PlusOnePlusOne, game.TriggerControllerYou),
			},
			OracleText: `
			If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on it instead.
			{6}: Put a +1/+1 counter on target legendary creature.
		`,
		},
	}
}
