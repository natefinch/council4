package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OverwhelmedApprentice is the card definition for Overwhelmed Apprentice.
//
// Type: Creature — Human Wizard
// Cost: {U}
//
// Oracle text:
//
//	When this creature enters, each opponent mills two cards. Then you scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)
var OverwhelmedApprentice = newOverwhelmedApprentice

func newOverwhelmedApprentice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Overwhelmed Apprentice",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount:      game.Fixed(2),
									PlayerGroup: game.OpponentsReference(),
								},
							},
							{
								Primitive: game.Scry{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, each opponent mills two cards. Then you scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)
		`,
		},
	}
}
