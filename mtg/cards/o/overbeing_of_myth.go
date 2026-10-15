package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OverbeingOfMyth is the card definition for Overbeing of Myth.
//
// Type: Creature — Spirit Avatar
// Cost: {G/U}{G/U}{G/U}{G/U}{G/U}
//
// Oracle text:
//
//	Overbeing of Myth's power and toughness are each equal to the number of cards in your hand.
//	At the beginning of your draw step, draw an additional card.
var OverbeingOfMyth = newOverbeingOfMyth

func newOverbeingOfMyth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Overbeing of Myth",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors:           []color.Color{color.Green, color.Blue},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Spirit, types.Avatar},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your draw step, draw an additional card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepDraw,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			Overbeing of Myth's power and toughness are each equal to the number of cards in your hand.
			At the beginning of your draw step, draw an additional card.
		`,
		},
	}
}
