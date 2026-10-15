package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarkslickShores is the card definition for Darkslick Shores.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or fewer other lands.
//	{T}: Add {U} or {B}.
var DarkslickShores = newDarkslickShores

func newDarkslickShores() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:  "Darkslick Shores",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control two or fewer other lands.", &game.Condition{
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, ExcludeSource: true},
						MinCount:  3,
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control two or fewer other lands.
			{T}: Add {U} or {B}.
		`,
		},
	}
}
