package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RimescaleDragon is the card definition for Rimescale Dragon.
//
// Type: Snow Creature — Dragon
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	Flying
//	{2}{S}: Tap target creature and put an ice counter on it. ({S} can be paid with one mana from a snow source.)
//	Creatures with ice counters on them don't untap during their controllers' untap steps.
var RimescaleDragon = newRimescaleDragon()

func newRimescaleDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rimescale Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dragon},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectDoesntUntap,
							PermanentTypes:    []types.Card{types.Creature},
							AffectedSelection: game.Selection{MatchCounter: true, RequiredCounter: counter.Ice},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{S}: Tap target creature and put an ice counter on it. ({S} can be paid with one mana from a snow source.)",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.S}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
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
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.Ice,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			{2}{S}: Tap target creature and put an ice counter on it. ({S} can be paid with one mana from a snow source.)
			Creatures with ice counters on them don't untap during their controllers' untap steps.
		`,
		},
	}
}
