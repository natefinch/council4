package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cloud Sprite
//
// Type: Token Creature — Faerie
//
// Oracle text:
//   Flying
//   Cloud Sprite can block only creatures with flying.

// CloudSpriteToken5b04de5c19b74767a38fd5420d5da472 is the card definition for Cloud Sprite.
var CloudSpriteToken5b04de5c19b74767a38fd5420d5da472 = newCloudSpriteToken5b04de5c19b74767a38fd5420d5da472()

func newCloudSpriteToken5b04de5c19b74767a38fd5420d5da472() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Cloud Sprite",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie},
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
			Cloud Sprite can block only creatures with flying.
		`,
		},
	}
}
