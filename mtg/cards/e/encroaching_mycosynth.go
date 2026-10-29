package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EncroachingMycosynth is the card definition for Encroaching Mycosynth.
//
// Type: Artifact
// Cost: {3}{U}
//
// Oracle text:
//
//	Nonland permanents you control are artifacts in addition to their other types. The same is true for permanent spells you control and nonland permanent cards you own that aren't on the battlefield.
var EncroachingMycosynth = newEncroachingMycosynth

func newEncroachingMycosynth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Encroaching Mycosynth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Nonland permanents you control are artifacts in addition to their other types. The same is true for permanent spells you control and nonland permanent cards you own that aren't on the battlefield.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{ExcludedTypes: []types.Card{types.Land}}),
							Layer:    game.LayerType,
							SpellSelection: &game.Selection{RequiredTypesAny: []types.Card{
								types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker,
							}},
							AddTypes: []types.Card{types.Artifact},
						},
					},
				},
			},
			OracleText: `
			Nonland permanents you control are artifacts in addition to their other types. The same is true for permanent spells you control and nonland permanent cards you own that aren't on the battlefield.
		`,
		},
	}
}
