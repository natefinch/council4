package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// WanderwineHub is the card definition for Wanderwine Hub.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Merfolk card from your hand. If you don't, this land enters tapped.
//	{T}: Add {W} or {U}.
var WanderwineHub = newWanderwineHub

func newWanderwineHub() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:  "Wanderwine Hub",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Merfolk card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Merfolk},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Merfolk card from your hand. If you don't, this land enters tapped.
			{T}: Add {W} or {U}.
		`,
		},
	}
}
