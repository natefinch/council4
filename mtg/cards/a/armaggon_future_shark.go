package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArmaggonFutureShark is the card definition for Armaggon, Future Shark.
//
// Type: Legendary Creature — Shark Horror Mutant
// Cost: {6}{B}{B}
//
// Oracle text:
//
//	Flash
//	When Armaggon enters, destroy up to three target creatures.
var ArmaggonFutureShark = newArmaggonFutureShark

func newArmaggonFutureShark() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Armaggon, Future Shark",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Shark, types.Horror, types.Mutant},
			Power:      opt.Val(game.PT{Value: 9}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Armaggon enters, destroy up to three target creatures.",
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 3,
										Constraint: "up to three target creatures",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(1),
										},
									},
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(2),
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
			Flash
			When Armaggon enters, destroy up to three target creatures.
		`,
		},
	}
}
