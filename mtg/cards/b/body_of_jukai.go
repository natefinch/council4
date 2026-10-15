package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BodyOfJukai is the card definition for Body of Jukai.
//
// Type: Creature — Spirit
// Cost: {7}{G}{G}
//
// Oracle text:
//
//	Trample
//	Soulshift 8 (When this creature dies, you may return target Spirit card with mana value 8 or less from your graveyard to your hand.)
var BodyOfJukai = newBodyOfJukai

func newBodyOfJukai() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Body of Jukai",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.SoulshiftTriggeredAbility(8),
			},
			OracleText: `
			Trample
			Soulshift 8 (When this creature dies, you may return target Spirit card with mana value 8 or less from your graveyard to your hand.)
		`,
		},
	}
}
