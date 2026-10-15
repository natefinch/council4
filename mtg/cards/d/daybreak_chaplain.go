package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DaybreakChaplain is the card definition for Daybreak Chaplain.
//
// Type: Creature — Human Cleric
// Cost: {1}{W}
//
// Oracle text:
//
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var DaybreakChaplain = newDaybreakChaplain

func newDaybreakChaplain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Daybreak Chaplain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			OracleText: `
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
