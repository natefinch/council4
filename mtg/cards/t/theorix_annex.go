package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheorixAnnex is the card definition for Theorix Annex.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a planeswalker.
//	{T}: Add {U} or {B}.
var TheorixAnnex = newTheorixAnnex

func newTheorixAnnex() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:  "Theorix Annex",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a planeswalker.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a planeswalker.
			{T}: Add {U} or {B}.
		`,
		},
	}
}
