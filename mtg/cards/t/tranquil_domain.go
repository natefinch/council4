package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TranquilDomain is the card definition for Tranquil Domain.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Destroy all non-Aura enchantments.
var TranquilDomain = newTranquilDomain

func newTranquilDomain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tranquil Domain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}, ExcludedSubtype: types.Sub("Aura")}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all non-Aura enchantments.
		`,
		},
	}
}
