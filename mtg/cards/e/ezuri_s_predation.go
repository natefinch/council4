package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EzuriSPredation is the card definition for Ezuri's Predation.
//
// Type: Sorcery
// Cost: {5}{G}{G}{G}
//
// Oracle text:
//
//	For each creature your opponents control, create a 4/4 green Phyrexian Beast creature token. Each of those tokens fights a different one of those creatures.
var EzuriSPredation = newEzuriSPredation

func newEzuriSPredation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ezuri's Predation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							}),
							Source:            game.TokenDef(ezuriSPredationToken),
							PublishLinked:     game.LinkedKey("correlated-fight-tokens"),
							PublishCountGroup: game.LinkedKey("correlated-fight-creatures"),
						},
					},
					{
						Primitive: game.CorrelatedFight{
							Subjects: game.LinkedKey("correlated-fight-tokens"),
							Objects:  game.LinkedKey("correlated-fight-creatures"),
						},
					},
				},
			}.Ability()),
			OracleText: `
			For each creature your opponents control, create a 4/4 green Phyrexian Beast creature token. Each of those tokens fights a different one of those creatures.
		`,
		},
	}
}

var ezuriSPredationToken = newEzuriSPredationToken()

func newEzuriSPredationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Phyrexian Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
