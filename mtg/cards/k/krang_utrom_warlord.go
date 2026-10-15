package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KrangUtromWarlord is the card definition for Krang, Utrom Warlord.
//
// Type: Legendary Artifact Creature — Utrom Robot
// Cost: {9}
//
// Oracle text:
//
//	Flying, trample, indestructible, haste
//	Other artifact creatures you control have flying, trample, indestructible, and haste.
var KrangUtromWarlord = newKrangUtromWarlord

func newKrangUtromWarlord() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Krang, Utrom Warlord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(9),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Utrom, types.Robot},
			Power:      opt.Val(game.PT{Value: 9}),
			Toughness:  opt.Val(game.PT{Value: 9}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
				game.IndestructibleStaticBody,
				game.HasteStaticBody,
				game.StaticAbility{
					Text: "Other artifact creatures you control have flying, trample, indestructible, and haste.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying, game.Trample, game.Indestructible, game.Haste},
						},
					},
				},
			},
			OracleText: `
			Flying, trample, indestructible, haste
			Other artifact creatures you control have flying, trample, indestructible, and haste.
		`,
		},
	}
}
