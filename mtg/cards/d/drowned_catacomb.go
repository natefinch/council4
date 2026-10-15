package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DrownedCatacomb is the card definition for Drowned Catacomb.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control an Island or a Swamp.
//	{T}: Add {U} or {B}.
var DrownedCatacomb = newDrownedCatacomb

func newDrownedCatacomb() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:  "Drowned Catacomb",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control an Island or a Swamp.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Sub("Island"), types.Sub("Swamp")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control an Island or a Swamp.
			{T}: Add {U} or {B}.
		`,
		},
	}
}
