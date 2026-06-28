package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HarborGuardian is the card definition for Harbor Guardian.
//
// Type: Creature — Gargoyle
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Whenever this creature attacks, defending player may draw a card.
var HarborGuardian = newHarborGuardian()

func newHarborGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Harbor Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
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
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.DefendingPlayerReference(),
								},
								Optional:      true,
								OptionalActor: opt.Val(game.DefendingPlayerReference()),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Whenever this creature attacks, defending player may draw a card.
		`,
		},
	}
}
