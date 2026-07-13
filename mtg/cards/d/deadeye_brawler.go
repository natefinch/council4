package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeadeyeBrawler is the card definition for Deadeye Brawler.
//
// Type: Creature — Human Pirate
// Cost: {2}{U}{B}
//
// Oracle text:
//
//	Deathtouch
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	Whenever this creature deals combat damage to a player, if you have the city's blessing, draw a card.
var DeadeyeBrawler = newDeadeyeBrawler

func newDeadeyeBrawler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Deadeye Brawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Pirate},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.AscendStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
						InterveningIf: "if you have the city's blessing",
						InterveningCondition: opt.Val(game.Condition{
							ControllerHasCityBlessing: true,
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Deathtouch
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			Whenever this creature deals combat damage to a player, if you have the city's blessing, draw a card.
		`,
		},
	}
}
