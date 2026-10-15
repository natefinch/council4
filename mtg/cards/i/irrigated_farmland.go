package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// IrrigatedFarmland is the card definition for Irrigated Farmland.
//
// Type: Land — Plains Island
//
// Oracle text:
//
//	({T}: Add {W} or {U}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var IrrigatedFarmland = newIrrigatedFarmland

func newIrrigatedFarmland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:     "Irrigated Farmland",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Island},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {W} or {U}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
