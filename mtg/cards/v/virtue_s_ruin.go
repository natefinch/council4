package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VirtueSRuin is the card definition for Virtue's Ruin.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Destroy all white creatures.
var VirtueSRuin = newVirtueSRuin

func newVirtueSRuin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Virtue's Ruin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.White}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all white creatures.
		`,
		},
	}
}
