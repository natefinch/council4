package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TourachDreadCantor is the card definition for Tourach, Dread Cantor.
//
// Type: Legendary Creature — Human Cleric
// Cost: {1}{B}
//
// Oracle text:
//
//	Kicker {B}{B} (You may pay an additional {B}{B} as you cast this spell.)
//	Protection from white
//	Whenever an opponent discards a card, put a +1/+1 counter on Tourach.
//	When Tourach enters, if it was kicked, target opponent discards two cards at random.
var TourachDreadCantor = newTourachDreadCantor()

func newTourachDreadCantor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tourach, Dread Cantor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Cleric},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.B, cost.B}},
					},
				},
				game.ProtectionFromColorsStaticAbility(color.White),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventCardDiscarded,
							Player: game.TriggerPlayerOpponent,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf:                        "if it was kicked",
						InterveningIfEventPermanentWasKicked: true,
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount:   game.Fixed(2),
									Player:   game.TargetPlayerReference(0),
									AtRandom: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Kicker {B}{B} (You may pay an additional {B}{B} as you cast this spell.)
			Protection from white
			Whenever an opponent discards a card, put a +1/+1 counter on Tourach.
			When Tourach enters, if it was kicked, target opponent discards two cards at random.
		`,
		},
	}
}
