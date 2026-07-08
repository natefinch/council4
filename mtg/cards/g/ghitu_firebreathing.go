package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GhituFirebreathing is the card definition for Ghitu Firebreathing.
//
// Type: Enchantment — Aura
// Cost: {1}{R}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Enchant creature
//	{R}: Enchanted creature gets +1/+0 until end of turn.
//	{R}: Return this Aura to its owner's hand.
var GhituFirebreathing = newGhituFirebreathing

func newGhituFirebreathing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ghitu Firebreathing",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
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
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{R}: Enchanted creature gets +1/+0 until end of turn.",
					ManaCost:       opt.Val(cost.Mana{cost.R}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
											PowerDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:           "{R}: Return this Aura to its owner's hand.",
					ManaCost:       opt.Val(cost.Mana{cost.R}),
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
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Enchant creature
			{R}: Enchanted creature gets +1/+0 until end of turn.
			{R}: Return this Aura to its owner's hand.
		`,
		},
	}
}
