package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GorgingVulture is the card definition for Gorging Vulture.
//
// Type: Creature — Bird
// Cost: {2}{B}
//
// Oracle text:
//
//	Flying
//	When this creature enters, mill four cards. You gain 1 life for each creature card milled this way. (To mill four cards, put the top four cards of your library into your graveyard.)
var GorgingVulture = newGorgingVulture

func newGorgingVulture() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gorging Vulture",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, mill four cards. You gain 1 life for each creature card milled this way. (To mill four cards, put the top four cards of your library into your graveyard.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(4),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
										},
										PublishResult: game.ResultKey("actual-card-quantity-1"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("actual-card-quantity-1")},
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
												ResultKey:  game.ResultKey("actual-card-quantity-1"),
											}),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:      game.ResultKey("actual-card-quantity-1"),
											Accepted: game.TriTrue,
										}),
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
			Flying
			When this creature enters, mill four cards. You gain 1 life for each creature card milled this way. (To mill four cards, put the top four cards of your library into your graveyard.)
		`,
		},
	}
}
