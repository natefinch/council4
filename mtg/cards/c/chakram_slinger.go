package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChakramSlinger is the card definition for Chakram Slinger.
//
// Type: Creature — Human Warrior
// Cost: {4}{R}
//
// Oracle text:
//
//	Partner with Chakram Retriever (When this creature enters, target player may put Chakram Retriever into their hand from their library, then shuffle.)
//	{R}, {T}: This creature deals 2 damage to target player or planeswalker.
var ChakramSlinger = newChakramSlinger

func newChakramSlinger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Chakram Slinger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.PartnerWithTriggeredAbility("Chakram Retriever"),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{R}, {T}: This creature deals 2 damage to target player or planeswalker.",
					ManaCost: opt.Val(cost.Mana{
						cost.R,
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
										Constraint: "target player or planeswalker",
										Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(2),
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
			OracleText: `
			Partner with Chakram Retriever (When this creature enters, target player may put Chakram Retriever into their hand from their library, then shuffle.)
			{R}, {T}: This creature deals 2 damage to target player or planeswalker.
		`,
		},
	}
}
