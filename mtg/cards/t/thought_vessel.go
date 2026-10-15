package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThoughtVessel is the card definition for Thought Vessel.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	You have no maximum hand size.
//	{T}: Add {C}.
var ThoughtVessel = newThoughtVessel

func newThoughtVessel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thought Vessel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			You have no maximum hand size.
			{T}: Add {C}.
		`,
		},
	}
}
