package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrideGuardian is the card definition for Pride Guardian.
//
// Type: Creature — Cat Monk
// Cost: {W}
//
// Oracle text:
//
//	Defender
//	Whenever this creature blocks, you gain 3 life.
var PrideGuardian = newPrideGuardian

func newPrideGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Pride Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Monk},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature blocks, you gain 3 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventBlockerDeclared,
							BlockerScope:    game.BlockerTriggerOnce,
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
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
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
			},
			OracleText: `
			Defender
			Whenever this creature blocks, you gain 3 life.
		`,
		},
	}
}
