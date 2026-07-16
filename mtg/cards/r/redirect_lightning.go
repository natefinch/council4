package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RedirectLightning is the card definition for Redirect Lightning.
//
// Type: Instant — Lesson
// Cost: {R}
//
// Oracle text:
//
//	As an additional cost to cast this spell, pay 5 life or pay {2}.
//	Change the target of target spell or ability with a single target.
var RedirectLightning = newRedirectLightning

func newRedirectLightning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Redirect Lightning",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Lesson},
			AdditionalCostChoices: []cost.AdditionalChoice{
				cost.AdditionalChoice{
					Options: []cost.AdditionalChoiceOption{
						cost.AdditionalChoiceOption{
							Label: "Pay 5 life",
							Costs: []cost.Additional{
								{
									Kind:   cost.AdditionalPayLife,
									Text:   "pay 5 life",
									Amount: 5,
								},
							},
						},
						cost.AdditionalChoiceOption{
							Label: "Pay {2}",
							Mana:  cost.Mana{cost.O(2)},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell or ability",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseNewTargets{
							Object: game.TargetStackObjectReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, pay 5 life or pay {2}.
			Change the target of target spell or ability with a single target.
		`,
		},
	}
}
