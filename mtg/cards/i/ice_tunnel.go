package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// IceTunnel is the card definition for Ice Tunnel.
//
// Type: Snow Land — Island Swamp
//
// Oracle text:
//
//	({T}: Add {U} or {B}.)
//	This land enters tapped.
var IceTunnel = newIceTunnel

func newIceTunnel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:       "Ice Tunnel",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Island, types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {U} or {B}.)
			This land enters tapped.
		`,
		},
	}
}
