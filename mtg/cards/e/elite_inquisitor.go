package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EliteInquisitor is the card definition for Elite Inquisitor.
//
// Type: Creature — Human Soldier
// Cost: {W}{W}
//
// Oracle text:
//
//	First strike, vigilance
//	Protection from Vampires, from Werewolves, and from Zombies
var EliteInquisitor = newEliteInquisitor

func newEliteInquisitor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Elite Inquisitor",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.VigilanceStaticBody,
				game.ProtectionFromSubtypesStaticAbility(types.Vampire, types.Werewolf, types.Zombie),
			},
			OracleText: `
			First strike, vigilance
			Protection from Vampires, from Werewolves, and from Zombies
		`,
		},
	}
}
