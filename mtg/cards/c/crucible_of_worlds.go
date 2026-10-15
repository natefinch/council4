package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrucibleOfWorlds is the card definition for Crucible of Worlds.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	You may play lands from your graveyard.
var CrucibleOfWorlds = newCrucibleOfWorlds

func newCrucibleOfWorlds() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Crucible of Worlds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.PlayLandsFromGraveyardStaticBody,
			},
			OracleText: `
			You may play lands from your graveyard.
		`,
		},
	}
}
