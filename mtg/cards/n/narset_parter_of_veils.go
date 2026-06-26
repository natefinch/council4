package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NarsetParterOfVeils is the card definition for Narset, Parter of Veils.
//
// Type: Legendary Planeswalker — Narset
// Cost: {1}{U}{U}
//
// Oracle text:
//
//	Each opponent can't draw more than one card each turn.
//	−2: Look at the top four cards of your library. You may reveal a noncreature, nonland card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var NarsetParterOfVeils = newNarsetParterOfVeils()

func newNarsetParterOfVeils() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Narset, Parter of Veils",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Planeswalker},
			Subtypes:   []types.Sub{types.Narset},
			Loyalty:    opt.Val(5),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectDrawLimitPerTurn,
							AffectedPlayer:   game.PlayerOpponent,
							DrawLimitPerTurn: 1,
						},
					},
				},
			},
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					LoyaltyCost: -2,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(4),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Each opponent can't draw more than one card each turn.
			−2: Look at the top four cards of your library. You may reveal a noncreature, nonland card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
