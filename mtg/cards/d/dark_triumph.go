package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarkTriumph is the card definition for Dark Triumph.
//
// Type: Instant
// Cost: {4}{B}
//
// Oracle text:
//
//	If you control a Swamp, you may sacrifice a creature rather than pay this spell's mana cost.
//	Creatures you control get +2/+0 until end of turn.
var DarkTriumph = newDarkTriumph

func newDarkTriumph() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Dark Triumph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice a creature",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "sacrifice a creature",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
						},
					},
					Condition:        cost.AlternativeConditionControlsPermanentSubtype,
					ConditionSubtype: types.Swamp,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:      game.LayerPowerToughnessModify,
									Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									PowerDelta: 2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			If you control a Swamp, you may sacrifice a creature rather than pay this spell's mana cost.
			Creatures you control get +2/+0 until end of turn.
		`,
		},
	}
}
