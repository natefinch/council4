package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cryptothrall is the card definition for Cryptothrall.
//
// Type: Artifact Creature — Construct
// Cost: {4}
//
// Oracle text:
//
//	Protector — Other artifact creatures you control have hexproof.
var Cryptothrall = newCryptothrall

func newCryptothrall() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Cryptothrall",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Protector — Other artifact creatures you control have hexproof.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Hexproof},
						},
					},
				},
			},
			OracleText: `
			Protector — Other artifact creatures you control have hexproof.
		`,
		},
	}
}
