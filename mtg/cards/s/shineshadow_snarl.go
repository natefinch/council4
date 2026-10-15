package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// ShineshadowSnarl is the card definition for Shineshadow Snarl.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Plains or Swamp card from your hand. If you don't, this land enters tapped.
//	{T}: Add {W} or {B}.
var ShineshadowSnarl = newShineshadowSnarl

func newShineshadowSnarl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:  "Shineshadow Snarl",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Plains or Swamp card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Plains, types.Swamp},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Plains or Swamp card from your hand. If you don't, this land enters tapped.
			{T}: Add {W} or {B}.
		`,
		},
	}
}
