package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SokkaWolfCoveSProtector is the card definition for Sokka, Wolf Cove's Protector.
//
// Type: Legendary Creature — Human Warrior Ally
// Cost: {2}{W}
//
// Oracle text:
//
//	Vigilance (Attacking doesn't cause this creature to tap.)
var SokkaWolfCoveSProtector = newSokkaWolfCoveSProtector

func newSokkaWolfCoveSProtector() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sokka, Wolf Cove's Protector",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior, types.Ally},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			OracleText: `
			Vigilance (Attacking doesn't cause this creature to tap.)
		`,
		},
	}
}
