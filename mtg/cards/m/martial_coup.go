package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MartialCoup is the card definition for Martial Coup.
//
// Type: Sorcery
// Cost: {X}{W}{W}
//
// Oracle text:
//
//	Create X 1/1 white Soldier creature tokens. If X is 5 or more, destroy all other creatures.
var MartialCoup = newMartialCoup

func newMartialCoup() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Martial Coup",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source: game.TokenDef(martialCoupToken),
						},
					},
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true}),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateSpellX, Op: compare.GreaterOrEqual, Value: 5}},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Create X 1/1 white Soldier creature tokens. If X is 5 or more, destroy all other creatures.
		`,
		},
	}
}

var martialCoupToken = newMartialCoupToken()

func newMartialCoupToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
