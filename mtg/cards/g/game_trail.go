package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// GameTrail is the card definition for Game Trail.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Mountain or Forest card from your hand. If you don't, this land enters tapped.
//	{T}: Add {R} or {G}.
var GameTrail = newGameTrail

func newGameTrail() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:  "Game Trail",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Mountain or Forest card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Mountain, types.Forest},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Mountain or Forest card from your hand. If you don't, this land enters tapped.
			{T}: Add {R} or {G}.
		`,
		},
	}
}
