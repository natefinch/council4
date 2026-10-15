package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SecludedSteppe is the card definition for Secluded Steppe.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {W}.
//	Cycling {W} ({W}, Discard this card: Draw a card.)
var SecludedSteppe = newSecludedSteppe

func newSecludedSteppe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:  "Secluded Steppe",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.W}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {W}.
			Cycling {W} ({W}, Discard this card: Draw a card.)
		`,
		},
	}
}
