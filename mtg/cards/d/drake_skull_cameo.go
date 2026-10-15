package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DrakeSkullCameo is the card definition for Drake-Skull Cameo.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {U} or {B}.
var DrakeSkullCameo = newDrakeSkullCameo

func newDrakeSkullCameo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Drake-Skull Cameo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			OracleText: `
			{T}: Add {U} or {B}.
		`,
		},
	}
}
