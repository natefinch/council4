package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CanopyVista is the card definition for Canopy Vista.
//
// Type: Land — Forest Plains
//
// Oracle text:
//
//	({T}: Add {G} or {W}.)
//	This land enters tapped unless you control two or more basic lands.
var CanopyVista = newCanopyVista

func newCanopyVista() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:     "Canopy Vista",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control two or more basic lands.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
						MinCount:  2,
					}),
				}),
			},
			OracleText: `
			({T}: Add {G} or {W}.)
			This land enters tapped unless you control two or more basic lands.
		`,
		},
	}
}
