package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AscendingAven is the card definition for Ascending Aven.
//
// Type: Creature — Bird Soldier
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Flying
//	This creature can block only creatures with flying.
//	Morph {2}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var AscendingAven = newAscendingAven

func newAscendingAven() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ascending Aven",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCanBlockOnlyCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlying,
							},
						},
					},
				},
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{Cost: cost.Mana{cost.O(2), cost.U}},
					},
				},
			},
			OracleText: `
			Flying
			This creature can block only creatures with flying.
			Morph {2}{U} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
