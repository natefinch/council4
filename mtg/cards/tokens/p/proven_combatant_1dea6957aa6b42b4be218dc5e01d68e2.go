package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Proven Combatant
//
// Type: Token Creature — Zombie Human Warrior
//
// Oracle text:

// ProvenCombatantToken1dea6957aa6b42b4be218dc5e01d68e2 is the card definition for Proven Combatant.
var ProvenCombatantToken1dea6957aa6b42b4be218dc5e01d68e2 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Proven Combatant",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Human, types.Warrior},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
