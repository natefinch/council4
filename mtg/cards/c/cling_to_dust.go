package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ClingToDust is the card definition for Cling to Dust.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Exile target card from a graveyard. If it was a creature card, you gain 3 life. Otherwise, you draw a card.
//	Escape—{3}{B}, Exile five other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
var ClingToDust = newClingToDust

func newClingToDust() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Cling to Dust",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SimpleKeyword{
							Kind: game.Escape,
						},
					},
				},
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Escape",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:          cost.AdditionalExile,
							Text:          "Exile five other cards from your graveyard",
							Amount:        5,
							Source:        zone.Graveyard,
							ExcludeSource: true,
						},
					},
					Mechanic: cost.AlternativeMechanicEscape,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target card from a graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind: game.CardReferenceTarget,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Exile,
						},
						PublishResult: game.ResultKey("condition-target-card-1"),
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:              opt.Val(game.TargetCardReference(0)),
								ObjectMatches:       opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
								TargetCardResultKey: game.ResultKey("condition-target-card-1"),
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Exile target card from a graveyard. If it was a creature card, you gain 3 life. Otherwise, you draw a card.
			Escape—{3}{B}, Exile five other cards from your graveyard. (You may cast this card from your graveyard for its escape cost.)
		`,
		},
	}
}
