package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BardKingOfDale is the card definition for Bard, King of Dale.
//
// Type: Legendary Creature — Human Noble Archer
// Cost: {4}{W}{U}
//
// Oracle text:
//
//	Reach, vigilance
//	If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.
//	If one or more tokens would be created under your control, twice that many of those tokens are created instead.
var BardKingOfDale = newBardKingOfDale

func newBardKingOfDale() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Bard, King of Dale",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble, types.Archer},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.VigilanceStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DrawCardMultiplierReplacement("If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.", 2, true),
				game.TokenCreationReplacement("If one or more tokens would be created under your control, twice that many of those tokens are created instead.", 2, game.TriggerControllerYou),
			},
			OracleText: `
			Reach, vigilance
			If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.
			If one or more tokens would be created under your control, twice that many of those tokens are created instead.
		`,
		},
	}
}
