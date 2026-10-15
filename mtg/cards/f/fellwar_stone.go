package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FellwarStone is the card definition for Fellwar Stone.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	{T}: Add one mana of any color that a land an opponent controls could produce.
var FellwarStone = newFellwarStone

func newFellwarStone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Fellwar Stone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaLandsProduceAbility(game.PlayerOpponent, false),
			},
			OracleText: `
			{T}: Add one mana of any color that a land an opponent controls could produce.
		`,
		},
	}
}
