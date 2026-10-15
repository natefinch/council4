package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GatekeeperOfMalakir is the card definition for Gatekeeper of Malakir.
//
// Type: Creature — Vampire Warrior
// Cost: {B}{B}
//
// Oracle text:
//
//	Kicker {B} (You may pay an additional {B} as you cast this spell.)
//	When this creature enters, if it was kicked, target player sacrifices a creature of their choice.
var GatekeeperOfMalakir = newGatekeeperOfMalakir

func newGatekeeperOfMalakir() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gatekeeper of Malakir",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker {B}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.B,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, if it was kicked, target player sacrifices a creature of their choice.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf:                        "if it was kicked",
						InterveningIfEventPermanentWasKicked: true,
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
										Primitive: game.SacrificePermanents{
											Player:    game.TargetPlayerReference(0),
											Amount:    game.Fixed(1),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
			Kicker {B} (You may pay an additional {B} as you cast this spell.)
			When this creature enters, if it was kicked, target player sacrifices a creature of their choice.
		`,
		},
	}
}
