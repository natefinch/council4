package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InspiringVantage is the card definition for Inspiring Vantage.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or fewer other lands.
//	{T}: Add {R} or {W}.
var InspiringVantage = newInspiringVantage

func newInspiringVantage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:  "Inspiring Vantage",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
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
			{T}: Add {R} or {W}.
		`,
		},
	}
}
