package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlotOutTheSky is the card definition for Blot Out the Sky.
//
// Type: Sorcery
// Cost: {X}{W}{B}
//
// Oracle text:
//
//	Create X tapped 2/1 white and black Inkling creature tokens with flying. If X is 6 or more, destroy all noncreature, nonland permanents.
var BlotOutTheSky = newBlotOutTheSky()

func newBlotOutTheSky() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Blot Out the Sky",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Source:      game.TokenDef(blotOutTheSkyToken),
							EntryTapped: true,
						},
					},
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{ExcludedTypes: []types.Card{types.Creature, types.Land}}),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateSpellX, Op: compare.GreaterOrEqual, Value: 6}},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Create X tapped 2/1 white and black Inkling creature tokens with flying. If X is 6 or more, destroy all noncreature, nonland permanents.
		`,
		},
	}
}

var blotOutTheSkyToken = newBlotOutTheSkyToken()

func newBlotOutTheSkyToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Inkling",
			Colors:    []color.Color{color.White, color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Inkling},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
