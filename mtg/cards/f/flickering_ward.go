package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FlickeringWard is the card definition for Flickering Ward.
//
// Type: Enchantment — Aura
// Cost: {W}
//
// Oracle text:
//
//	Enchant creature
//	As this Aura enters, choose a color.
//	Enchanted creature has protection from the chosen color. This effect doesn't remove this Aura.
//	{W}: Return this Aura to its owner's hand.
var FlickeringWard = newFlickeringWard()

func newFlickeringWard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Flickering Ward",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:   []color.Color{color.White},
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
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ProtectionFromChosenColorStaticAbility()),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{W}: Return this Aura to its owner's hand.",
					ManaCost:       opt.Val(cost.Mana{cost.W}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryColorChoiceReplacement("As this Aura enters, choose a color."),
			},
			OracleText: `
			Enchant creature
			As this Aura enters, choose a color.
			Enchanted creature has protection from the chosen color. This effect doesn't remove this Aura.
			{W}: Return this Aura to its owner's hand.
		`,
		},
	}
}
