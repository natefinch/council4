package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MyojinOfRoaringBlades is the card definition for Myojin of Roaring Blades.
//
// Type: Legendary Creature — Spirit
// Cost: {5}{R}{R}{R}
//
// Oracle text:
//
//	Myojin of Roaring Blades enters with an indestructible counter on it if you cast it from your hand.
//	Remove an indestructible counter from Myojin of Roaring Blades: It deals 7 damage to each of up to three targets.
var MyojinOfRoaringBlades = newMyojinOfRoaringBlades

func newMyojinOfRoaringBlades() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Myojin of Roaring Blades",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove an indestructible counter from Myojin of Roaring Blades: It deals 7 damage to each of up to three targets.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove an indestructible counter from Myojin of Roaring Blades",
							Amount:      1,
							CounterKind: counter.Indestructible,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 3,
										Constraint: "up to three targets",
										Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(7),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(7),
											Recipient:    game.AnyTargetDamageRecipient(1),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(7),
											Recipient:    game.AnyTargetDamageRecipient(2),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Myojin of Roaring Blades enters with an indestructible counter on it if you cast it from your hand.", &game.Condition{
					EventPermanentWasCastFromControllerHand: true,
				}, game.CounterPlacement{Kind: counter.Indestructible, Amount: 1}),
			},
			OracleText: `
			Myojin of Roaring Blades enters with an indestructible counter on it if you cast it from your hand.
			Remove an indestructible counter from Myojin of Roaring Blades: It deals 7 damage to each of up to three targets.
		`,
		},
	}
}
