package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// UmbralExpanse is the card definition for Umbral Expanse.
//
// Type: Land — Plains Swamp
//
// Oracle text:
//
//	({T}: Add {W} or {B}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var UmbralExpanse = newUmbralExpanse

func newUmbralExpanse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:     "Umbral Expanse",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Swamp},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {W} or {B}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
