package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NatureSLore is the card definition for Nature's Lore.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Search your library for a Forest card, put that card onto the battlefield, then shuffle.
var NatureSLore = &game.CardDef{
	Name: "Nature's Lore",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(1),
		mana.ColoredMana(mana.Green),
	}),
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []types.Card{types.Sorcery},
	OracleText:    "Search your library for a Forest card, put that card onto the battlefield, then shuffle.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.SpellAbility,
			Text: "Search your library for a Forest card, put that card onto the battlefield, then shuffle.",
			Effects: []game.Effect{
				{
					Type:        game.EffectSearch,
					TargetIndex: game.TargetIndexController,
					Search: opt.Val(game.SearchSpec{
						SourceZone:  game.ZoneLibrary,
						Destination: game.ZoneBattlefield,
						CardType:    opt.Val(types.Land),
						SubtypesAny: []types.Sub{types.Forest},
						Shuffle:     true,
					}),
				},
			},
		},
	},
}
