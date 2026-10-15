package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelSentinel is the card definition for Darksteel Sentinel.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Vigilance
//	Indestructible (Damage and effects that say "destroy" don't destroy this creature. If its toughness is 0 or less, it's still put into its owner's graveyard.)
var DarksteelSentinel = newDarksteelSentinel

func newDarksteelSentinel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Sentinel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.VigilanceStaticBody,
				game.IndestructibleStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Vigilance
			Indestructible (Damage and effects that say "destroy" don't destroy this creature. If its toughness is 0 or less, it's still put into its owner's graveyard.)
		`,
		},
	}
}
