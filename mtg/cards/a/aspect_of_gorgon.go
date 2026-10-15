package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AspectOfGorgon is the card definition for Aspect of Gorgon.
//
// Type: Enchantment — Aura
// Cost: {2}{B}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature gets +1/+3 and has deathtouch. (Any amount of damage it deals to a creature is enough to destroy it.)
var AspectOfGorgon = newAspectOfGorgon

func newAspectOfGorgon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Aspect of Gorgon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					Text: "Enchanted creature gets +1/+3 and has deathtouch. (Any amount of damage it deals to a creature is enough to destroy it.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 3,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Deathtouch},
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature gets +1/+3 and has deathtouch. (Any amount of damage it deals to a creature is enough to destroy it.)
		`,
		},
	}
}
