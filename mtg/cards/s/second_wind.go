package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SecondWind is the card definition for Second Wind.
//
// Type: Enchantment — Aura
// Cost: {2}{U}
//
// Oracle text:
//
//	Enchant creature
//	{T}: Tap enchanted creature.
//	{T}: Untap enchanted creature.
var SecondWind = newSecondWind()

func newSecondWind() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Second Wind",
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
					Text:            "{T}: Tap enchanted creature.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
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
					Text:            "{T}: Untap enchanted creature.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
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
			{T}: Tap enchanted creature.
			{T}: Untap enchanted creature.
		`,
		},
	}
}
