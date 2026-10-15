package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TempleGarden is the card definition for Temple Garden.
//
// Type: Land — Forest Plains
//
// Oracle text:
//
//	({T}: Add {G} or {W}.)
//	As this land enters, you may pay 2 life. If you don't, it enters tapped.
var TempleGarden = newTempleGarden

func newTempleGarden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:     "Temple Garden",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
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
			({T}: Add {G} or {W}.)
			As this land enters, you may pay 2 life. If you don't, it enters tapped.
		`,
		},
	}
}
