package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZagothCrystal is the card definition for Zagoth Crystal.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {B}, {G}, or {U}.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var ZagothCrystal = newZagothCrystal

func newZagothCrystal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Zagoth Crystal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G, mana.U),
			},
			OracleText: `
			{T}: Add {B}, {G}, or {U}.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
