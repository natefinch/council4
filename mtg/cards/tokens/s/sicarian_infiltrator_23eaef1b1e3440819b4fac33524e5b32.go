package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SicarianInfiltrator is the card definition for Sicarian Infiltrator.
//
// Type: Token Artifact Creature — Human Soldier
//
// Oracle text:
//   When Sicarian Infiltrator enters the battlefield, draw a card.

// SicarianInfiltratorToken23eaef1b1e3440819b4fac33524e5b32 is the card definition for Sicarian Infiltrator.
var SicarianInfiltratorToken23eaef1b1e3440819b4fac33524e5b32 = newSicarianInfiltratorToken23eaef1b1e3440819b4fac33524e5b32()

func newSicarianInfiltratorToken23eaef1b1e3440819b4fac33524e5b32() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Sicarian Infiltrator",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Sicarian Infiltrator enters the battlefield, draw a card.",
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
										Primitive: game.Draw{
											Amount: game.Fixed(1),
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
			When Sicarian Infiltrator enters the battlefield, draw a card.
		`,
		},
	}
}
