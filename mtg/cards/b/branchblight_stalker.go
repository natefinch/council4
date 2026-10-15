package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BranchblightStalker is the card definition for Branchblight Stalker.
//
// Type: Creature — Phyrexian Elf Scout
// Cost: {1}{G}
//
// Oracle text:
//
//	Toxic 2 (Players dealt combat damage by this creature also get two poison counters.)
var BranchblightStalker = newBranchblightStalker

func newBranchblightStalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Branchblight Stalker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Elf, types.Scout},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Toxic 2",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 2,
						},
					},
				},
			},
			OracleText: `
			Toxic 2 (Players dealt combat damage by this creature also get two poison counters.)
		`,
		},
	}
}
