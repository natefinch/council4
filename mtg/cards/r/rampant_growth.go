package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
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
	Name: "Rampant Growth",
	ManaCost: opt.Val(cost.Mana{
		cost.O(1),
		cost.G,
	}),
	Colors:        []color.Color{color.Green},
	ColorIdentity: mana.NewColorIdentity(color.Green),
	Types:         []types.Card{types.Sorcery},
	OracleText:    "Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle.",
			Effects: []game.Effect{
				{
					Type:        game.EffectSearch,
					TargetIndex: game.TargetIndexController,
					Search: opt.Val(game.SearchSpec{
						SourceZone:   game.ZoneLibrary,
						Destination:  game.ZoneBattlefield,
						CardType:     opt.Val(types.Land),
						Supertype:    opt.Val(types.Basic),
						Shuffle:      true,
						EntersTapped: true,
					}),
				},
			},
		},
	},
}
