package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// GodlessShrine is the card definition for Godless Shrine.
//
// Type: Land — Plains Swamp
//
// Oracle text:
//
//	({T}: Add {W} or {B}.)
//	As this land enters, you may pay 2 life. If you don't, it enters tapped.
var GodlessShrine = newGodlessShrine

func newGodlessShrine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:     "Godless Shrine",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may pay 2 life. If you don't, it enters tapped.", game.ResolutionPayment{
					Prompt: "Pay 2 life?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Amount: 2,
						},
					},
				}),
			},
			OracleText: `
			({T}: Add {W} or {B}.)
			As this land enters, you may pay 2 life. If you don't, it enters tapped.
		`,
		},
	}
}
