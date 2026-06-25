package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HauntedRidge is the card definition for Haunted Ridge.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control two or more other lands.
//	{T}: Add {B} or {R}.
var HauntedRidge = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Haunted Ridge",
			Types: []types.Card{types.Land},
			OracleText: `
				This land enters tapped unless you control two or more other lands.
				{T}: Add {B} or {R}.
			`,
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control two or more other lands.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{
							RequiredTypes: []types.Card{types.Land},
						},
						MinCount: 2,
					}),
				}),
			},
		},
	}

	card.ManaAbilities = append(card.ManaAbilities, game.TapManaChoiceAbility(mana.B, mana.R))
	return card
}()
