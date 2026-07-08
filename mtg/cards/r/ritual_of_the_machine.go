package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RitualOfTheMachine is the card definition for Ritual of the Machine.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature.
//	Gain control of target nonartifact, nonblack creature.
var RitualOfTheMachine = newRitualOfTheMachine

func newRitualOfTheMachine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ritual of the Machine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:               cost.AdditionalSacrifice,
					Text:               "sacrifice a creature",
					Amount:             1,
					MatchPermanentType: true,
					PermanentType:      types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonartifact, nonblack creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedTypes: []types.Card{types.Artifact}, ExcludedColors: []color.Color{color.Black}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:         game.LayerControl,
									NewController: opt.Val(game.Player1),
								},
							},
							Duration: game.DurationPermanent,
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature.
			Gain control of target nonartifact, nonblack creature.
		`,
		},
	}
}
