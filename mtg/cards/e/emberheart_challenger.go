package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmberheartChallenger is the card definition for Emberheart Challenger.
//
// Type: Creature — Mouse Warrior
// Cost: {1}{R}
//
// Oracle text:
//
//	Haste
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
//	Valiant — Whenever this creature becomes the target of a spell or ability you control for the first time each turn, exile the top card of your library. Until end of turn, you may play that card.
var EmberheartChallenger = newEmberheartChallenger()

func newEmberheartChallenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Emberheart Challenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Mouse, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.ProwessStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventObjectBecameTarget,
							Source:          game.TriggerSourceSelf,
							CauseController: game.TriggerControllerYou,
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(1),
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Haste
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			Valiant — Whenever this creature becomes the target of a spell or ability you control for the first time each turn, exile the top card of your library. Until end of turn, you may play that card.
		`,
		},
	}
}
