package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PestilentCauldron is the card definition for Pestilent Cauldron // Restorative Burst.
//
// Type: Artifact // Sorcery
// Face: Restorative Burst — Sorcery ({3}{G}{G})
//
// Oracle text:
//
//	{T}, Discard a card: Create a 1/1 black and green Pest creature token with "When this token dies, you gain 1 life."
//	{1}, {T}: Each opponent mills cards equal to the amount of life you gained this turn.
//	{4}, {T}: Exile four target cards from a single graveyard. Draw a card.
var PestilentCauldron = newPestilentCauldron()

func newPestilentCauldron() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Pestilent Cauldron",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Discard a card: Create a 1/1 black and green Pest creature token with \"When this token dies, you gain 1 life.\"",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard a card",
							Amount: 1,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(pestilentCauldronToken),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{1}, {T}: Each opponent mills cards equal to the amount of life you gained this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountLifeGainedThisTurn,
										Multiplier: 1,
									}),
									PlayerGroup: game.OpponentsReference(),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{4}, {T}: Exile four target cards from a single graveyard. Draw a card.",
					ManaCost:        opt.Val(cost.Mana{cost.O(4)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 4,
								MaxTargets: 4,
								Constraint: "four target cards from a single graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 3},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}, Discard a card: Create a 1/1 black and green Pest creature token with "When this token dies, you gain 1 life."
			{1}, {T}: Each opponent mills cards equal to the amount of life you gained this turn.
			{4}, {T}: Exile four target cards from a single graveyard. Draw a card.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name: "Restorative Burst",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 2,
						Constraint: "up to two target creature, land, and/or planeswalker cards from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land, types.Planeswalker}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MoveCard{
							Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.GainLife{
							Amount:      game.Fixed(4),
							PlayerGroup: game.AllPlayersReference(),
						},
					},
					{
						Primitive: game.Exile{
							SourceSpell: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return up to two target creature, land, and/or planeswalker cards from your graveyard to your hand. Each player gains 4 life. Exile Restorative Burst.
		`,
		}),
	}
}

var pestilentCauldronToken = newPestilentCauldronToken()

func newPestilentCauldronToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Pest",
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pest},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
