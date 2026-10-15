package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DraconicVisitor is the card definition for Draconic Visitor.
//
// Type: Creature — Dragon
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Flying
//	If one or more artifact tokens would be created under your control, that many 5/5 red Dragon creature tokens with flying are created instead.
var DraconicVisitor = newDraconicVisitor

func newDraconicVisitor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Draconic Visitor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacementFiltered("If one or more artifact tokens would be created under your control, that many 5/5 red Dragon creature tokens with flying are created instead.", &game.TokenCreationReplacementSpec{Multiplier: 1, Types: []types.Card{types.Artifact}, ReplaceDef: draconicVisitorToken, Filter: game.TriggerControllerYou}),
			},
			OracleText: `
			Flying
			If one or more artifact tokens would be created under your control, that many 5/5 red Dragon creature tokens with flying are created instead.
		`,
		},
	}
}

var draconicVisitorToken = newDraconicVisitorToken()

func newDraconicVisitorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Dragon",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
