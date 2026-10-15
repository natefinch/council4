package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// JetmirSGarden is the card definition for Jetmir's Garden.
//
// Type: Land — Mountain Forest Plains
//
// Oracle text:
//
//	({T}: Add {R}, {G}, or {W}.)
//	This land enters tapped.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var JetmirSGarden = newJetmirSGarden

func newJetmirSGarden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name:     "Jetmir's Garden",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Forest, types.Plains},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {R}, {G}, or {W}.)
			This land enters tapped.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
