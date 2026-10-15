package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelMyr is the card definition for Darksteel Myr.
//
// Type: Artifact Creature — Myr
// Cost: {3}
//
// Oracle text:
//
//	Indestructible (Damage and effects that say "destroy" don't destroy this creature. If its toughness is 0 or less, it still dies.)
var DarksteelMyr = newDarksteelMyr

func newDarksteelMyr() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Myr",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			OracleText: `
			Indestructible (Damage and effects that say "destroy" don't destroy this creature. If its toughness is 0 or less, it still dies.)
		`,
		},
	}
}
