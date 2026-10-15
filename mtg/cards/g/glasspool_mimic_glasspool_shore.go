package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlasspoolMimic is the card definition for Glasspool Mimic // Glasspool Shore.
//
// Type: Creature — Shapeshifter Rogue // Land
// Face: Glasspool Shore — Land
//
// Oracle text:
//
//	You may have this creature enter as a copy of a creature you control, except it's a Shapeshifter Rogue in addition to its other types.
var GlasspoolMimic = newGlasspoolMimic

func newGlasspoolMimic() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glasspool Mimic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter, types.Rogue},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of a creature you control, except it's a Shapeshifter Rogue in addition to its other types.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, true, false, nil, false, nil, []types.Sub{types.Shapeshifter, types.Rogue}),
			},
			OracleText: `
			You may have this creature enter as a copy of a creature you control, except it's a Shapeshifter Rogue in addition to its other types.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Glasspool Shore",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {U}.
		`,
		}),
	}
}
