package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinMorningstar is the card definition for Goblin Morningstar.
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +1/+0 and has trample.
//	Equip {1}{R}
//	When this Equipment enters, roll a d20.
//	1—9 | Create a 1/1 red Goblin creature token.
//	10—20 | Create a 1/1 red Goblin creature token, then attach this Equipment to it.
var GoblinMorningstar = newGoblinMorningstar

func newGoblinMorningstar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Morningstar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:      game.LayerPowerToughnessModify,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta: 1,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Trample,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(1), cost.R}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RollDie{
									Sides: 20,
								},
								PublishResult: game.ResultKey("die-roll-result"),
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(goblinMorningstarToken),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:         "die-roll-result",
									AmountRange: opt.Val(game.IntRange{Min: 1, Max: 9}),
								}),
							},
							{
								Primitive: game.CreateToken{
									Amount:        game.Fixed(1),
									Source:        game.TokenDef(goblinMorningstarToken),
									PublishLinked: game.LinkedKey("created-token"),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:         "die-roll-result",
									AmountRange: opt.Val(game.IntRange{Min: 10, Max: 20}),
								}),
							},
							{
								Primitive: game.Attach{
									Attachment: game.SourcePermanentReference(),
									Target:     game.LinkedObjectReference("created-token"),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:         "die-roll-result",
									AmountRange: opt.Val(game.IntRange{Min: 10, Max: 20}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Equipped creature gets +1/+0 and has trample.
			Equip {1}{R}
			When this Equipment enters, roll a d20.
			1—9 | Create a 1/1 red Goblin creature token.
			10—20 | Create a 1/1 red Goblin creature token, then attach this Equipment to it.
		`,
		},
	}
}

var goblinMorningstarToken = newGoblinMorningstarToken()

func newGoblinMorningstarToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
