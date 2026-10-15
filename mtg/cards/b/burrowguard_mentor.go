package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BurrowguardMentor is the card definition for Burrowguard Mentor.
//
// Type: Creature — Rabbit Soldier
// Cost: {G}{W}
//
// Oracle text:
//
//	Trample
//	Burrowguard Mentor's power and toughness are each equal to the number of creatures you control.
var BurrowguardMentor = newBurrowguardMentor

func newBurrowguardMentor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Burrowguard Mentor",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
			}),
			Colors:           []color.Color{color.Green, color.White},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Rabbit, types.Soldier},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample
			Burrowguard Mentor's power and toughness are each equal to the number of creatures you control.
		`,
		},
	}
}
