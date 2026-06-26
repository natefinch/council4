package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Bird
//
// Type: Token Creature — Bird
//
// Oracle text:
//   Flying
//   This creature can block only creatures with flying.

// BirdTokenf50b144c021a40259dc71639fdc65505 is the card definition for Bird.
var BirdTokenf50b144c021a40259dc71639fdc65505 = newBirdTokenf50b144c021a40259dc71639fdc65505()

func newBirdTokenf50b144c021a40259dc71639fdc65505() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Bird",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
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
			},
			OracleText: `
			Flying
			This creature can block only creatures with flying.
		`,
		},
	}
}
