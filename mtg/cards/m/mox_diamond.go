package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MoxDiamond is the card definition for Mox Diamond.
//
// Type: Artifact
// Cost: {0}
//
// Oracle text:
//
//	If this artifact would enter, you may discard a land card instead. If you do, put this artifact onto the battlefield. If you don't, put it into its owner's graveyard.
//	{T}: Add one mana of any color.
var MoxDiamond = newMoxDiamond

func newMoxDiamond() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mox Diamond",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersUnlessPaidElseZoneReplacement("If this artifact would enter, you may discard a land card instead. If you do, put this artifact onto the battlefield. If you don't, put it into its owner's graveyard.", game.ResolutionPayment{
					Prompt: "Pay the alternative cost?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalDiscard,
							Amount:        1,
							Source:        zone.Hand,
							MatchCardType: true,
							CardType:      types.Land,
						},
					},
				}, zone.Graveyard),
			},
			OracleText: `
			If this artifact would enter, you may discard a land card instead. If you do, put this artifact onto the battlefield. If you don't, put it into its owner's graveyard.
			{T}: Add one mana of any color.
		`,
		},
	}
}
