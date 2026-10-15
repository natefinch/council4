package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// FurycalmSnarl is the card definition for Furycalm Snarl.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Mountain or Plains card from your hand. If you don't, this land enters tapped.
//	{T}: Add {R} or {W}.
var FurycalmSnarl = newFurycalmSnarl

func newFurycalmSnarl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:  "Furycalm Snarl",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Mountain or Plains card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Mountain, types.Plains},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Mountain or Plains card from your hand. If you don't, this land enters tapped.
			{T}: Add {R} or {W}.
		`,
		},
	}
}
