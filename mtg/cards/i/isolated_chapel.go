package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IsolatedChapel is the card definition for Isolated Chapel.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Plains or a Swamp.
//	{T}: Add {W} or {B}.
var IsolatedChapel = newIsolatedChapel

func newIsolatedChapel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:  "Isolated Chapel",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Plains or a Swamp.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Sub("Plains"), types.Sub("Swamp")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Plains or a Swamp.
			{T}: Add {W} or {B}.
		`,
		},
	}
}
