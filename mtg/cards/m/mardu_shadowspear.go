package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MarduShadowspear is the card definition for Mardu Shadowspear.
//
// Type: Creature — Human Warrior
// Cost: {B}
//
// Oracle text:
//
//	Whenever this creature attacks, each opponent loses 1 life.
//	Dash {1}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
var MarduShadowspear = newMarduShadowspear

func newMarduShadowspear() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Mardu Shadowspear",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
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
								Primitive: game.LoseLife{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
								},
							},
						},
					}.Ability(),
				},
				game.DashTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Dash",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.B}),
					Mechanic: cost.AlternativeMechanicDash,
				},
			},
			OracleText: `
			Whenever this creature attacks, each opponent loses 1 life.
			Dash {1}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
		`,
		},
	}
}
