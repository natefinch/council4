package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MouthOfRonom is the card definition for Mouth of Ronom.
//
// Type: Snow Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{4}{S}, {T}, Sacrifice this land: It deals 4 damage to target creature. ({S} can be paid with one mana from a snow source.)
var MouthOfRonom = newMouthOfRonom

func newMouthOfRonom() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Mouth of Ronom",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}{S}, {T}, Sacrifice this land: It deals 4 damage to target creature. ({S} can be paid with one mana from a snow source.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.S,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this land",
							Amount:      1,
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
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(4),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{4}{S}, {T}, Sacrifice this land: It deals 4 damage to target creature. ({S} can be paid with one mana from a snow source.)
		`,
		},
	}
}
