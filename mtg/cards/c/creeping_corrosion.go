package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CreepingCorrosion is the card definition for Creeping Corrosion.
//
// Type: Sorcery
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Destroy all artifacts.
var CreepingCorrosion = newCreepingCorrosion

func newCreepingCorrosion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Creeping Corrosion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all artifacts.
		`,
		},
	}
}
