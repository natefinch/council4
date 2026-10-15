package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WhiteScarab is the card definition for White Scarab.
//
// Type: Enchantment — Aura
// Cost: {W}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature can't be blocked by white creatures.
//	Enchanted creature gets +2/+2 as long as an opponent controls a white permanent.
var WhiteScarab = newWhiteScarab

func newWhiteScarab() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "White Scarab",
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
					Text: "Enchanted creature can't be blocked by white creatures.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedAttached: true,
							Duration:         game.DurationPermanent,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionColor,
								Color: color.White,
							},
						},
					},
				},
				game.StaticAbility{
					Text: "Enchanted creature gets +2/+2 as long as an opponent controls a white permanent.",
					Condition: opt.Val(game.Condition{
						Text: "as long as an opponent controls a white permanent",
						AnyOpponentControls: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.White}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature can't be blocked by white creatures.
			Enchanted creature gets +2/+2 as long as an opponent controls a white permanent.
		`,
		},
	}
}
