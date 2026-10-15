package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AscendantPackleader is the card definition for Ascendant Packleader.
//
// Type: Creature — Wolf
// Cost: {G}
//
// Oracle text:
//
//	This creature enters with a +1/+1 counter on it if you control a permanent with mana value 4 or greater.
//	Whenever you cast a spell with mana value 4 or greater, put a +1/+1 counter on this creature.
var AscendantPackleader = newAscendantPackleader

func newAscendantPackleader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ascendant Packleader",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a spell with mana value 4 or greater, put a +1/+1 counter on this creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventSpellCast,
							Controller: game.TriggerControllerYou,
							CardSelection: game.Selection{ManaValue: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 4,
							})},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("This creature enters with a +1/+1 counter on it if you control a permanent with mana value 4 or greater.", &game.Condition{
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
					}),
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			This creature enters with a +1/+1 counter on it if you control a permanent with mana value 4 or greater.
			Whenever you cast a spell with mana value 4 or greater, put a +1/+1 counter on this creature.
		`,
		},
	}
}
