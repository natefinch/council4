package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GhostSpectralSaboteur is the card definition for Ghost, Spectral Saboteur.
//
// Type: Legendary Creature — Human Rogue Villain
// Cost: {2}{U/B}
//
// Oracle text:
//
//	Flash
//	Intangibility — Ghost can't be blocked.
var GhostSpectralSaboteur = newGhostSpectralSaboteur

func newGhostSpectralSaboteur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Ghost, Spectral Saboteur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue, types.Villain},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.CantBeBlockedStaticBody,
			},
			OracleText: `
			Flash
			Intangibility — Ghost can't be blocked.
		`,
		},
	}
}
