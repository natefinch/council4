package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GrizzlyFate is the card definition for Grizzly Fate.
//
// Type: Sorcery
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Create two 2/2 green Bear creature tokens.
//	Threshold — Create four 2/2 green Bear creature tokens instead if there are seven or more cards in your graveyard.
//	Flashback {5}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var GrizzlyFate = newGrizzlyFate

func newGrizzlyFate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Grizzly Fate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Flashback {5}{G}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{
							Cost: cost.Mana{
								cost.O(5),
								cost.G,
								cost.G,
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(grizzlyFateToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:     true,
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7}},
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(4),
							Source: game.TokenDef(grizzlyFateToken),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Create two 2/2 green Bear creature tokens.
			Threshold — Create four 2/2 green Bear creature tokens instead if there are seven or more cards in your graveyard.
			Flashback {5}{G}{G} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}

var grizzlyFateToken = newGrizzlyFateToken()

func newGrizzlyFateToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Bear",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bear},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
