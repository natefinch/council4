package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VampireNighthawk is the card definition for Vampire Nighthawk.
//
// Type: Creature — Vampire Shaman
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Flying
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var VampireNighthawk = newVampireNighthawk

func newVampireNighthawk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vampire Nighthawk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DeathtouchStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Flying
			Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
