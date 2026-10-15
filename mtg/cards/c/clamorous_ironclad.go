package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ClamorousIronclad is the card definition for Clamorous Ironclad.
//
// Type: Artifact — Vehicle
// Cost: {3}{R}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	Crew 3 (Tap any number of creatures you control with total power 3 or more: This Vehicle becomes an artifact creature until end of turn.)
//	Cycling {R} ({R}, Discard this card: Draw a card.)
var ClamorousIronclad = newClamorousIronclad

func newClamorousIronclad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Clamorous Ironclad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(3),
				game.CyclingActivatedAbility(cost.Mana{cost.R}),
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			Crew 3 (Tap any number of creatures you control with total power 3 or more: This Vehicle becomes an artifact creature until end of turn.)
			Cycling {R} ({R}, Discard this card: Draw a card.)
		`,
		},
	}
}
