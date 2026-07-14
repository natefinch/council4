package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MidnightClock is the card definition for Midnight Clock.
//
// Type: Artifact
// Cost: {2}{U}
//
// Oracle text:
//
//	{T}: Add {U}.
//	{2}{U}: Put an hour counter on this artifact.
//	At the beginning of each upkeep, put an hour counter on this artifact.
//	When the twelfth hour counter is put on this artifact, shuffle your hand and graveyard into your library, then draw seven cards. Exile this artifact.
var MidnightClock = newMidnightClock

func newMidnightClock() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Midnight Clock",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{2}{U}: Put an hour counter on this artifact.",
					ManaCost:       opt.Val(cost.Mana{cost.O(2), cost.U}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Hour,
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
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
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.Hour,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventCountersAdded,
							Source:           game.TriggerSourceSelf,
							MatchCounterKind: true,
							CounterKind:      counter.Hour,
							CounterThreshold: 12,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ShuffleGraveyardIntoLibrary{
									Player:      game.ControllerReference(),
									IncludeHand: true,
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(7),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Exile{
									Object: game.SourceCardPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add {U}.
			{2}{U}: Put an hour counter on this artifact.
			At the beginning of each upkeep, put an hour counter on this artifact.
			When the twelfth hour counter is put on this artifact, shuffle your hand and graveyard into your library, then draw seven cards. Exile this artifact.
		`,
		},
	}
}
