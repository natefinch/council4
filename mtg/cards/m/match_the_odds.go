package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MatchTheOdds is the card definition for Match the Odds.
//
// Type: Sorcery — Lesson
// Cost: {2}{G}
//
// Oracle text:
//
//	Create a 1/1 white Ally creature token. Put a +1/+1 counter on it for each creature your opponents control.
var MatchTheOdds = newMatchTheOdds()

func newMatchTheOdds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Match the Odds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Lesson},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:        game.Fixed(1),
							Source:        game.TokenDef(matchTheOddsToken),
							PublishLinked: game.LinkedKey("created-token"),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							}),
							Object:      game.LinkedObjectReference("created-token"),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 1/1 white Ally creature token. Put a +1/+1 counter on it for each creature your opponents control.
		`,
		},
	}
}

var matchTheOddsToken = newMatchTheOddsToken()

func newMatchTheOddsToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Ally",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ally},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
