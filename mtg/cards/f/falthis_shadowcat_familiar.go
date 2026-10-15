package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FalthisShadowcatFamiliar is the card definition for Falthis, Shadowcat Familiar.
//
// Type: Legendary Creature — Nightmare Cat
// Cost: {2}{B}
//
// Oracle text:
//
//	Commanders you control have menace and deathtouch.
//	Partner (You can have two commanders if both have partner.)
var FalthisShadowcatFamiliar = newFalthisShadowcatFamiliar

func newFalthisShadowcatFamiliar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Falthis, Shadowcat Familiar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Nightmare, types.Cat},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Commanders you control have menace and deathtouch.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{MatchCommander: true}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Menace, game.Deathtouch},
						},
					},
				},
				game.PartnerStaticBody,
			},
			OracleText: `
			Commanders you control have menace and deathtouch.
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}
