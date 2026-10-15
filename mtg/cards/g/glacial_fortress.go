package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlacialFortress is the card definition for Glacial Fortress.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Plains or an Island.
//	{T}: Add {W} or {U}.
var GlacialFortress = newGlacialFortress

func newGlacialFortress() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:  "Glacial Fortress",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Plains or an Island.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Sub("Plains"), types.Sub("Island")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Plains or an Island.
			{T}: Add {W} or {U}.
		`,
		},
	}
}
