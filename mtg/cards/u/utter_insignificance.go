package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UtterInsignificance is the card definition for Utter Insignificance.
//
// Type: Enchantment — Aura
// Cost: {1}{U}
//
// Oracle text:
//
//	Flash
//	Enchant creature
//	Enchanted creature loses all abilities and has base power and toughness 1/1.
//	{2}{C}: Exile enchanted creature.
var UtterInsignificance = newUtterInsignificance()

func newUtterInsignificance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Utter Insignificance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:              game.LayerAbility,
							Group:              game.AttachedObjectGroup(game.SourcePermanentReference()),
							RemoveAllAbilities: true,
						},
						game.ContinuousEffect{
							Layer:        game.LayerPowerToughnessSet,
							Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
							SetPower:     opt.Val(game.PT{Value: 1}),
							SetToughness: opt.Val(game.PT{Value: 1}),
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{C}: Exile enchanted creature.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.C}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object: game.SourceAttachedPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			Enchant creature
			Enchanted creature loses all abilities and has base power and toughness 1/1.
			{2}{C}: Exile enchanted creature.
		`,
		},
	}
}
