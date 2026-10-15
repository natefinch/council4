package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// PortTown is the card definition for Port Town.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Plains or Island card from your hand. If you don't, this land enters tapped.
//	{T}: Add {W} or {U}.
var PortTown = newPortTown

func newPortTown() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:  "Port Town",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Plains or Island card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Plains, types.Island},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Plains or Island card from your hand. If you don't, this land enters tapped.
			{T}: Add {W} or {U}.
		`,
		},
	}
}
