package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmberwildeCaliph is the card definition for Emberwilde Caliph.
//
// Type: Creature — Djinn
// Cost: {2}{U}{R}
//
// Oracle text:
//
//	Flying, trample
//	This creature attacks each combat if able.
//	Whenever this creature deals damage, you lose that much life.
var EmberwildeCaliph = newEmberwildeCaliph()

func newEmberwildeCaliph() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Emberwilde Caliph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.R,
			}),
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Djinn},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
				game.MustAttackStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:   game.EventDamageDealt,
							Source:  game.TriggerSourceSelf,
							Subject: game.TriggerSubjectDamageSource,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventDamage,
										Multiplier: 1,
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, trample
			This creature attacks each combat if able.
			Whenever this creature deals damage, you lose that much life.
		`,
		},
	}
}
