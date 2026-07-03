package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RonomSerpent is the card definition for Ronom Serpent.
var RonomSerpent = newRonomSerpent()

func newRonomSerpent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ronom Serpent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Serpent},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                            game.RuleEffectCantAttack,
							AffectedSource:                  true,
							AttackDefenderControlsSelection: game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Snow}},
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
									Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Snow}},
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
			This creature can't attack unless defending player controls a snow land.
			When you control no snow lands, sacrifice this creature.
		`,
		},
	}
}
