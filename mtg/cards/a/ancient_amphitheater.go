package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// AncientAmphitheater is the card definition for Ancient Amphitheater.
//
// Type: Land
//
// Oracle text:
//
//	As this land enters, you may reveal a Giant card from your hand. If you don't, this land enters tapped.
//	{T}: Add {R} or {W}.
var AncientAmphitheater = newAncientAmphitheater

func newAncientAmphitheater() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:  "Ancient Amphitheater",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may reveal a Giant card from your hand. If you don't, this land enters tapped.", game.ResolutionPayment{
					Prompt: "Reveal a matching card?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalReveal,
							Amount:      1,
							Source:      zone.Hand,
							SubtypesAny: cost.SubtypeSet{types.Giant},
						},
					},
				}),
			},
			OracleText: `
			As this land enters, you may reveal a Giant card from your hand. If you don't, this land enters tapped.
			{T}: Add {R} or {W}.
		`,
		},
	}
}
