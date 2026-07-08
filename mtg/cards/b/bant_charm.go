package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BantCharm is the card definition for Bant Charm.
var BantCharm = newBantCharm

func newBantCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Bant Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors: []color.Color{color.Green, color.Blue, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Destroy target artifact.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Put target creature on the bottom of its owner's library.",
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
								Primitive: game.PutPermanentOnLibrary{
									Object: game.TargetPermanentReference(0),
									Bottom: true,
								},
							},
						},
					},
					game.Mode{
						Text: "Counter target instant spell.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target instant spell",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									SpellCardTypes:   []types.Card{types.Instant},
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
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Destroy target artifact.
			• Put target creature on the bottom of its owner's library.
			• Counter target instant spell.
		`,
		},
	}
}
