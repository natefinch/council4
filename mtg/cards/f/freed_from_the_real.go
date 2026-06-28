package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FreedFromTheReal is the card definition for Freed from the Real.
//
// Type: Enchantment — Aura
// Cost: {2}{U}
//
// Oracle text:
//
//	Enchant creature
//	{U}: Tap enchanted creature.
//	{U}: Untap enchanted creature.
var FreedFromTheReal = newFreedFromTheReal()

func newFreedFromTheReal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Freed from the Real",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
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
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{U}: Tap enchanted creature.",
					ManaCost:       opt.Val(cost.Mana{cost.U}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.SourceAttachedPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{U}: Untap enchanted creature.",
					ManaCost:       opt.Val(cost.Mana{cost.U}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.SourceAttachedPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			{U}: Tap enchanted creature.
			{U}: Untap enchanted creature.
		`,
		},
	}
}
