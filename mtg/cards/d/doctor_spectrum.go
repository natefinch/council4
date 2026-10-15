package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DoctorSpectrum is the card definition for Doctor Spectrum.
//
// Type: Legendary Creature — Human Hero
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flying
//	When Doctor Spectrum enters, choose one —
//	• Create a 0/4 colorless Wall creature token with defender.
//	• Put a +1/+1 counter on each other Hero you control.
//	• Destroy target enchantment.
var DoctorSpectrum = newDoctorSpectrum

func newDoctorSpectrum() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Doctor Spectrum",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Doctor Spectrum enters, choose one —\n• Create a 0/4 colorless Wall creature token with defender.\n• Put a +1/+1 counter on each other Hero you control.\n• Destroy target enchantment.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Create a 0/4 colorless Wall creature token with defender.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(doctorSpectrumToken),
										},
									},
								},
							},
							game.Mode{
								Text: "Put a +1/+1 counter on each other Hero you control.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Group:       game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub("Hero")}, Controller: game.ControllerYou}, game.SourcePermanentReference()),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
							game.Mode{
								Text: "Destroy target enchantment.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target enchantment",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Enchantment}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
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
			Flying
			When Doctor Spectrum enters, choose one —
			• Create a 0/4 colorless Wall creature token with defender.
			• Put a +1/+1 counter on each other Hero you control.
			• Destroy target enchantment.
		`,
		},
	}
}

var doctorSpectrumToken = newDoctorSpectrumToken()

func newDoctorSpectrumToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wall",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
		},
	}
}
