package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OneThousandLashes is the card definition for One Thousand Lashes.
//
// Type: Enchantment — Aura
// Cost: {2}{W}{B}
//
// Oracle text:
//
//	Enchant creature
//	Enchanted creature can't attack or block, and its activated abilities can't be activated.
//	At the beginning of the upkeep of enchanted creature's controller, that player loses 1 life.
var OneThousandLashes = newOneThousandLashes()

func newOneThousandLashes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "One Thousand Lashes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.B,
			}),
			Colors:   []color.Color{color.Black, color.White},
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
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantAttack,
							AffectedAttached: true,
						},
						game.RuleEffect{
							Kind:             game.RuleEffectCantBlock,
							AffectedAttached: true,
						},
						game.RuleEffect{
							Kind:             game.RuleEffectCantActivateAbilitiesOfPermanent,
							AffectedAttached: true,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:                             game.EventBeginningOfStep,
							Step:                              game.StepUpkeep,
							StepPlayerSourceAttachedSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(1),
									Player: game.EventPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			Enchanted creature can't attack or block, and its activated abilities can't be activated.
			At the beginning of the upkeep of enchanted creature's controller, that player loses 1 life.
		`,
		},
	}
}
