package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CurseOfChains is the card definition for Curse of Chains.
//
// Type: Enchantment — Aura
// Cost: {1}{W/U}
//
// Oracle text:
//
//	Enchant creature
//	At the beginning of each upkeep, tap enchanted creature.
var CurseOfChains = newCurseOfChains()

func newCurseOfChains() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Curse of Chains",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.W, mana.U),
			}),
			Colors:   []color.Color{color.Blue, color.White},
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepUpkeep,
						},
					},
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
			},
			OracleText: `
			Enchant creature
			At the beginning of each upkeep, tap enchanted creature.
		`,
		},
	}
}
