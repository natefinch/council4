package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VineshaperProdigy is the card definition for Vineshaper Prodigy.
//
// Type: Creature — Elf Druid
// Cost: {1}{G}
//
// Oracle text:
//
//	Kicker {1}{U} (You may pay an additional {1}{U} as you cast this spell.)
//	When this creature enters, if it was kicked, look at the top three cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.
var VineshaperProdigy = newVineshaperProdigy()

func newVineshaperProdigy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Vineshaper Prodigy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Druid},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(1), cost.U}},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf:                        "if it was kicked",
						InterveningIfEventPermanentWasKicked: true,
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(3),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Kicker {1}{U} (You may pay an additional {1}{U} as you cast this spell.)
			When this creature enters, if it was kicked, look at the top three cards of your library. Put one of them into your hand and the rest on the bottom of your library in any order.
		`,
		},
	}
}
