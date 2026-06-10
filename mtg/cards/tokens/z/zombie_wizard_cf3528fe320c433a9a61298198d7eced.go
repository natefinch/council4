package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie Wizard
//
// Type: Token Creature — Zombie Wizard
//
// Oracle text:

// ZombieWizardTokencf3528fe320c433a9a61298198d7eced is the card definition for Zombie Wizard.
var ZombieWizardTokencf3528fe320c433a9a61298198d7eced = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie Wizard",
		Colors:    []color.Color{color.Black, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Wizard},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
