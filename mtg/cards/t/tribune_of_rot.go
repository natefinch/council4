package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TribuneOfRot is the card definition for Tribune of Rot.
//
// Type: Creature — Elf Shaman
// Cost: {1}{B/G}{B/G}
//
// Oracle text:
//
//	Whenever this creature attacks, mill two cards. For each creature card milled this way, create a 1/1 green Saproling creature token.
var TribuneOfRot = newTribuneOfRot

func newTribuneOfRot() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Tribune of Rot",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.B, mana.G),
				cost.HybridMana(mana.B, mana.G),
			}),
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Shaman},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, mill two cards. For each creature card milled this way, create a 1/1 green Saproling creature token.",
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
											Amount:      game.Fixed(2),
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
											Source: game.TokenDef(tribuneOfRotToken),
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
			Whenever this creature attacks, mill two cards. For each creature card milled this way, create a 1/1 green Saproling creature token.
		`,
		},
	}
}

var tribuneOfRotToken = newTribuneOfRotToken()

func newTribuneOfRotToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Saproling",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Saproling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
