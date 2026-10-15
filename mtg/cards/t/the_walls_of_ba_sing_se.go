package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheWallsOfBaSingSe is the card definition for The Walls of Ba Sing Se.
//
// Type: Legendary Artifact Creature — Wall
// Cost: {8}
//
// Oracle text:
//
//	Defender
//	Other permanents you control have indestructible.
var TheWallsOfBaSingSe = newTheWallsOfBaSingSe

func newTheWallsOfBaSingSe() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "The Walls of Ba Sing Se",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Wall},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 30}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.StaticAbility{
					Text: "Other permanents you control have indestructible.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Indestructible},
						},
					},
				},
			},
			OracleText: `
			Defender
			Other permanents you control have indestructible.
		`,
		},
	}
}
