package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VineKami is the card definition for Vine Kami.
//
// Type: Creature — Spirit
// Cost: {6}{G}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	Soulshift 6 (When this creature dies, you may return target Spirit card with mana value 6 or less from your graveyard to your hand.)
var VineKami = newVineKami

func newVineKami() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vine Kami",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.SoulshiftTriggeredAbility(6),
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			Soulshift 6 (When this creature dies, you may return target Spirit card with mana value 6 or less from your graveyard to your hand.)
		`,
		},
	}
}
