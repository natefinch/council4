package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Warrior
//
// Type: Token Creature — Warrior
//
// Oracle text:
//   Whenever this creature and at least one other creature token attack, put a +1/+1 counter on this creature.

// WarriorTokenb7fc2167531344e8abbfa999fba79a93 is the card definition for Warrior.
var WarriorTokenb7fc2167531344e8abbfa999fba79a93 = newWarriorTokenb7fc2167531344e8abbfa999fba79a93()

func newWarriorTokenb7fc2167531344e8abbfa999fba79a93() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:      "Warrior",
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                     game.EventAttackerDeclared,
							Source:                    game.TriggerSourceSelf,
							AttacksAlongsideCount:     1,
							AttacksAlongsideSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, TokenOnly: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature and at least one other creature token attack, put a +1/+1 counter on this creature.
		`,
		},
	}
}
