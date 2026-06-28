package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AnuridBrushhopper is the card definition for Anurid Brushhopper.
//
// Type: Creature — Frog Beast
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	Discard two cards: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
var AnuridBrushhopper = newAnuridBrushhopper()

func newAnuridBrushhopper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Anurid Brushhopper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Frog, types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Discard two cards: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard two cards",
							Amount: 2,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.SourcePermanentReference(),
									ExileLinkedKey: game.LinkedKey("delayed-self-blink"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.PutOnBattlefield{
														Source: game.LinkedBattlefieldSource(game.LinkedKey("delayed-self-blink")),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Discard two cards: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.
		`,
		},
	}
}
