package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PromisedKannushi is the card definition for Promised Kannushi.
//
// Type: Creature — Human Druid
// Cost: {G}
//
// Oracle text:
//
//	Soulshift 7 (When this creature dies, you may return target Spirit card with mana value 7 or less from your graveyard to your hand.)
var PromisedKannushi = newPromisedKannushi

func newPromisedKannushi() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Promised Kannushi",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.SoulshiftTriggeredAbility(7),
			},
			OracleText: `
			Soulshift 7 (When this creature dies, you may return target Spirit card with mana value 7 or less from your graveyard to your hand.)
		`,
		},
	}
}
