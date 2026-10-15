package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheirNameIsDeath is the card definition for Their Name Is Death.
//
// Type: Sorcery
// Cost: {3}{B}{B}{B}
//
// Oracle text:
//
//	Destroy all nonartifact creatures.
var TheirNameIsDeath = newTheirNameIsDeath

func newTheirNameIsDeath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Their Name Is Death",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedTypes: []types.Card{types.Artifact}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all nonartifact creatures.
		`,
		},
	}
}
