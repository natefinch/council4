package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Obliterate is the card definition for Obliterate.
//
// Type: Sorcery
// Cost: {6}{R}{R}
//
// Oracle text:
//
//	This spell can't be countered.
//	Destroy all artifacts, creatures, and lands. They can't be regenerated.
var Obliterate = newObliterate

func newObliterate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Obliterate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
			},
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
			This spell can't be countered.
			Destroy all artifacts, creatures, and lands. They can't be regenerated.
		`,
		},
	}
}
