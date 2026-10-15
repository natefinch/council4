package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodletterOfAclazotz is the card definition for Bloodletter of Aclazotz.
//
// Type: Creature — Vampire Demon
// Cost: {1}{B}{B}{B}
//
// Oracle text:
//
//	Flying
//	If an opponent would lose life during your turn, they lose twice that much life instead. (Damage causes loss of life.)
var BloodletterOfAclazotz = newBloodletterOfAclazotz

func newBloodletterOfAclazotz() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bloodletter of Aclazotz",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Demon},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.LifeLossReplacement("If an opponent would lose life during your turn, they lose twice that much life instead. (Damage causes loss of life.)", 2, 0, true, true),
			},
			OracleText: `
			Flying
			If an opponent would lose life during your turn, they lose twice that much life instead. (Damage causes loss of life.)
		`,
		},
	}
}
