package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MonstrousCarabid is the card definition for Monstrous Carabid.
//
// Type: Creature — Insect
// Cost: {3}{B}{R}
//
// Oracle text:
//
//	This creature attacks each combat if able.
//	Cycling {B/R} ({B/R}, Discard this card: Draw a card.)
var MonstrousCarabid = newMonstrousCarabid

func newMonstrousCarabid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Monstrous Carabid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.R,
			}),
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.HybridMana(mana.B, mana.R)}),
			},
			OracleText: `
			This creature attacks each combat if able.
			Cycling {B/R} ({B/R}, Discard this card: Draw a card.)
		`,
		},
	}
}
