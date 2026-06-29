package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CyclopeanSnare is the card definition for Cyclopean Snare.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{3}, {T}: Tap target creature, then return this artifact to its owner's hand.
var CyclopeanSnare = newCyclopeanSnare()

func newCyclopeanSnare() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Cyclopean Snare",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{3}, {T}: Tap target creature, then return this artifact to its owner's hand.",
					ManaCost:        opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
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
								Primitive: game.Bounce{
									Object: game.SourcePermanentReference(),
								},
							},
							{
								Primitive: game.Bounce{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{3}, {T}: Tap target creature, then return this artifact to its owner's hand.
		`,
		},
	}
}
