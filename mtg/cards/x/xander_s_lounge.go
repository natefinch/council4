package x

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// XanderSLounge is the card definition for Xander's Lounge.
//
// Type: Land — Island Swamp Mountain
//
// Oracle text:
//
//	({T}: Add {U}, {B}, or {R}.)
//	This land enters tapped.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var XanderSLounge = newXanderSLounge

func newXanderSLounge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name:     "Xander's Lounge",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Swamp, types.Mountain},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {U}, {B}, or {R}.)
			This land enters tapped.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
