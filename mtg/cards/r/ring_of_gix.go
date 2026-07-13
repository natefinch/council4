package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RingOfGix is the card definition for Ring of Gix.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Echo {3} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	{1}, {T}: Tap target artifact, creature, or land.
var RingOfGix = newRingOfGix

func newRingOfGix() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ring of Gix",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{1}, {T}: Tap target artifact, creature, or land.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact, creature, or land",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			Echo {3} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			{1}, {T}: Tap target artifact, creature, or land.
		`,
		},
	}
}
