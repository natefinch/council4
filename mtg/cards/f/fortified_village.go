package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// FortifiedVillage is the card definition for Fortified Village.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Forest or Plains card from your hand. If you don't, this land enters tapped.
//	{T}: Add {G} or {W}.
var FortifiedVillage = newFortifiedVillage

func newFortifiedVillage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:  "Fortified Village",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Forest or Plains card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Forest, types.Plains},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Forest or Plains card from your hand. If you don't, this land enters tapped.
			{T}: Add {G} or {W}.
		`,
		},
	}
}
