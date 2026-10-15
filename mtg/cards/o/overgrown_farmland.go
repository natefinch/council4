package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OvergrownFarmland is the card definition for Overgrown Farmland.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or more other lands.
//	{T}: Add {G} or {W}.
var OvergrownFarmland = newOvergrownFarmland

func newOvergrownFarmland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:  "Overgrown Farmland",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control two or more other lands.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, ExcludeSource: true},
						MinCount:  2,
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control two or more other lands.
			{T}: Add {G} or {W}.
		`,
		},
	}
}
