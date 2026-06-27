package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThingFromTheDeep is the card definition for Thing from the Deep.
//
// Type: Creature — Leviathan
// Cost: {6}{U}{U}{U}
//
// Oracle text:
//
//	Whenever this creature attacks, sacrifice it unless you sacrifice an Island.
var ThingFromTheDeep = newThingFromTheDeep()

func newThingFromTheDeep() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thing from the Deep",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Leviathan},
			Power:     opt.Val(game.PT{Value: 9}),
			Toughness: opt.Val(game.PT{Value: 9}),
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
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Sacrifice an Island?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:        cost.AdditionalSacrifice,
												Text:        "sacrifice an Island",
												Amount:      1,
												SubtypesAny: cost.SubtypeSet{types.Island},
											},
										},
									},
								},
								PublishResult: game.ResultKey("sacrifice-unless-paid"),
							},
							{
								Primitive: game.Sacrifice{
									Object: game.EventPermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "sacrifice-unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, sacrifice it unless you sacrifice an Island.
		`,
		},
	}
}
