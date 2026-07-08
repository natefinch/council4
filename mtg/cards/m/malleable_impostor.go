package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MalleableImpostor is the card definition for Malleable Impostor.
//
// Type: Creature — Faerie Shapeshifter
// Cost: {3}{U}
//
// Oracle text:
//
//	Flash
//	Flying
//	You may have this creature enter as a copy of a creature an opponent controls, except it's a Faerie Shapeshifter in addition to its other types and it has flying.
var MalleableImpostor = newMalleableImpostor

func newMalleableImpostor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Malleable Impostor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of a creature an opponent controls, except it's a Faerie Shapeshifter in addition to its other types and it has flying.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}, true, false, nil, false, []game.Keyword{game.Flying}, []types.Sub{types.Faerie, types.Shapeshifter}),
			},
			OracleText: `
			Flash
			Flying
			You may have this creature enter as a copy of a creature an opponent controls, except it's a Faerie Shapeshifter in addition to its other types and it has flying.
		`,
		},
	}
}
