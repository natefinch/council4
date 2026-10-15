package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LuXunScholarGeneral is the card definition for Lu Xun, Scholar General.
//
// Type: Legendary Creature — Human Soldier
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
//	Whenever Lu Xun deals damage to an opponent, you may draw a card.
var LuXunScholarGeneral = newLuXunScholarGeneral

func newLuXunScholarGeneral() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Lu Xun, Scholar General",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HorsemanshipStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Lu Xun deals damage to an opponent, you may draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventDamageDealt,
							Source:          game.TriggerSourceSelf,
							Player:          game.TriggerPlayerOpponent,
							Subject:         game.TriggerSubjectDamageSource,
							StackObjectKind: game.StackSpell,
							DamageRecipient: game.DamageRecipientPlayer,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
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
			Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
			Whenever Lu Xun deals damage to an opponent, you may draw a card.
		`,
		},
	}
}
