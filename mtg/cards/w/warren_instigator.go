package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WarrenInstigator is the card definition for Warren Instigator.
//
// Type: Creature — Goblin Berserker
// Cost: {R}{R}
//
// Oracle text:
//
//	Double strike
//	Whenever this creature deals damage to an opponent, you may put a Goblin creature card from your hand onto the battlefield.
var WarrenInstigator = newWarrenInstigator

func newWarrenInstigator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Warren Instigator",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Berserker},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature deals damage to an opponent, you may put a Goblin creature card from your hand onto the battlefield.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventDamageDealt,
							Source:          game.TriggerSourceSelf,
							Player:          game.TriggerPlayerOpponent,
							Subject:         game.TriggerSubjectDamageSource,
							StackObjectKind: game.StackSpell,
							DamageRecipient: game.DamageRecipientPlayer,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Hand,
											Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Goblin}},
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Battlefield,
											},
											Prompt: "Choose a card to put onto the battlefield",
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
			Double strike
			Whenever this creature deals damage to an opponent, you may put a Goblin creature card from your hand onto the battlefield.
		`,
		},
	}
}
