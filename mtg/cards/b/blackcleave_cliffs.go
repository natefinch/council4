package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlackcleaveCliffs is the card definition for Blackcleave Cliffs.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or fewer other lands.
//	{T}: Add {B} or {R}.
var BlackcleaveCliffs = newBlackcleaveCliffs

func newBlackcleaveCliffs() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Blackcleave Cliffs",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
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
			{T}: Add {B} or {R}.
		`,
		},
	}
}
