package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// AuntieSHovel is the card definition for Auntie's Hovel.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Goblin card from your hand. If you don't, this land enters tapped.
//	{T}: Add {B} or {R}.
var AuntieSHovel = newAuntieSHovel

func newAuntieSHovel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Auntie's Hovel",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Goblin card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Goblin},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Goblin card from your hand. If you don't, this land enters tapped.
			{T}: Add {B} or {R}.
		`,
		},
	}
}
