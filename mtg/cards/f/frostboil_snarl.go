package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// FrostboilSnarl is the card definition for Frostboil Snarl.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal an Island or Mountain card from your hand. If you don't, this land enters tapped.
//	{T}: Add {U} or {R}.
var FrostboilSnarl = newFrostboilSnarl

func newFrostboilSnarl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:  "Frostboil Snarl",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal an Island or Mountain card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Island, types.Mountain},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal an Island or Mountain card from your hand. If you don't, this land enters tapped.
			{T}: Add {U} or {R}.
		`,
		},
	}
}
