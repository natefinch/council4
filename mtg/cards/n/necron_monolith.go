package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NecronMonolith is the card definition for Necron Monolith.
//
// Type: Artifact — Vehicle
// Cost: {7}
//
// Oracle text:
//
//	Flying, indestructible
//	Eternity Gate — Whenever this Vehicle attacks, mill three cards. For each creature card milled this way, create a 2/2 black Necron Warrior artifact creature token.
//	Crew 4 (Tap any number of creatures you control with total power 4 or more: This Vehicle becomes an artifact creature until end of turn.)
var NecronMonolith = newNecronMonolith

func newNecronMonolith() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Necron Monolith",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.IndestructibleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(4),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Eternity Gate — Whenever this Vehicle attacks, mill three cards. For each creature card milled this way, create a 2/2 black Necron Warrior artifact creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
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
											Amount:      game.Fixed(3),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
										},
										PublishResult: game.ResultKey("actual-card-quantity-1"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("actual-card-quantity-1")},
										},
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												Selection:  &game.Selection{RequiredTypes: []types.Card{types.Creature}},
												ResultKey:  game.ResultKey("actual-card-quantity-1"),
											}),
											Source: game.TokenDef(necronMonolithToken),
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
			Flying, indestructible
			Eternity Gate — Whenever this Vehicle attacks, mill three cards. For each creature card milled this way, create a 2/2 black Necron Warrior artifact creature token.
			Crew 4 (Tap any number of creatures you control with total power 4 or more: This Vehicle becomes an artifact creature until end of turn.)
		`,
		},
	}
}

var necronMonolithToken = newNecronMonolithToken()

func newNecronMonolithToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Necron Warrior",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Necron, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
