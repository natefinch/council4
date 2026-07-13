package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MogisSWarhound is the card definition for Mogis's Warhound.
//
// Type: Enchantment Creature — Dog
// Cost: {1}{R}
//
// Oracle text:
//
//	Bestow {2}{R} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	This creature attacks each combat if able.
//	Enchanted creature gets +2/+2 and attacks each combat if able.
var MogisSWarhound = newMogisSWarhound

func newMogisSWarhound() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mogis's Warhound",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Dog},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.BestowStaticAbility(cost.Mana{cost.O(2), cost.R}, &game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.MustAttackStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectMustAttack,
							AffectedAttached: true,
						},
					},
				},
			},
			OracleText: `
			Bestow {2}{R} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
			This creature attacks each combat if able.
			Enchanted creature gets +2/+2 and attacks each combat if able.
		`,
		},
	}
}
