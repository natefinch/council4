package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Giant Wizard
//
// Type: Token Creature — Giant Wizard
//
// Oracle text:

// GiantWizardTokenaa2d34c15e8241de9cc90f17d1537b5b is the card definition for Giant Wizard.
var GiantWizardTokenaa2d34c15e8241de9cc90f17d1537b5b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Giant Wizard",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Giant, types.Wizard},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
