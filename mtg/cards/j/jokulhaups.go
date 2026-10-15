package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Jokulhaups is the card definition for Jokulhaups.
//
// Type: Sorcery
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Destroy all artifacts, creatures, and lands. They can't be regenerated.
var Jokulhaups = newJokulhaups

func newJokulhaups() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Jokulhaups",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group:               game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}}),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all artifacts, creatures, and lands. They can't be regenerated.
		`,
		},
	}
}
