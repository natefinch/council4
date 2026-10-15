package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// VaultOfWhispers is the card definition for Vault of Whispers.
//
// Type: Artifact Land
//
// Oracle text:
//
//	{T}: Add {B}.
var VaultOfWhispers = newVaultOfWhispers

func newVaultOfWhispers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:  "Vault of Whispers",
			Types: []types.Card{types.Artifact, types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			OracleText: `
			{T}: Add {B}.
		`,
		},
	}
}
