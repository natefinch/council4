package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BogStriderAsh is the card definition for Bog-Strider Ash.
//
// Type: Creature — Treefolk Shaman
// Cost: {3}{G}
//
// Oracle text:
//
//	Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
//	Whenever a player casts a Goblin spell, you may pay {G}. If you do, you gain 2 life.
var BogStriderAsh = newBogStriderAsh()

func newBogStriderAsh() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bog-Strider Ash",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.SwampwalkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							CardSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {G}?",
										ManaCost: opt.Val(cost.Mana{
											cost.G,
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
			Whenever a player casts a Goblin spell, you may pay {G}. If you do, you gain 2 life.
		`,
		},
	}
}
