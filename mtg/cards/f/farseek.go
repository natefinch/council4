package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Farseek is the card definition for Farseek.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Search your library for a Plains, Island, Swamp, or Mountain card, put it onto the battlefield tapped, then shuffle.
var Farseek = &game.CardDef{CardFace: game.CardFace{Name: "Farseek",
	ManaCost: opt.Val(cost.Mana{
		cost.O(1),
		cost.G,
	}),
	Colors: []color.Color{color.Green},

	Types:      []types.Card{types.Sorcery},
	OracleText: "Search your library for a Plains, Island, Swamp, or Mountain card, put it onto the battlefield tapped, then shuffle.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Search your library for a Plains, Island, Swamp, or Mountain card, put it onto the battlefield tapped, then shuffle.",
			Effects: []game.Effect{
				{
					Type:        game.EffectSearch,
					TargetIndex: game.TargetIndexController,
					Search: opt.Val(game.SearchSpec{
						SourceZone:   game.ZoneLibrary,
						Destination:  game.ZoneBattlefield,
						CardType:     opt.Val(types.Land),
						SubtypesAny:  []types.Sub{types.Plains, types.Island, types.Swamp, types.Mountain},
						EntersTapped: true,
						Shuffle:      true,
					}),
				},
			},
		},
	}}, ColorIdentity: color.NewIdentity(color.Green),
}
