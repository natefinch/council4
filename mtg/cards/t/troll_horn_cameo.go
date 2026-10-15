package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TrollHornCameo is the card definition for Troll-Horn Cameo.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {R} or {G}.
var TrollHornCameo = newTrollHornCameo

func newTrollHornCameo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Troll-Horn Cameo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			OracleText: `
			{T}: Add {R} or {G}.
		`,
		},
	}
}
