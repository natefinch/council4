package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlassOfTheGuildpact is the card definition for Glass of the Guildpact.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Multicolored creatures you control get +1/+1.
var GlassOfTheGuildpact = newGlassOfTheGuildpact

func newGlassOfTheGuildpact() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Glass of the Guildpact",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Multicolored creatures you control get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Multicolored: true}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Multicolored creatures you control get +1/+1.
		`,
		},
	}
}
