package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// NecroblossomSnarl is the card definition for Necroblossom Snarl.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Swamp or Forest card from your hand. If you don't, this land enters tapped.
//	{T}: Add {B} or {G}.
var NecroblossomSnarl = newNecroblossomSnarl

func newNecroblossomSnarl() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:  "Necroblossom Snarl",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Swamp or Forest card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Swamp, types.Forest},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Swamp or Forest card from your hand. If you don't, this land enters tapped.
			{T}: Add {B} or {G}.
		`,
		},
	}
}
