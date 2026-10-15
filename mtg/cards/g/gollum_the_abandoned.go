package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GollumTheAbandoned is the card definition for Gollum the Abandoned.
//
// Type: Legendary Creature — Halfling Horror
// Cost: {1}{B}
//
// Oracle text:
//
//	Gollum can't block.
//	When Gollum enters, exile up to one target card from an opponent's graveyard. Each opponent loses 2 life.
//	{2}, Sacrifice an artifact or creature: Return this card from your graveyard to your hand. Activate only as a sorcery.
var GollumTheAbandoned = newGollumTheAbandoned

func newGollumTheAbandoned() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gollum the Abandoned",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Halfling, types.Horror},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, Sacrifice an artifact or creature: Return this card from your graveyard to your hand. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:               cost.AdditionalSacrifice,
							Text:               "Sacrifice an artifact or creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							PermanentTypeAlt:   types.Creature,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Timing:         game.SorceryOnly,
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
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Gollum enters, exile up to one target card from an opponent's graveyard. Each opponent loses 2 life.",
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
										MaxTargets: 1,
										Constraint: "up to one target card from an opponent's graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{Controller: game.ControllerOpponent}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(2),
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
			OracleText: `
			Gollum can't block.
			When Gollum enters, exile up to one target card from an opponent's graveyard. Each opponent loses 2 life.
			{2}, Sacrifice an artifact or creature: Return this card from your graveyard to your hand. Activate only as a sorcery.
		`,
		},
	}
}
