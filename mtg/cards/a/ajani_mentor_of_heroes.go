package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AjaniMentorOfHeroes is the card definition for Ajani, Mentor of Heroes.
//
// Type: Legendary Planeswalker — Ajani
// Cost: {3}{G}{W}
//
// Oracle text:
//
//	+1: Distribute three +1/+1 counters among one, two, or three target creatures you control.
//	+1: Look at the top four cards of your library. You may reveal an Aura, creature, or planeswalker card from among them and put it into your hand. Put the rest on the bottom of your library in any order.
//	−8: You gain 100 life.
var AjaniMentorOfHeroes = newAjaniMentorOfHeroes

func newAjaniMentorOfHeroes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Ajani, Mentor of Heroes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Ajani},
			Loyalty:    opt.Val(4),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 3,
								Constraint: "one, two, or three target creatures you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(3),
									Object:      game.AllTargetPermanentsReference(0),
									CounterKind: counter.PlusOnePlusOne,
									Distribute:  true,
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(4),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, SubtypesAny: []types.Sub{types.Sub("Aura")}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
				game.LoyaltyAbility{
					LoyaltyCost: -8,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(100),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			+1: Distribute three +1/+1 counters among one, two, or three target creatures you control.
			+1: Look at the top four cards of your library. You may reveal an Aura, creature, or planeswalker card from among them and put it into your hand. Put the rest on the bottom of your library in any order.
			−8: You gain 100 life.
		`,
		},
	}
}
