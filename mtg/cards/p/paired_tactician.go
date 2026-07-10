package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PairedTactician is the card definition for Paired Tactician.
//
// Type: Creature — Human Warrior
// Cost: {2}{W}
//
// Oracle text:
//
//	Whenever this creature and at least one other Warrior attack, put a +1/+1 counter on this creature.
var PairedTactician = newPairedTactician

func newPairedTactician() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Paired Tactician",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
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
							AttacksAlongsideSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Warrior")}},
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
			Whenever this creature and at least one other Warrior attack, put a +1/+1 counter on this creature.
		`,
		},
	}
}
