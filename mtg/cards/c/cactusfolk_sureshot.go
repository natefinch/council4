package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CactusfolkSureshot is the card definition for Cactusfolk Sureshot.
//
// Type: Creature — Plant Mercenary
// Cost: {2}{R}{G}
//
// Oracle text:
//
//	Reach
//	Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
//	At the beginning of combat on your turn, other creatures you control with power 4 or greater gain trample and haste until end of turn.
var CactusfolkSureshot = newCactusfolkSureshot

func newCactusfolkSureshot() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Cactusfolk Sureshot",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Mercenary},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})}, game.SourcePermanentReference()),
											AddKeywords: []game.Keyword{
												game.Trample,
												game.Haste,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Reach
			Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
			At the beginning of combat on your turn, other creatures you control with power 4 or greater gain trample and haste until end of turn.
		`,
		},
	}
}
