package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LadyZhurongWarriorQueen is the card definition for Lady Zhurong, Warrior Queen.
//
// Type: Legendary Creature — Human Soldier Warrior
// Cost: {4}{G}
//
// Oracle text:
//
//	Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
var LadyZhurongWarriorQueen = newLadyZhurongWarriorQueen

func newLadyZhurongWarriorQueen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lady Zhurong, Warrior Queen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier, types.Warrior},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HorsemanshipStaticBody,
			},
			OracleText: `
			Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
		`,
		},
	}
}
