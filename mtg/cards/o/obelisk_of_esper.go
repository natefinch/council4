package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObeliskOfEsper is the card definition for Obelisk of Esper.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {W}, {U}, or {B}.
var ObeliskOfEsper = newObeliskOfEsper

func newObeliskOfEsper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Obelisk of Esper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B),
			},
			OracleText: `
			{T}: Add {W}, {U}, or {B}.
		`,
		},
	}
}
