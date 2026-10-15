package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WilsonRefinedGrizzly is the card definition for Wilson, Refined Grizzly.
//
// Type: Legendary Creature — Bear Warrior
// Cost: {1}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Reach, vigilance, trample
//	Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
//	Choose a Background (You can have a Background as a second commander.)
var WilsonRefinedGrizzly = newWilsonRefinedGrizzly

func newWilsonRefinedGrizzly() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wilson, Refined Grizzly",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Bear, types.Warrior},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.ReachStaticBody,
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
				game.ChooseABackgroundStaticBody,
			},
			OracleText: `
			This spell can't be countered.
			Reach, vigilance, trample
			Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
			Choose a Background (You can have a Background as a second commander.)
		`,
		},
	}
}
