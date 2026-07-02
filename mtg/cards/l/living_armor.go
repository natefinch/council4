package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LivingArmor is the card definition for Living Armor.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{T}, Sacrifice this artifact: Put X +0/+1 counters on target creature, where X is that creature's mana value.
var LivingArmor = newLivingArmor()

func newLivingArmor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Living Armor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice this artifact: Put X +0/+1 counters on target creature, where X is that creature's mana value.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this artifact",
							Amount: 1,
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
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectManaValue,
										Multiplier: 1,
										Object:     game.TargetPermanentReference(0),
									}),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusZeroPlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}, Sacrifice this artifact: Put X +0/+1 counters on target creature, where X is that creature's mana value.
		`,
		},
	}
}
