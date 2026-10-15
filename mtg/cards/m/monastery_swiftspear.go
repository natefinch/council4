package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MonasterySwiftspear is the card definition for Monastery Swiftspear.
//
// Type: Creature — Human Monk
// Cost: {R}
//
// Oracle text:
//
//	Haste
//	Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
var MonasterySwiftspear = newMonasterySwiftspear

func newMonasterySwiftspear() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Monastery Swiftspear",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Monk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.ProwessStaticBody,
			},
			OracleText: `
			Haste
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
		},
	}
}
