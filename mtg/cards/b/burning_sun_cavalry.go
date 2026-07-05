package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BurningSunCavalry is the card definition for Burning Sun Cavalry.
//
// Type: Creature — Human Knight
// Cost: {1}{R}
//
// Oracle text:
//
//	Whenever this creature attacks or blocks while you control a Dinosaur, this creature gets +1/+1 until end of turn.
var BurningSunCavalry = newBurningSunCavalry()

func newBurningSunCavalry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Burning Sun Cavalry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventBlockerDeclared,
						},
						InterveningIf: "while you control a Dinosaur",
						InterveningCondition: opt.Val(game.Condition{
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Dinosaur")}},
							}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks or blocks while you control a Dinosaur, this creature gets +1/+1 until end of turn.
		`,
		},
	}
}
