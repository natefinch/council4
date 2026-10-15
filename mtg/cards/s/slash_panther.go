package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlashPanther is the card definition for Slash Panther.
//
// Type: Artifact Creature — Phyrexian Cat
// Cost: {4}{R/P}
//
// Oracle text:
//
//	({R/P} can be paid with either {R} or 2 life.)
//	Haste
var SlashPanther = newSlashPanther

func newSlashPanther() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Slash Panther",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.PhyrexianMana(mana.R),
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Cat},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			OracleText: `
			({R/P} can be paid with either {R} or 2 life.)
			Haste
		`,
		},
	}
}
