package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// OvergrownTomb is the card definition for Overgrown Tomb.
//
// Type: Land — Swamp Forest
//
// Oracle text:
//
//	({T}: Add {B} or {G}.)
//	As this land enters, you may pay 2 life. If you don't, it enters tapped.
var OvergrownTomb = newOvergrownTomb

func newOvergrownTomb() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:     "Overgrown Tomb",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp, types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedUnlessPaidReplacement("As this land enters, you may pay 2 life. If you don't, it enters tapped.", game.ResolutionPayment{
					Prompt: "Pay 2 life?",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Amount: 2,
						},
					},
				}),
			},
			OracleText: `
			({T}: Add {B} or {G}.)
			As this land enters, you may pay 2 life. If you don't, it enters tapped.
		`,
		},
	}
}
