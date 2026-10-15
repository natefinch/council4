package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImposingVantasaur is the card definition for Imposing Vantasaur.
//
// Type: Creature — Dinosaur
// Cost: {5}{W}
//
// Oracle text:
//
//	Vigilance
//	Cycling {1} ({1}, Discard this card: Draw a card.)
var ImposingVantasaur = newImposingVantasaur

func newImposingVantasaur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Imposing Vantasaur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
			},
			OracleText: `
			Vigilance
			Cycling {1} ({1}, Discard this card: Draw a card.)
		`,
		},
	}
}
