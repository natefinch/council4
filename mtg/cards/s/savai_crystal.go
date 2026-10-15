package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SavaiCrystal is the card definition for Savai Crystal.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {R}, {W}, or {B}.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var SavaiCrystal = newSavaiCrystal

func newSavaiCrystal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Savai Crystal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W, mana.B),
			},
			OracleText: `
			{T}: Add {R}, {W}, or {B}.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
