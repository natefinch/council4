package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Extruder is the card definition for Extruder.
//
// Type: Artifact Creature — Juggernaut
// Cost: {4}
//
// Oracle text:
//
//	Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	Sacrifice an artifact: Put a +1/+1 counter on target creature.
var Extruder = newExtruder

func newExtruder() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Extruder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Juggernaut},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice an artifact: Put a +1/+1 counter on target creature.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice an artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
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
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(4)}),
			},
			OracleText: `
			Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			Sacrifice an artifact: Put a +1/+1 counter on target creature.
		`,
		},
	}
}
