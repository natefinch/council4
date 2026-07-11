package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrimSCure is the card definition for Orim's Cure.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	If you control a Plains, you may tap an untapped creature you control rather than pay this spell's mana cost.
//	Prevent the next 4 damage that would be dealt to any target this turn.
var OrimSCure = newOrimSCure

func newOrimSCure() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Orim's Cure",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Tap an untapped creature you control",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "tap an untapped creature you control",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
						},
					},
					Condition:        cost.AlternativeConditionControlsPermanentSubtype,
					ConditionSubtype: types.Plains,
				},
			},
			SpellAbility: opt.Val(game.Mode{
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
						Primitive: game.PreventDamage{
							AnyTarget: game.AnyTargetDamageRecipient(0),
							Amount:    game.Fixed(4),
						},
					},
				},
			}.Ability()),
			OracleText: `
			If you control a Plains, you may tap an untapped creature you control rather than pay this spell's mana cost.
			Prevent the next 4 damage that would be dealt to any target this turn.
		`,
		},
	}
}
