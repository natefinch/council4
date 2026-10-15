package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RaugrinCrystal is the card definition for Raugrin Crystal.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {U}, {R}, or {W}.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var RaugrinCrystal = newRaugrinCrystal

func newRaugrinCrystal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Raugrin Crystal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R, mana.W),
			},
			OracleText: `
			{T}: Add {U}, {R}, or {W}.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
