package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WaxenShapethief is the card definition for Waxen Shapethief.
//
// Type: Creature — Shapeshifter
// Cost: {3}{U}
//
// Oracle text:
//
//	Flash
//	You may have this creature enter as a copy of an artifact or creature you control.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var WaxenShapethief = newWaxenShapethief

func newWaxenShapethief() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Waxen Shapethief",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of an artifact or creature you control.", &game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}, Controller: game.ControllerYou}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			Flash
			You may have this creature enter as a copy of an artifact or creature you control.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
