package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OketraSAttendant is the card definition for Oketra's Attendant.
//
// Type: Creature — Bird Soldier
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flying
//	Cycling {2} ({2}, Discard this card: Draw a card.)
//	Embalm {3}{W}{W} ({3}{W}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Bird Soldier with no mana cost. Embalm only as a sorcery.)
var OketraSAttendant = newOketraSAttendant

func newOketraSAttendant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Oketra's Attendant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
				game.EmbalmActivatedBody(cost.Mana{cost.O(3), cost.W, cost.W}, types.Sub("Bird"), types.Sub("Soldier")),
			},
			OracleText: `
			Flying
			Cycling {2} ({2}, Discard this card: Draw a card.)
			Embalm {3}{W}{W} ({3}{W}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a white Zombie Bird Soldier with no mana cost. Embalm only as a sorcery.)
		`,
		},
	}
}
