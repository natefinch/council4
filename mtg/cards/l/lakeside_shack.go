package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// LakesideShack is the card definition for Lakeside Shack.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless a player has 13 or less life.
//	{T}: Add {G} or {U}.
var LakesideShack = newLakesideShack

func newLakesideShack() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:  "Lakeside Shack",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless a player has 13 or less life.", &game.Condition{
					Negate:              true,
					AnyPlayerLifeAtMost: 13,
				}),
			},
			OracleText: `
			This land enters tapped unless a player has 13 or less life.
			{T}: Add {G} or {U}.
		`,
		},
	}
}
