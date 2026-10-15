package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RobeOfMirrors is the card definition for Robe of Mirrors.
//
// Type: Enchantment — Aura
// Cost: {U}
//
// Oracle text:
//
//	Enchant creature (Target a creature as you cast this. This card enters attached to that creature.)
//	Enchanted creature has shroud. (It can't be the target of spells or abilities.)
var RobeOfMirrors = newRobeOfMirrors

func newRobeOfMirrors() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Robe of Mirrors",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
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
					Text: "Enchanted creature has shroud. (It can't be the target of spells or abilities.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Shroud},
						},
					},
				},
			},
			OracleText: `
			Enchant creature (Target a creature as you cast this. This card enters attached to that creature.)
			Enchanted creature has shroud. (It can't be the target of spells or abilities.)
		`,
		},
	}
}
