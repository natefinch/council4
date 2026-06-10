package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human
//
// Type: Token Creature — Human
//
// Oracle text:
//   Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)

// HumanTokendad08fbf8765467a995297f65664401c is the card definition for Human.
var HumanTokendad08fbf8765467a995297f65664401c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Human",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.WardStaticAbility(cost.Mana{cost.O(2)}),
		},
		OracleText: `
			Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
		`,
	},
}
