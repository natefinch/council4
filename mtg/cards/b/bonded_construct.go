package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BondedConstruct is the card definition for Bonded Construct.
//
// Type: Artifact Creature — Construct
// Cost: {1}
//
// Oracle text:
//
//	This creature can't attack alone.
var BondedConstruct = newBondedConstruct()

func newBondedConstruct() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Bonded Construct",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.CantAttackAloneStaticBody,
			},
			OracleText: `
			This creature can't attack alone.
		`,
		},
	}
}
