package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Endbringer is the card definition for Endbringer.
//
// Type: Creature — Eldrazi
// Cost: {5}{C}
//
// Oracle text:
//
//	Untap this creature during each other player's untap step.
//	{T}: This creature deals 1 damage to any target.
//	{C}, {T}: Target creature can't attack or block this turn.
//	{C}{C}, {T}: Draw a card.
var Endbringer = newEndbringer

func newEndbringer() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Endbringer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.C,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectUntapDuringOtherPlayersUntapStep,
							AffectedSource: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: This creature deals 1 damage to any target.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{C}, {T}: Target creature can't attack or block this turn.",
					ManaCost:        opt.Val(cost.Mana{cost.C}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantAttack,
										},
										game.RuleEffect{
											Kind: game.RuleEffectCantBlock,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "{C}{C}, {T}: Draw a card.",
					ManaCost:        opt.Val(cost.Mana{cost.C, cost.C}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
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
			Untap this creature during each other player's untap step.
			{T}: This creature deals 1 damage to any target.
			{C}, {T}: Target creature can't attack or block this turn.
			{C}{C}, {T}: Draw a card.
		`,
		},
	}
}
