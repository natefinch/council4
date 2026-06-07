package r

import (
	"github.com/natefinch/council4/mtg/cards/common"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RampantGrowth is the card definition for Rampant Growth.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.
var RampantGrowth = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Rampant Growth",
		ManaCost: opt.Val(cost.Mana{
			cost.O(1),
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Sorcery},
		OracleText: `
			Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.
		`,
		SpellAbility: opt.Val(game.SpellAbilityBody{
			Text: `
				Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.
			`,
			Content: common.RampLand{Tapped: true, Basic: true}.Ability(),
		}),
	},
}
