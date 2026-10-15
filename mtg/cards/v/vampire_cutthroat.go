package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VampireCutthroat is the card definition for Vampire Cutthroat.
//
// Type: Creature — Vampire Rogue
// Cost: {B}
//
// Oracle text:
//
//	Skulk (This creature can't be blocked by creatures with greater power.)
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var VampireCutthroat = newVampireCutthroat

func newVampireCutthroat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vampire Cutthroat",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.SkulkStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Skulk (This creature can't be blocked by creatures with greater power.)
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
