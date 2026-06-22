package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonskullSummit is the card definition for Dragonskull Summit.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Swamp or a Mountain.
//	{T}: Add {B} or {R}.
var DragonskullSummit = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Dragonskull Summit",
			Types: []types.Card{types.Land},
			OracleText: `
				This land enters tapped unless you control a Swamp or a Mountain.
				{T}: Add {B} or {R}.
			`,
		},
	}
	card.ReplacementAbilities = append(card.ReplacementAbilities,
		game.EntersTappedIfReplacement("This land enters tapped unless you control a Swamp or a Mountain.", &game.Condition{
			Negate: true,
			ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{
					SubtypesAny: []types.Sub{types.Swamp, types.Mountain},
				},
			}),
		}),
	)
	card.ManaAbilities = append(card.ManaAbilities, game.TapManaChoiceAbility(mana.B, mana.R))
	return card
}()
