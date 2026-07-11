package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RamosianRally is the card definition for Ramosian Rally.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	If you control a Plains, you may tap an untapped creature you control rather than pay this spell's mana cost.
//	Creatures you control get +1/+1 until end of turn.
var RamosianRally = newRamosianRally

func newRamosianRally() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ramosian Rally",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:          game.LayerPowerToughnessModify,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									PowerDelta:     1,
									ToughnessDelta: 1,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			If you control a Plains, you may tap an untapped creature you control rather than pay this spell's mana cost.
			Creatures you control get +1/+1 until end of turn.
		`,
		},
	}
}
