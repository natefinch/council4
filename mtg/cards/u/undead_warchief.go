package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UndeadWarchief is the card definition for Undead Warchief.
//
// Type: Creature — Zombie
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Zombie spells you cast cost {1} less to cast.
//	Zombie creatures you control get +2/+1.
var UndeadWarchief = newUndeadWarchief

func newUndeadWarchief() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Undead Warchief",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Zombie spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Zombie}},
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Zombie creatures you control get +2/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Zombie")}}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Zombie spells you cast cost {1} less to cast.
			Zombie creatures you control get +2/+1.
		`,
		},
	}
}
