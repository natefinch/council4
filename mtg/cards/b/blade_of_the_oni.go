package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BladeOfTheOni is the card definition for Blade of the Oni.
//
// Type: Artifact Creature — Equipment Demon
// Cost: {1}{B}
//
// Oracle text:
//
//	Menace
//	Equipped creature has base power and toughness 5/5, has menace, and is a black Demon in addition to its other colors and types.
//	Reconfigure {2}{B}{B} ({2}{B}{B}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
var BladeOfTheOni = newBladeOfTheOni

func newBladeOfTheOni() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blade of the Oni",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Equipment, types.Demon},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.StaticAbility{
					Text: "Equipped creature has base power and toughness 5/5, has menace, and is a black Demon in addition to its other colors and types.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 5,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 5,
							}),
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Menace},
						},
						game.ContinuousEffect{
							Duration:  game.DurationPermanent,
							Group:     game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:     game.LayerColor,
							AddColors: []color.Color{color.Black},
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerType,
							AddSubtypes: []types.Sub{types.Demon},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ReconfigureActivatedAbility(cost.Mana{cost.O(2), cost.B, cost.B}),
			},
			OracleText: `
			Menace
			Equipped creature has base power and toughness 5/5, has menace, and is a black Demon in addition to its other colors and types.
			Reconfigure {2}{B}{B} ({2}{B}{B}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
		`,
		},
	}
}
