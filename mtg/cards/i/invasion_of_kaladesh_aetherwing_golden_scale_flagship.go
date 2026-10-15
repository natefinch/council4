package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InvasionOfKaladesh is the card definition for Invasion of Kaladesh // Aetherwing, Golden-Scale Flagship.
//
// Type: Battle — Siege // Legendary Artifact — Vehicle
// Face: Aetherwing, Golden-Scale Flagship — Legendary Artifact — Vehicle
//
// Oracle text:
//
//	(As a Siege enters, choose an opponent to protect it. You and others can attack it. When it's defeated, exile it, then cast it transformed.)
//	When this Siege enters, create a 1/1 colorless Thopter artifact creature token with flying.
var InvasionOfKaladesh = newInvasionOfKaladesh

func newInvasionOfKaladesh() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Invasion of Kaladesh",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.R,
			}),
			Colors:   []color.Color{color.Red, color.Blue},
			Types:    []types.Card{types.Battle},
			Subtypes: []types.Sub{types.Siege},
			Defense:  opt.Val(4),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this Siege enters, create a 1/1 colorless Thopter artifact creature token with flying.",
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
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(invasionOfKaladeshToken),
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
			(As a Siege enters, choose an opponent to protect it. You and others can attack it. When it's defeated, exile it, then cast it transformed.)
			When this Siege enters, create a 1/1 colorless Thopter artifact creature token with flying.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:         "Aetherwing, Golden-Scale Flagship",
			Colors:       []color.Color{color.Red, color.Blue},
			Supertypes:   []types.Super{types.Legendary},
			Types:        []types.Card{types.Artifact},
			Subtypes:     []types.Sub{types.Vehicle},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerArtifactCount}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(1),
			},
			OracleText: `
			Flying
			Aetherwing's power is equal to the number of artifacts you control.
			Crew 1 (Tap any number of creatures you control with total power 1 or more: This Vehicle becomes an artifact creature until end of turn.)
		`,
		}),
	}
}

var invasionOfKaladeshToken = newInvasionOfKaladeshToken()

func newInvasionOfKaladeshToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Thopter",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Thopter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
