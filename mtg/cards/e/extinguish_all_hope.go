package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ExtinguishAllHope is the card definition for Extinguish All Hope.
//
// Type: Sorcery
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Destroy all nonenchantment creatures.
var ExtinguishAllHope = newExtinguishAllHope

func newExtinguishAllHope() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Extinguish All Hope",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedTypes: []types.Card{types.Enchantment}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonenchantment creatures.
		`,
		},
	}
}
