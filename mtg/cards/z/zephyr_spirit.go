package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZephyrSpirit is the card definition for Zephyr Spirit.
//
// Type: Creature — Spirit
// Cost: {5}{U}
//
// Oracle text:
//
//	When this creature blocks, return it to its owner's hand.
var ZephyrSpirit = newZephyrSpirit

func newZephyrSpirit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Zephyr Spirit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 6}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventBlockerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.EventPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature blocks, return it to its owner's hand.
		`,
		},
	}
}
