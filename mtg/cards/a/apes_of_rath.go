package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ApesOfRath is the card definition for Apes of Rath.
//
// Type: Creature — Ape
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Whenever this creature attacks, it doesn't untap during its controller's next untap step.
var ApesOfRath = newApesOfRath()

func newApesOfRath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Apes of Rath",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ape},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
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
								Primitive: game.SkipNextUntap{
									Object: game.EventPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, it doesn't untap during its controller's next untap step.
		`,
		},
	}
}
