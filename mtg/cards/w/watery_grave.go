package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// WateryGrave is the card definition for Watery Grave.
//
// Type: Land — Island Swamp
//
// Oracle text:
//
//	({T}: Add {U} or {B}.)
//	As this land enters, you may pay 2 life. If you don't, it enters tapped.
var WateryGrave = newWateryGrave

func newWateryGrave() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:     "Watery Grave",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
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
			({T}: Add {U} or {B}.)
			As this land enters, you may pay 2 life. If you don't, it enters tapped.
		`,
		},
	}
}
