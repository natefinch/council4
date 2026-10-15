package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZhangFeiFierceWarrior is the card definition for Zhang Fei, Fierce Warrior.
//
// Type: Legendary Creature — Human Soldier Warrior
// Cost: {4}{W}{W}
//
// Oracle text:
//
//	Vigilance; horsemanship (This creature can't be blocked except by creatures with horsemanship.)
var ZhangFeiFierceWarrior = newZhangFeiFierceWarrior

func newZhangFeiFierceWarrior() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Zhang Fei, Fierce Warrior",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier, types.Warrior},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.HorsemanshipStaticBody,
			},
			OracleText: `
			Vigilance; horsemanship (This creature can't be blocked except by creatures with horsemanship.)
		`,
		},
	}
}
