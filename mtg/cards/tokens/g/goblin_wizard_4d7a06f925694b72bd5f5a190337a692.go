package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Goblin Wizard
//
// Type: Token Creature — Goblin Wizard
//
// Oracle text:
//   Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)

// GoblinWizardToken4d7a06f925694b72bd5f5a190337a692 is the card definition for Goblin Wizard.
var GoblinWizardToken4d7a06f925694b72bd5f5a190337a692 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Goblin Wizard",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin, types.Wizard},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.ProwessStaticBody,
		},
		OracleText: `
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
		`,
	},
}
