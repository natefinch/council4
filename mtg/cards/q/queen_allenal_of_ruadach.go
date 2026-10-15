package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QueenAllenalOfRuadach is the card definition for Queen Allenal of Ruadach.
//
// Type: Legendary Creature — Elf Noble
// Cost: {G}{W}{W}
//
// Oracle text:
//
//	Queen Allenal of Ruadach's power and toughness are each equal to the number of creatures you control.
//	If one or more creature tokens would be created under your control, those tokens plus a 1/1 white Soldier creature token are created instead.
var QueenAllenalOfRuadach = newQueenAllenalOfRuadach

func newQueenAllenalOfRuadach() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Queen Allenal of Ruadach",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
				cost.W,
			}),
			Colors:           []color.Color{color.Green, color.White},
			Supertypes:       []types.Super{types.Legendary},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Elf, types.Noble},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacementFiltered("If one or more creature tokens would be created under your control, those tokens plus a 1/1 white Soldier creature token are created instead.", &game.TokenCreationReplacementSpec{Multiplier: 1, Addend: 1, Types: []types.Card{types.Creature}, AddendDef: queenAllenalOfRuadachToken, Filter: game.TriggerControllerYou}),
			},
			OracleText: `
			Queen Allenal of Ruadach's power and toughness are each equal to the number of creatures you control.
			If one or more creature tokens would be created under your control, those tokens plus a 1/1 white Soldier creature token are created instead.
		`,
		},
	}
}

var queenAllenalOfRuadachToken = newQueenAllenalOfRuadachToken()

func newQueenAllenalOfRuadachToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Soldier",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
