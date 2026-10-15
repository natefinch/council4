package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArcaneSignet is the card definition for Arcane Signet.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{T}: Add one mana of any color in your commander's color identity.
var ArcaneSignet = newArcaneSignet

func newArcaneSignet() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Arcane Signet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaCommanderIdentityAbility(),
			},
			OracleText: `
			{T}: Add one mana of any color in your commander's color identity.
		`,
		},
	}
}
