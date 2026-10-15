package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AvacynAngelOfHope is the card definition for Avacyn, Angel of Hope.
//
// Type: Legendary Creature — Angel
// Cost: {5}{W}{W}{W}
//
// Oracle text:
//
//	Flying, vigilance, indestructible
//	Other permanents you control have indestructible.
var AvacynAngelOfHope = newAvacynAngelOfHope

func newAvacynAngelOfHope() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Avacyn, Angel of Hope",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel},
			Power:      opt.Val(game.PT{Value: 8}),
			Toughness:  opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
				game.IndestructibleStaticBody,
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
			Flying, vigilance, indestructible
			Other permanents you control have indestructible.
		`,
		},
	}
}
