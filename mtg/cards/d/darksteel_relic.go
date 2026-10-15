package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelRelic is the card definition for Darksteel Relic.
//
// Type: Artifact
// Cost: {0}
//
// Oracle text:
//
//	Indestructible (Effects that say "destroy" don't destroy this artifact.)
var DarksteelRelic = newDarksteelRelic

func newDarksteelRelic() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Relic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			OracleText: `
			Indestructible (Effects that say "destroy" don't destroy this artifact.)
		`,
		},
	}
}
