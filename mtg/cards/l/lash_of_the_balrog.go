package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LashOfTheBalrog is the card definition for Lash of the Balrog.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature or pay {4}.
//	Destroy target creature.
var LashOfTheBalrog = newLashOfTheBalrog

func newLashOfTheBalrog() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Lash of the Balrog",
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
							Label: "Pay {4}",
							Mana:  cost.Mana{cost.O(4)},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
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
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature or pay {4}.
			Destroy target creature.
		`,
		},
	}
}
