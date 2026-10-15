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

// MinasTirith is the card definition for Minas Tirith.
//
// Type: Legendary Land
//
// Oracle text:
//
//	Minas Tirith enters tapped unless you control a legendary creature.
//	{T}: Add {W}.
//	{1}{W}, {T}: Draw a card. Activate only if you attacked with two or more creatures this turn.
var MinasTirith = newMinasTirith

func newMinasTirith() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Minas Tirith",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{W}, {T}: Draw a card. Activate only if you attacked with two or more creatures this turn.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.W,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text:              "only if you attacked with two or more creatures this turn",
						SourceCounterKind: counter.PlusOnePlusOne,
						EventHistory: opt.Val(game.EventHistoryCondition{
							Pattern: game.TriggerPattern{
								Event:           game.EventAttackerDeclared,
								Controller:      game.TriggerControllerYou,
								StackObjectKind: game.StackSpell,
								CounterKind:     counter.PlusOnePlusOne,
							},
							Window:   game.EventHistoryCurrentTurn,
							MinCount: 2,
						}),
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("Minas Tirith enters tapped unless you control a legendary creature.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}},
					}),
				}),
			},
			OracleText: `
			Minas Tirith enters tapped unless you control a legendary creature.
			{T}: Add {W}.
			{1}{W}, {T}: Draw a card. Activate only if you attacked with two or more creatures this turn.
		`,
		},
	}
}
