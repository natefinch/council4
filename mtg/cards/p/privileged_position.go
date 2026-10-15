package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrivilegedPosition is the card definition for Privileged Position.
//
// Type: Enchantment
// Cost: {2}{G/W}{G/W}{G/W}
//
// Oracle text:
//
//	({G/W} can be paid with either {G} or {W}.)
//	Other permanents you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)
var PrivilegedPosition = newPrivilegedPosition

func newPrivilegedPosition() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Privileged Position",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors: []color.Color{color.Green, color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Other permanents you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{}, game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Hexproof},
						},
					},
				},
			},
			OracleText: `
			({G/W} can be paid with either {G} or {W}.)
			Other permanents you control have hexproof. (They can't be the targets of spells or abilities your opponents control.)
		`,
		},
	}
}
