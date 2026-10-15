package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DuskMangler is the card definition for Dusk Mangler.
//
// Type: Creature — Horror
// Cost: {5}{B}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature, discard a card, or pay 4 life.
//	When this creature enters, each opponent sacrifices a creature of their choice, discards a card, and loses 4 life.
var DuskMangler = newDuskMangler

func newDuskMangler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dusk Mangler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, each opponent sacrifices a creature of their choice, discards a card, and loses 4 life.",
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
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											PlayerGroup: game.OpponentsReference(),
											Amount:      game.Fixed(1),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
										},
									},
									game.Instruction{
										Primitive: game.Discard{
											Amount:      game.Fixed(1),
											PlayerGroup: game.OpponentsReference(),
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(4),
											PlayerGroup: game.OpponentsReference(),
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
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice a creature",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
					ChoiceGroup:        1,
				},
				{
					Kind:        cost.AdditionalDiscard,
					Text:        "discard a card",
					Amount:      1,
					Source:      zone.Hand,
					ChoiceGroup: 1,
				},
				{
					Kind:        cost.AdditionalPayLife,
					Text:        "pay 4 life",
					Amount:      4,
					ChoiceGroup: 1,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature, discard a card, or pay 4 life.
			When this creature enters, each opponent sacrifices a creature of their choice, discards a card, and loses 4 life.
		`,
		},
	}
}
