package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ContainmentConstruct is the card definition for Containment Construct.
//
// Type: Artifact Creature — Construct
// Cost: {2}
//
// Oracle text:
//
//	Whenever you discard a card, you may exile that card from your graveyard. If you do, you may play that card this turn.
var ContainmentConstruct = newContainmentConstruct

func newContainmentConstruct() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Containment Construct",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you discard a card, you may exile that card from your graveyard. If you do, you may play that card this turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventCardDiscarded,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ExileForPlay{
											Card: game.CardReference{
												Kind: game.CardReferenceEvent,
											},
											FromZone: zone.Graveyard,
											Duration: game.DurationThisTurn,
										},
										Optional: true,
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
			Whenever you discard a card, you may exile that card from your graveyard. If you do, you may play that card this turn.
		`,
		},
	}
}
