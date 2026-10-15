package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LotusGuardian is the card definition for Lotus Guardian.
//
// Type: Artifact Creature — Dragon
// Cost: {7}
//
// Oracle text:
//
//	Flying
//	{T}: Add one mana of any color.
var LotusGuardian = newLotusGuardian

func newLotusGuardian() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Lotus Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			Flying
			{T}: Add one mana of any color.
		`,
		},
	}
}
