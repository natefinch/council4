package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// VineglimmerSnarl is the card definition for Vineglimmer Snarl.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Forest or Island card from your hand. If you don't, this land enters tapped.
//	{T}: Add {G} or {U}.
var VineglimmerSnarl = newVineglimmerSnarl

func newVineglimmerSnarl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:  "Vineglimmer Snarl",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Forest or Island card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Forest, types.Island},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Forest or Island card from your hand. If you don't, this land enters tapped.
			{T}: Add {G} or {U}.
		`,
		},
	}
}
