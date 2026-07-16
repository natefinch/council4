package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SparkHarvest is the card definition for Spark Harvest.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature or pay {3}{B}.
//	Destroy target creature or planeswalker.
var SparkHarvest = newSparkHarvest

func newSparkHarvest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Spark Harvest",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			AdditionalCostChoices: []cost.AdditionalChoice{
				cost.AdditionalChoice{
					Options: []cost.AdditionalChoiceOption{
						cost.AdditionalChoiceOption{
							Label: "Sacrifice a creature",
							Costs: []cost.Additional{
								{
									Kind:               cost.AdditionalSacrifice,
									Text:               "sacrifice a creature",
									Amount:             1,
									MatchPermanentType: true,
									PermanentType:      types.Creature,
								},
							},
						},
						cost.AdditionalChoiceOption{
							Label: "Pay {3}{B}",
							Mana:  cost.Mana{cost.O(3), cost.B},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature or pay {3}{B}.
			Destroy target creature or planeswalker.
		`,
		},
	}
}
