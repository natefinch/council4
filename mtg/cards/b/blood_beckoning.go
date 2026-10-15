package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BloodBeckoning is the card definition for Blood Beckoning.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	Return target creature card from your graveyard to your hand. If this spell was kicked, instead return two target creature cards from your graveyard to your hand.
var BloodBeckoning = newBloodBeckoning

func newBloodBeckoning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Blood Beckoning",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker {3}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(3),
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature card from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
						Gate:       game.TargetGateSpellNotKicked,
					},
					game.TargetSpec{
						MinTargets: 2,
						MaxTargets: 2,
						Constraint: "two target creature cards from your graveyard",
						Allow:      game.TargetAllowCard,
						TargetZone: zone.Graveyard,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
						Gate:       game.TargetGateSpellKicked,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind: game.CardReferenceTarget,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:         true,
								SpellWasKicked: true,
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind:        game.CardReferenceTarget,
								TargetIndex: 1,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
					{
						Primitive: game.MoveCard{
							Card: game.CardReference{
								Kind:        game.CardReferenceTarget,
								TargetIndex: 2,
							},
							FromZone:    zone.Graveyard,
							Destination: zone.Hand,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {3} (You may pay an additional {3} as you cast this spell.)
			Return target creature card from your graveyard to your hand. If this spell was kicked, instead return two target creature cards from your graveyard to your hand.
		`,
		},
	}
}
