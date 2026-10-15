package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImprisonedInTheMoon is the card definition for Imprisoned in the Moon.
//
// Type: Enchantment — Aura
// Cost: {2}{U}
//
// Oracle text:
//
//	Enchant creature, land, or planeswalker
//	Enchanted permanent is a colorless land with "{T}: Add {C}" and loses all other card types and abilities.
var ImprisonedInTheMoon = newImprisonedInTheMoon

func newImprisonedInTheMoon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Imprisoned in the Moon",
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
					Constraint: "creature or land or planeswalker",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land, types.Planeswalker}}),
				}),
				game.StaticAbility{
					Text: "Enchanted permanent is a colorless land with \"{T}: Add {C}\" and loses all other card types and abilities.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:           game.DurationPermanent,
							Group:              game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:              game.LayerAbility,
							RemoveAllAbilities: true,
						},
						game.ContinuousEffect{
							Duration:     game.DurationPermanent,
							Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:        game.LayerColor,
							SetColorless: true,
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerType,
							SetTypes: []types.Card{types.Land},
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								new(game.TapManaAbility(mana.C)),
							},
						},
					},
				},
			},
			OracleText: `
			Enchant creature, land, or planeswalker
			Enchanted permanent is a colorless land with "{T}: Add {C}" and loses all other card types and abilities.
		`,
		},
	}
}
