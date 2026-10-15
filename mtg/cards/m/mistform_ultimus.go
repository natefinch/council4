package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MistformUltimus is the card definition for Mistform Ultimus.
//
// Type: Legendary Creature — Illusion
// Cost: {3}{U}
//
// Oracle text:
//
//	Mistform Ultimus is every creature type (even if this card isn't on the battlefield).
var MistformUltimus = newMistformUltimus

func newMistformUltimus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mistform Ultimus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Illusion},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Mistform Ultimus is every creature type (even if this card isn't on the battlefield).",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:             game.DurationPermanent,
							AffectedSource:       true,
							Layer:                game.LayerType,
							AddEveryCreatureType: true,
						},
					},
				},
			},
			OracleText: `
			Mistform Ultimus is every creature type (even if this card isn't on the battlefield).
		`,
		},
	}
}
