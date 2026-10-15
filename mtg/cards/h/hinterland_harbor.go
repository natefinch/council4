package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HinterlandHarbor is the card definition for Hinterland Harbor.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Forest or an Island.
//	{T}: Add {G} or {U}.
var HinterlandHarbor = newHinterlandHarbor

func newHinterlandHarbor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:  "Hinterland Harbor",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Forest or an Island.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Sub("Forest"), types.Sub("Island")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Forest or an Island.
			{T}: Add {G} or {U}.
		`,
		},
	}
}
