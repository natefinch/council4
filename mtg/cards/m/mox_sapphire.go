package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoxSapphire is the card definition for Mox Sapphire.
//
// Type: Artifact
// Cost: {0}
//
// Oracle text:
//
//	{T}: Add {U}.
var MoxSapphire = newMoxSapphire

func newMoxSapphire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mox Sapphire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			OracleText: `
			{T}: Add {U}.
		`,
		},
	}
}
