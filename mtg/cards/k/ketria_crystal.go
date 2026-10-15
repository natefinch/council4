package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KetriaCrystal is the card definition for Ketria Crystal.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {G}, {U}, or {R}.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var KetriaCrystal = newKetriaCrystal

func newKetriaCrystal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Ketria Crystal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U, mana.R),
			},
			OracleText: `
			{T}: Add {G}, {U}, or {R}.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
