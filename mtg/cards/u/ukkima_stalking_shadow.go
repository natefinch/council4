package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UkkimaStalkingShadow is the card definition for Ukkima, Stalking Shadow.
//
// Type: Legendary Creature — Whale Wolf
// Cost: {1}{U}{B}
//
// Oracle text:
//
//	Partner with Cazur, Ruthless Stalker (When this creature enters, target player may put Cazur into their hand from their library, then shuffle.)
//	Ukkima can't be blocked.
//	When Ukkima leaves the battlefield, it deals X damage to target player and you gain X life, where X is its power.
var UkkimaStalkingShadow = newUkkimaStalkingShadow

func newUkkimaStalkingShadow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Ukkima, Stalking Shadow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Whale, types.Wolf},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
				game.CantBeBlockedStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Ukkima leaves the battlefield, it deals X damage to target player and you gain X life, where X is its power.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventZoneChanged,
							Source:          game.TriggerSourceSelf,
							MatchFromZone:   true,
							FromZone:        zone.Battlefield,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target player",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectPower,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
											}),
											Recipient:    game.AnyTargetDamageRecipient(0),
											DamageSource: opt.Val(game.EventPermanentReference()),
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectPower,
												Multiplier: 1,
												Object:     game.EventPermanentReference(),
											}),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.PartnerWithTriggeredAbility("Cazur, Ruthless Stalker"),
			},
			OracleText: `
			Partner with Cazur, Ruthless Stalker (When this creature enters, target player may put Cazur into their hand from their library, then shuffle.)
			Ukkima can't be blocked.
			When Ukkima leaves the battlefield, it deals X damage to target player and you gain X life, where X is its power.
		`,
		},
	}
}
