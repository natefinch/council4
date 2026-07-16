package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DisruptionProtocol is the card definition for Disruption Protocol.
//
// Type: Instant
// Cost: {U}{U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, tap an untapped artifact you control or pay {1}.
//	Counter target spell.
var DisruptionProtocol = newDisruptionProtocol

func newDisruptionProtocol() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Disruption Protocol",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AdditionalCostChoices: []cost.AdditionalChoice{
				cost.AdditionalChoice{
					Options: []cost.AdditionalChoiceOption{
						cost.AdditionalChoiceOption{
							Label: "Tap an untapped artifact you control",
							Costs: []cost.Additional{
								{
									Kind:               cost.AdditionalTapPermanents,
									Text:               "tap an untapped artifact you control",
									Amount:             1,
									MatchPermanentType: true,
									PermanentType:      types.Artifact,
								},
							},
						},
						cost.AdditionalChoiceOption{
							Label: "Pay {1}",
							Mana:  cost.Mana{cost.O(1)},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, tap an untapped artifact you control or pay {1}.
			Counter target spell.
		`,
		},
	}
}
