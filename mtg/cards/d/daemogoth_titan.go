package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DaemogothTitan is the card definition for Daemogoth Titan.
//
// Type: Creature — Demon
// Cost: {B/G}{B/G}{B/G}{B/G}
//
// Oracle text:
//
//	Whenever this creature attacks or blocks, sacrifice a creature.
var DaemogothTitan = newDaemogothTitan

func newDaemogothTitan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Daemogoth Titan",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Demon},
			Power:     opt.Val(game.PT{Value: 11}),
			Toughness: opt.Val(game.PT{Value: 10}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks or blocks, sacrifice a creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							UnionEvent:      game.EventBlockerDeclared,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											Player:    game.ControllerReference(),
											Amount:    game.Fixed(1),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Whenever this creature attacks or blocks, sacrifice a creature.
		`,
		},
	}
}
