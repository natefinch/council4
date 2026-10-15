package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeashellCameo is the card definition for Seashell Cameo.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {W} or {U}.
var SeashellCameo = newSeashellCameo

func newSeashellCameo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Seashell Cameo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			OracleText: `
			{T}: Add {W} or {U}.
		`,
		},
	}
}
