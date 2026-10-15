package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MeddlingYouths is the card definition for Meddling Youths.
//
// Type: Creature — Human Detective
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Haste
//	Whenever you attack with three or more creatures, investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
var MeddlingYouths = newMeddlingYouths

func newMeddlingYouths() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Meddling Youths",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Detective},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you attack with three or more creatures, investigate. (Create a Clue token. It's an artifact with \"{2}, Sacrifice this token: Draw a card.\")",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventAttackerDeclared,
							Controller:           game.TriggerControllerYou,
							StackObjectKind:      game.StackSpell,
							AttackerCountAtLeast: 3,
							OneOrMore:            true,
							CounterKind:          counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Investigate{
											Amount: game.Fixed(1),
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
			Haste
			Whenever you attack with three or more creatures, investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
		`,
		},
	}
}
