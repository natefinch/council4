package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IchorspitBasilisk is the card definition for Ichorspit Basilisk.
//
// Type: Creature — Phyrexian Basilisk
// Cost: {2}{G}
//
// Oracle text:
//
//	Deathtouch
//	Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
var IchorspitBasilisk = newIchorspitBasilisk

func newIchorspitBasilisk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ichorspit Basilisk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Basilisk},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.StaticAbility{
					Text: "Toxic 1",
					KeywordAbilities: []game.KeywordAbility{
						game.ToxicKeyword{
							Amount: 1,
						},
					},
				},
			},
			OracleText: `
			Deathtouch
			Toxic 1 (Players dealt combat damage by this creature also get a poison counter.)
		`,
		},
	}
}
