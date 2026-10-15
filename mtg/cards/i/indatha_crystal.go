package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IndathaCrystal is the card definition for Indatha Crystal.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {W}, {B}, or {G}.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var IndathaCrystal = newIndathaCrystal

func newIndathaCrystal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Indatha Crystal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B, mana.G),
			},
			OracleText: `
			{T}: Add {W}, {B}, or {G}.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
