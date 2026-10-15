package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EldraziRavager is the card definition for Eldrazi Ravager.
//
// Type: Creature — Eldrazi
// Cost: {5}{C}
//
// Oracle text:
//
//	Annihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent of their choice.)
//	Sacrifice two Eldrazi: Return this card from your graveyard to your hand.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var EldraziRavager = newEldraziRavager

func newEldraziRavager() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Eldrazi Ravager",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.C,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice two Eldrazi: Return this card from your graveyard to your hand.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice two Eldrazi",
							Amount:      2,
							SubtypesAny: cost.SubtypeSet{0: types.Eldrazi},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, defending player sacrifices a permanent of their choice.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
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
											Player: game.DefendingPlayerReference(),
											Amount: game.Fixed(1),
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
			Annihilator 1 (Whenever this creature attacks, defending player sacrifices a permanent of their choice.)
			Sacrifice two Eldrazi: Return this card from your graveyard to your hand.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
