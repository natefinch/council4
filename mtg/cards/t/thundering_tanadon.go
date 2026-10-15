package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThunderingTanadon is the card definition for Thundering Tanadon.
//
// Type: Artifact Creature — Phyrexian Beast
// Cost: {4}{G/P}{G/P}
//
// Oracle text:
//
//	({G/P} can be paid with either {G} or 2 life.)
//	Trample
var ThunderingTanadon = newThunderingTanadon

func newThunderingTanadon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Thundering Tanadon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.PhyrexianMana(mana.G),
				cost.PhyrexianMana(mana.G),
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			({G/P} can be paid with either {G} or 2 life.)
			Trample
		`,
		},
	}
}
