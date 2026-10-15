package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShimmerMyr is the card definition for Shimmer Myr.
//
// Type: Artifact Creature — Myr
// Cost: {3}
//
// Oracle text:
//
//	Flash
//	You may cast artifact spells as though they had flash.
var ShimmerMyr = newShimmerMyr

func newShimmerMyr() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Shimmer Myr",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "You may cast artifact spells as though they had flash.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellTypes:     []types.Card{types.Artifact},
						},
					},
				},
			},
			OracleText: `
			Flash
			You may cast artifact spells as though they had flash.
		`,
		},
	}
}
