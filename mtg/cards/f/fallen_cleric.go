package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FallenCleric is the card definition for Fallen Cleric.
//
// Type: Creature — Zombie Cleric
// Cost: {4}{B}
//
// Oracle text:
//
//	Protection from Clerics
//	Morph {4}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var FallenCleric = newFallenCleric

func newFallenCleric() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fallen Cleric",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Cleric},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromSubtypesStaticAbility(types.Cleric),
				game.StaticAbility{
					Text: "Morph {4}{B}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(4),
								cost.B,
							},
						},
					},
				},
			},
			OracleText: `
			Protection from Clerics
			Morph {4}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
