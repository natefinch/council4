package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GempalmAvenger is the card definition for Gempalm Avenger.
//
// Type: Creature — Human Soldier
// Cost: {5}{W}
//
// Oracle text:
//
//	Cycling {2}{W} ({2}{W}, Discard this card: Draw a card.)
//	When you cycle this card, Soldier creatures get +1/+1 and gain first strike until end of turn.
var GempalmAvenger = newGempalmAvenger

func newGempalmAvenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Gempalm Avenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2), cost.W}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventCycled,
							Source: game.TriggerSourceSelf,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Soldier")}}),
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Soldier")}}),
											AddKeywords: []game.Keyword{
												game.FirstStrike,
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
			Cycling {2}{W} ({2}{W}, Discard this card: Draw a card.)
			When you cycle this card, Soldier creatures get +1/+1 and gain first strike until end of turn.
		`,
		},
	}
}
