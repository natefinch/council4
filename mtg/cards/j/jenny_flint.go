package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JennyFlint is the card definition for Jenny Flint.
//
// Type: Legendary Creature — Human Detective
// Cost: {1}{U}{R}
//
// Oracle text:
//
//	Partner with Madame Vastra
//	First strike
//	Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
//	Whenever you sacrifice a Clue or Food, put a +1/+1 counter on another target creature you control.
var JennyFlint = newJennyFlint

func newJennyFlint() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Jenny Flint",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.R,
			}),
			Colors:     []color.Color{color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Detective},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
				game.FirstStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TrainingTriggeredBody,
				game.TriggeredAbility{
					Text: "Whenever you sacrifice a Clue or Food, put a +1/+1 counter on another target creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentSacrificed,
							Player:           game.TriggerPlayerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Clue, types.Food}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "another target creature you control",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludeSource: true, Controller: game.ControllerYou}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.PartnerWithTriggeredAbility("Madame Vastra"),
			},
			OracleText: `
			Partner with Madame Vastra
			First strike
			Training (Whenever this creature attacks with another creature with greater power, put a +1/+1 counter on this creature.)
			Whenever you sacrifice a Clue or Food, put a +1/+1 counter on another target creature you control.
		`,
		},
	}
}
