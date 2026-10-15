package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlightBreathCatoblepas is the card definition for Blight-Breath Catoblepas.
//
// Type: Creature — Beast
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	When this creature enters, target creature an opponent controls gets -X/-X until end of turn, where X is your devotion to black. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)
var BlightBreathCatoblepas = newBlightBreathCatoblepas

func newBlightBreathCatoblepas() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blight-Breath Catoblepas",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, target creature an opponent controls gets -X/-X until end of turn, where X is your devotion to black. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ModifyPT{
											Object: game.TargetPermanentReference(0),
											PowerDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountDevotion,
												Multiplier: -1,
												Colors:     []color.Color{color.Black},
											}),
											ToughnessDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountDevotion,
												Multiplier: -1,
												Colors:     []color.Color{color.Black},
											}),
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			When this creature enters, target creature an opponent controls gets -X/-X until end of turn, where X is your devotion to black. (Each {B} in the mana costs of permanents you control counts toward your devotion to black.)
		`,
		},
	}
}
