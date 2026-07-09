package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CosmicCrucible is the card definition for Cosmic Crucible.
//
// Type: Enchantment
// Cost: {4}{G}{U}
//
// Oracle text:
//
//	At the beginning of your first main phase, add four mana in any combination of colors.
//	Whenever you cast a noncreature spell, you may copy it. You may choose new targets for the copy. Do this only once each turn. (A copy of a permanent spell becomes a token.)
var CosmicCrucible = newCosmicCrucible

func newCosmicCrucible() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Cosmic Crucible",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.U,
			}),
			Colors: []color.Color{color.Green, color.Blue},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepPrecombatMain,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:            game.Fixed(4),
									CombinationColors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						},
					},
					Optional:           true,
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CopyStackObject{
									Object:              game.EventStackObjectReference(),
									MayChooseNewTargets: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your first main phase, add four mana in any combination of colors.
			Whenever you cast a noncreature spell, you may copy it. You may choose new targets for the copy. Do this only once each turn. (A copy of a permanent spell becomes a token.)
		`,
		},
	}
}
