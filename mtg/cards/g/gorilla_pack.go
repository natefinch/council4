package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GorillaPack is the card definition for Gorilla Pack.
var GorillaPack = newGorillaPack

func newGorillaPack() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gorilla Pack",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ape},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                            game.RuleEffectCantAttack,
							AffectedSource:                  true,
							AttackDefenderControlsSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Forest")}},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerState,
						State: opt.Val(game.StateTriggerCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Forest")}},
									MinCount:  1,
								}),
							}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.SourceCardPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This creature can't attack unless defending player controls a Forest.
			When you control no Forests, sacrifice this creature.
		`,
		},
	}
}
