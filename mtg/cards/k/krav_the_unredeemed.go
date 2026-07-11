package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KravTheUnredeemed is the card definition for Krav, the Unredeemed.
//
// Type: Legendary Creature — Demon
// Cost: {4}{B}
//
// Oracle text:
//
//	Partner with Regna, the Redeemer (When this creature enters, target player may put Regna into their hand from their library, then shuffle.)
//	{B}, Sacrifice X creatures: Target player draws X cards and gains X life. Put X +1/+1 counters on Krav.
var KravTheUnredeemed = newKravTheUnredeemed

func newKravTheUnredeemed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Krav, the Unredeemed",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Demon},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{B}, Sacrifice X creatures: Target player draws X cards and gains X life. Put X +1/+1 counters on Krav.",
					ManaCost: opt.Val(cost.Mana{cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice X creatures",
							AmountFromX:        true,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Partner with Regna, the Redeemer (When this creature enters, target player may put Regna into their hand from their library, then shuffle.)
			{B}, Sacrifice X creatures: Target player draws X cards and gains X life. Put X +1/+1 counters on Krav.
		`,
		},
	}
}
