package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GildedLotus is the card definition for Gilded Lotus.
//
// Type: Artifact
// Cost: {5}
//
// Oracle text:
//
//	{T}: Add three mana of any one color.
var GildedLotus = newGildedLotus

func newGildedLotus() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Gilded Lotus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceCountAbility("{T}: Add three mana of any one color.", 3, mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			{T}: Add three mana of any one color.
		`,
		},
	}
}
