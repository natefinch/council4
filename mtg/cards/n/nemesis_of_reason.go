package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NemesisOfReason is the card definition for Nemesis of Reason.
//
// Type: Creature — Leviathan Horror
// Cost: {3}{U}{B}
//
// Oracle text:
//
//	Whenever this creature attacks, defending player mills ten cards.
var NemesisOfReason = newNemesisOfReason()

func newNemesisOfReason() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Nemesis of Reason",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leviathan, types.Horror},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 7}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Fixed(10),
									Player: game.DefendingPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, defending player mills ten cards.
		`,
		},
	}
}
