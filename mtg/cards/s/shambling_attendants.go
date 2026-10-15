package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShamblingAttendants is the card definition for Shambling Attendants.
//
// Type: Creature — Zombie
// Cost: {7}{B}
//
// Oracle text:
//
//	Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var ShamblingAttendants = newShamblingAttendants

func newShamblingAttendants() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Shambling Attendants",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DelveStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
			Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
		`,
		},
	}
}
