package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IlluminateHistory is the card definition for Illuminate History.
//
// Type: Sorcery — Lesson
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Discard any number of cards, then draw that many cards. Then if there are seven or more cards in your graveyard, create a 3/2 red and white Spirit creature token.
var IlluminateHistory = newIlluminateHistory

func newIlluminateHistory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Illuminate History",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Lesson},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.DiscardThenDraw{
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(illuminateHistoryToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7}},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Discard any number of cards, then draw that many cards. Then if there are seven or more cards in your graveyard, create a 3/2 red and white Spirit creature token.
		`,
		},
	}
}

var illuminateHistoryToken = newIlluminateHistoryToken()

func newIlluminateHistoryToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
