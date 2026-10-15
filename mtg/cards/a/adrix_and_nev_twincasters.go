package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdrixAndNevTwincasters is the card definition for Adrix and Nev, Twincasters.
//
// Type: Legendary Creature — Merfolk Wizard
// Cost: {2}{G}{U}
//
// Oracle text:
//
//	Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
//	If one or more tokens would be created under your control, twice that many of those tokens are created instead.
var AdrixAndNevTwincasters = newAdrixAndNevTwincasters

func newAdrixAndNevTwincasters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Adrix and Nev, Twincasters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Merfolk, types.Wizard},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacement("If one or more tokens would be created under your control, twice that many of those tokens are created instead.", 2, game.TriggerControllerYou),
			},
			OracleText: `
			Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
			If one or more tokens would be created under your control, twice that many of those tokens are created instead.
		`,
		},
	}
}
