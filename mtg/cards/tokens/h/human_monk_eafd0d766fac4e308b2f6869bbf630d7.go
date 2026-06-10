package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human Monk
//
// Type: Token Creature — Human Monk
//
// Oracle text:
//   {T}: Add {G}.

// HumanMonkTokeneafd0d766fac4e308b2f6869bbf630d7 is the card definition for Human Monk.
var HumanMonkTokeneafd0d766fac4e308b2f6869bbf630d7 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Human Monk",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Monk},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ManaAbilities: []game.ManaAbility{
			game.TapManaAbility(mana.G),
		},
		OracleText: `
			{T}: Add {G}.
		`,
	},
}
