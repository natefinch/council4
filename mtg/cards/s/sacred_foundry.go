package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SacredFoundry is the card definition for Sacred Foundry.
//
// Type: Land — Mountain Plains
//
// Oracle text:
//
//	({T}: Add {R} or {W}.)
//	As this land enters, you may pay 2 life. If you don't, it enters tapped.
var SacredFoundry = newSacredFoundry

func newSacredFoundry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:     "Sacred Foundry",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
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
			({T}: Add {R} or {W}.)
			As this land enters, you may pay 2 life. If you don't, it enters tapped.
		`,
		},
	}
}
